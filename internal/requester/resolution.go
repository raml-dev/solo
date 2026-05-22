// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package requester

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"solo/internal/collection"
	"solo/internal/tools"
	"strings"
)

type resolutionContext struct {
	sessionVars    map[string]string
	envVars        map[string]string
	collectionVars map[string]string
}

var placeholderPattern = regexp.MustCompile(tools.PLACEHOLDER_REGEXP)
var placeholderBraceDecoder = strings.NewReplacer("%7B", "{", "%7D", "}", "%7b", "{", "%7d", "}")

// resolveTemplateString replaces {{placeholders}} using session vars first, then environment vars.
func resolveTemplateString(input string, ctx resolutionContext) string {
	if input == "" {
		return input
	}

	return placeholderPattern.ReplaceAllStringFunc(input, func(match string) string {
		groups := placeholderPattern.FindStringSubmatch(match)
		if len(groups) < 2 {
			return match
		}

		key := strings.TrimSpace(groups[1])
		if val, ok := resolveScopedValue(key, ctx); ok {
			return val
		}
		return match
	})
}

func resolveScopedValue(key string, ctx resolutionContext) (string, bool) {
	if val, ok := ctx.sessionVars[key]; ok {
		return val, true
	}

	if val, ok := ctx.envVars[key]; ok {
		if strings.TrimSpace(val) != "" {
			return val, true
		}
		if collectionVal, ok := ctx.collectionVars[key]; ok {
			return collectionVal, true
		}
		return val, true
	}

	if val, ok := ctx.collectionVars[key]; ok {
		return val, true
	}

	return "", false
}

// resolveHeaders resolves placeholders in both header names and values.
func resolveHeaders(headers map[string]any, ctx resolutionContext) map[string]any {
	if len(headers) == 0 {
		return map[string]any{}
	}

	resolved := make(map[string]any, len(headers))
	for key, value := range headers {
		resolvedKey := resolveTemplateString(key, ctx)
		switch typed := value.(type) {
		case string:
			resolved[resolvedKey] = resolveTemplateString(typed, ctx)
		default:
			resolved[resolvedKey] = value
		}
	}

	return resolved
}

// resolveCookies resolves placeholders in both cookie names and values.
func resolveCookies(cookies map[string]any, ctx resolutionContext) map[string]any {
	if len(cookies) == 0 {
		return map[string]any{}
	}

	resolved := make(map[string]any, len(cookies))
	for key, value := range cookies {
		resolvedKey := resolveTemplateString(key, ctx)
		switch typed := value.(type) {
		case string:
			resolved[resolvedKey] = resolveTemplateString(typed, ctx)
		default:
			resolved[resolvedKey] = value
		}
	}

	return resolved
}

// resolveParams resolves placeholders in both param keys and values.
// Unlike resolveHeaders/resolveCookies, params are typed as map[string]string
// because query parameter keys and values are always plain strings.
func resolveParams(params map[string]string, ctx resolutionContext) map[string]string {
	if len(params) == 0 {
		return map[string]string{}
	}

	resolved := make(map[string]string, len(params))
	for key, value := range params {
		resolvedKey := resolveTemplateString(key, ctx)
		resolved[resolvedKey] = resolveTemplateString(value, ctx)
	}

	return resolved
}

// resolveAuthConfig resolves placeholders in the OAuth configuration fields.
func resolveAuthConfig(authCfg *collection.AuthConfiguration, ctx resolutionContext) *collection.AuthConfiguration {
	if authCfg == nil {
		return nil
	}

	resolved := &collection.AuthConfiguration{
		Enabled:   authCfg.Enabled,
		TokenURL:  resolveTemplateString(authCfg.TokenURL, ctx),
		TokenPath: authCfg.TokenPath,
		Template:  make(map[string]string, len(authCfg.Template)),
	}

	for key, value := range authCfg.Template {
		resolved.Template[key] = resolveTemplateString(value, ctx)
	}

	return resolved
}

// resolve returns a copy of execution options with all supported fields resolved.
func (opts ExecutionOptions) resolve(ctx resolutionContext) ExecutionOptions {
	resolved := opts
	resolved.URL = resolveTemplateString(opts.URL, ctx)
	resolved.Body = resolveTemplateString(opts.Body, ctx)
	resolved.Headers = resolveHeaders(opts.Headers, ctx)
	resolved.Cookies = resolveCookies(opts.Cookies, ctx)
	resolved.Params = resolveParams(opts.Params, ctx)
	resolved.Auth = resolveAuthConfig(opts.Auth, ctx)
	return resolved
}

// buildRequestFromOptions creates an http.Request from already-resolved options.
func buildRequestFromOptions(ctx context.Context, opts ExecutionOptions) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, opts.Method, opts.URL, bytes.NewReader([]byte(opts.Body)))
	if err != nil {
		return nil, err
	}

	for key, value := range opts.Headers {
		vStr, ok := value.(string)
		if ok {
			request.Header.Set(key, vStr)
		}
	}

	for key, value := range opts.Cookies {
		vStr, ok := value.(string)
		if ok {
			request.AddCookie(&http.Cookie{Name: key, Value: vStr})
		}
	}

	// Append query params to the URL. We preserve any query string already
	// present in the URL string and add params on top via url.Values.Set.
	// Params with an empty value are included as "?key=" (standard net/url behaviour).
	if len(opts.Params) > 0 {
		q := request.URL.Query()
		for key, value := range opts.Params {
			q.Set(key, value)
		}
		request.URL.RawQuery = q.Encode()
	}

	return request, nil
}

// resolveRequestInPlace applies a second resolution pass on a request mutated by scripts.
func resolveRequestInPlace(req *http.Request, ctx resolutionContext) {
	if req == nil {
		return
	}

	if req.URL != nil {
		rawURL := placeholderBraceDecoder.Replace(req.URL.String())
		if resolvedURL, err := url.Parse(resolveTemplateString(rawURL, ctx)); err == nil {
			req.URL = resolvedURL
		}
	}

	body := readRequestBody(req)
	resolvedBody := resolveTemplateString(body, ctx)
	req.Body = io.NopCloser(bytes.NewBufferString(resolvedBody))
	req.ContentLength = int64(len(resolvedBody))

	if len(req.Header) > 0 {
		resolvedHeader := make(http.Header, len(req.Header))
		for key, values := range req.Header {
			resolvedKey := resolveTemplateString(key, ctx)
			for _, value := range values {
				resolvedHeader.Add(resolvedKey, resolveTemplateString(value, ctx))
			}
		}
		req.Header = resolvedHeader
	}
}

// readRequestBody reads and restores the request body so it can be safely reused.
func readRequestBody(req *http.Request) string {
	if req == nil || req.Body == nil {
		return ""
	}

	bodyBytes, _ := io.ReadAll(req.Body)
	req.Body.Close()
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	return string(bodyBytes)
}
