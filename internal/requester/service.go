// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package requester

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"solo/internal/auth"
	"solo/internal/collection"
	"solo/internal/configuration"
	"solo/internal/environment"
	"solo/internal/host"
	"solo/internal/script"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ExecutionOptions encapsulates all parameters required to execute a request
type ExecutionOptions struct {
	Method              string                                 `json:"method"`
	URL                 string                                 `json:"url"`
	Body                string                                 `json:"body"`
	Headers             map[string]any                         `json:"headers"`
	Cookies             map[string]any                         `json:"cookies"`
	CollectionVariables map[string]string                      `json:"collectionVariables,omitempty"`
	Settings            *configuration.RequestSettingsOverride `json:"settings,omitempty"`
	Auth                *collection.AuthConfiguration          `json:"auth,omitempty"`
	PreRequestScript    string                                 `json:"preRequestScript,omitempty"`
	PostResponseScript  string                                 `json:"postResponseScript,omitempty"`
}

type Service struct {
	hostManager        *host.HostManager
	configManager      *configuration.ConfigurationManager
	environmentManager *environment.EnvironmentManager
	scriptManager      *script.ScriptManager
	authManager        *auth.AuthManager
	ctx                context.Context
}

type executeResult struct {
	response       *http.Response
	request        *http.Request
	requestBody    string
	envVars        map[string]string
	collectionVars map[string]string
}

func NewService(cm *configuration.ConfigurationManager, em *environment.EnvironmentManager, sm *script.ScriptManager, hm *host.HostManager, am *auth.AuthManager) *Service {
	return &Service{
		hostManager:        hm,
		configManager:      cm,
		environmentManager: em,
		scriptManager:      sm,
		authManager:        am,
	}
}

// SetContext stores the Wails context used to emit runtime events.
func (s *Service) SetContext(ctx context.Context) {
	s.ctx = ctx
}

// emitEvent sends an event to the frontend and safely ignores invalid contexts.
func (s *Service) emitEvent(eventName string, data ...interface{}) {
	if s.ctx == nil || s.ctx == context.TODO() || s.ctx == context.Background() {
		return
	}

	defer func() {
		if r := recover(); r != nil {
			slog.Debug("Skipping event emit: invalid Wails context", "event", eventName, "recover", r)
		}
	}()

	runtime.EventsEmit(s.ctx, eventName, data...)
}

// loadSelectedEnvironmentValues loads the currently selected environment as key/value strings.
func (s *Service) loadSelectedEnvironmentValues() map[string]string {
	if s.configManager == nil || s.environmentManager == nil {
		return map[string]string{}
	}

	selectedEnvName := strings.TrimSpace(s.configManager.GetSelectedEnvironment())
	if selectedEnvName == "" {
		return map[string]string{}
	}

	selectedEnv, err := s.environmentManager.LoadEnvironment(selectedEnvName)
	if err != nil {
		slog.Warn("Failed to load selected environment for request execution",
			"environment", selectedEnvName,
			"error", err)
		return map[string]string{}
	}

	values := make(map[string]string, len(selectedEnv.Values))
	for key, value := range selectedEnv.Values {
		values[key] = value.Value
	}

	return values
}

// applyRequestSettings applies global request settings and per-request overrides to the client.
func (s *Service) applyRequestSettings(client *http.Client, overrides *configuration.RequestSettingsOverride) {
	if s.configManager == nil {
		return
	}

	cfg := s.configManager.Get()
	hasOverride := overrides != nil

	timeout := cfg.Request.TimeoutSeconds
	if hasOverride && overrides.TimeoutSeconds != nil {
		timeout = *overrides.TimeoutSeconds
	}
	if timeout > 0 {
		client.Timeout = time.Duration(timeout) * time.Second
	}

	followRedirects := cfg.Request.FollowRedirects
	if hasOverride && overrides.FollowRedirects != nil {
		followRedirects = *overrides.FollowRedirects
	}

	maxRedirects := cfg.Request.MaxRedirects
	if hasOverride && overrides.MaxRedirects != nil {
		maxRedirects = *overrides.MaxRedirects
	}

	if !followRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else if maxRedirects > 0 {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return fmt.Errorf("stopped after %d redirects", maxRedirects)
			}
			return nil
		}
	}

	proxyURL := cfg.Request.ProxyURL
	if hasOverride && overrides.ProxyURL != nil {
		proxyURL = *overrides.ProxyURL
	}

	if proxyURL != "" {
		parsedProxyURL, err := url.Parse(proxyURL)
		if err == nil {
			if transport, ok := client.Transport.(*http.Transport); ok {
				newTransport := transport.Clone()
				newTransport.Proxy = http.ProxyURL(parsedProxyURL)
				client.Transport = newTransport
			}
		}
	}

	validateSSL := cfg.Request.ValidateSSL
	if hasOverride && overrides.ValidateSSL != nil {
		validateSSL = *overrides.ValidateSSL
	}

	if !validateSSL {
		if transport, ok := client.Transport.(*http.Transport); ok {
			newTransport := transport.Clone()
			if newTransport.TLSClientConfig == nil {
				newTransport.TLSClientConfig = &tls.Config{}
			}
			newTransport.TLSClientConfig.InsecureSkipVerify = true
			client.Transport = newTransport
		}
	}

	slog.Debug("Request settings applied",
		"timeout", timeout,
		"follow_redirects", followRedirects,
		"max_redirects", maxRedirects,
		"proxy", proxyURL,
		"validate_ssl", validateSSL)
}

// applyDefaultUserAgent sets the User-Agent header using config defaults and local overrides.
func (s *Service) applyDefaultUserAgent(request *http.Request, overrides *configuration.RequestSettingsOverride) {
	if s.configManager == nil || request == nil {
		return
	}

	cfg := s.configManager.Get()
	userAgent := cfg.Request.DefaultUserAgent
	if overrides != nil && overrides.DefaultUserAgent != nil {
		override := strings.TrimSpace(*overrides.DefaultUserAgent)
		if override != "" {
			userAgent = override
		}
	}
	if userAgent != "" {
		request.Header.Set("User-Agent", userAgent)
	}
}

// applyHostCookies adds host-level cookies configured for the request host.
func (s *Service) applyHostCookies(request *http.Request) {
	if request == nil || request.URL == nil {
		return
	}

	if hostCfg, ok := s.hostManager.GetHost(request.URL.Host); ok {
		for name, value := range hostCfg.Cookies {
			request.AddCookie(&http.Cookie{Name: name, Value: value})
		}
	}
}

// injectAuthorization fetches and injects an OAuth2 bearer token when auth is enabled.
// If Authorization is already set, it is left untouched.
func (s *Service) injectAuthorization(request *http.Request, authCfg *collection.AuthConfiguration) error {
	if request == nil || authCfg == nil || !authCfg.Enabled || s.authManager == nil {
		return nil
	}

	if strings.TrimSpace(request.Header.Get("Authorization")) != "" {
		slog.Debug("Skipping OAuth2 Authorization injection because request already defines Authorization")
		return nil
	}

	token, refreshed, err := s.authManager.GetOrRefreshToken(*authCfg)
	if err != nil {
		slog.Error("Failed to get/refresh OAuth2 token", "error", err)
		return fmt.Errorf("authentication failed: %w", err)
	}

	if token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
		slog.Debug("Injected OAuth2 Bearer token", "refreshed", refreshed)
	}

	return nil
}

// Execute prepares and sends the HTTP request using backend-owned resolution logic.
// It resolves placeholders, runs pre-request scripts, and returns execution metadata.
func (s *Service) Execute(opts ExecutionOptions) (*executeResult, error) {
	slog.Debug("Preparing HTTP request",
		"method", opts.Method,
		"url", opts.URL,
		"headers_count", len(opts.Headers),
		"body_length", len(opts.Body))

	envVars := s.loadSelectedEnvironmentValues()
	collectionVars := opts.CollectionVariables
	if collectionVars == nil {
		collectionVars = map[string]string{}
	}
	sessionVars := map[string]string{}
	if s.scriptManager != nil {
		sessionVars = s.scriptManager.GetSessionVars()
	}

	resolutionCtx := resolutionContext{
		sessionVars:    sessionVars,
		envVars:        envVars,
		collectionVars: collectionVars,
	}
	resolvedOpts := opts.resolve(resolutionCtx)

	request, err := buildRequestFromOptions(resolvedOpts)
	if err != nil {
		slog.Error("Failed to create HTTP request", "method", resolvedOpts.Method, "url", resolvedOpts.URL, "error", err)
		return nil, err
	}

	s.applyDefaultUserAgent(request, resolvedOpts.Settings)

	// Execute pre-request script (may mutate method, url, headers, body)
	if s.scriptManager != nil && opts.PreRequestScript != "" {
		sessionVars, err = s.scriptManager.ExecutePreRequestWithScope(opts.PreRequestScript, request, envVars, collectionVars)
		if err != nil {
			slog.Warn("Pre-request script error", "error", err)
			return nil, fmt.Errorf("pre-request script error: %w", err)
		}
		resolveRequestInPlace(request, resolutionContext{
			sessionVars:    sessionVars,
			envVars:        envVars,
			collectionVars: collectionVars,
		})
		slog.Debug("Pre-request script executed successfully")
	}

	client, err := s.hostManager.GetClientForUrl(request.URL.String())
	if err != nil {
		slog.Error("Failed to get HTTP client", "url", request.URL.String(), "error", err)
		return nil, err
	}
	s.applyRequestSettings(client, resolvedOpts.Settings)

	s.applyHostCookies(request)

	finalAuthConfig := resolveAuthConfig(opts.Auth, resolutionContext{
		sessionVars:    sessionVars,
		envVars:        envVars,
		collectionVars: collectionVars,
	})
	if err := s.injectAuthorization(request, finalAuthConfig); err != nil {
		return nil, err
	}

	finalRequestBody := readRequestBody(request)
	response, err := client.Do(request)

	if err != nil {
		slog.Error("Error occurred in HTTP request", "method", request.Method, "url", request.URL.String(), "error", err)
		return &executeResult{request: request, requestBody: finalRequestBody, envVars: envVars, collectionVars: collectionVars}, err
	}

	return &executeResult{
		response:       response,
		request:        request,
		requestBody:    finalRequestBody,
		envVars:        envVars,
		collectionVars: collectionVars,
	}, nil

}

type ResponseData struct {
	StatusCode     int               `json:"statusCode"`
	Headers        map[string]string `json:"headers"`
	RequestHeaders map[string]string `json:"requestHeaders"`
	Body           string            `json:"body"`
	Duration       int64             `json:"duration"`
}

// ExecuteRequest wraps Execute and produces the response payload sent to the frontend.
// It also runs post-response scripts and emits request history events.
func (s *Service) ExecuteRequest(opts ExecutionOptions) (*ResponseData, error) {
	start := time.Now()
	execResult, err := s.Execute(opts)
	duration := time.Since(start).Milliseconds()

	var resp *http.Response
	var req *http.Request
	var finalRequestBody string
	var envVars map[string]string
	var collectionVars map[string]string
	if execResult != nil {
		resp = execResult.response
		req = execResult.request
		finalRequestBody = execResult.requestBody
		envVars = execResult.envVars
		collectionVars = execResult.collectionVars
	}

	var responseData *ResponseData
	var executionError error

	if err != nil {
		slog.Error("HTTP request failed",
			"method", opts.Method,
			"url", opts.URL,
			"duration_ms", duration,
			"error", err)
		executionError = err
	} else {
		defer resp.Body.Close()

		// Read response body
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			slog.Error("Failed to read response body", "url", req.URL.String(), "error", err)
			return nil, err
		}

		// Convert response headers
		respHeaders := make(map[string]string, len(resp.Header))
		for k, v := range resp.Header {
			respHeaders[k] = v[0]
		}

		// Convert request headers (after all injections/scripts)
		// We use the 'req' object returned by Execute which is the final one sent
		reqHeaders := make(map[string]string, len(req.Header))
		for k, v := range req.Header {
			reqHeaders[k] = v[0]
		}

		// Execute post-response script (read-only on request/response, can write session vars)
		if s.scriptManager != nil && opts.PostResponseScript != "" {
			if err := s.scriptManager.ExecutePostResponseWithScope(opts.PostResponseScript, req, resp, string(bodyBytes), duration, envVars, collectionVars); err != nil {
				slog.Warn("Post-response script error", "error", err)
				// Non-fatal: log the error but return the response anyway
			} else {
				slog.Debug("Post-response script executed successfully")
			}
		}

		slog.Info("HTTP request completed",
			"method", req.Method,
			"url", req.URL.String(),
			"status", resp.StatusCode,
			"duration_ms", duration)

		slog.Debug("Response details",
			"headers_count", len(respHeaders),
			"body_length", len(bodyBytes))

		responseData = &ResponseData{
			StatusCode:     resp.StatusCode,
			Headers:        respHeaders,
			RequestHeaders: reqHeaders,
			Body:           string(bodyBytes),
			Duration:       duration,
		}
	}

	// Emit event to frontend for console/history
	//TODO: absolutely no
	if s.ctx != nil {
		finalReqHeaders := opts.Headers
		// If we have the actual executed request, use its headers
		if req != nil {
			finalReqHeaders = make(map[string]any)
			for k, v := range req.Header {
				finalReqHeaders[k] = v[0]
			}
		}

		eventData := map[string]interface{}{
			"options": map[string]interface{}{
				"method":  opts.Method,
				"url":     opts.URL,
				"body":    opts.Body,
				"headers": finalReqHeaders,
				"auth":    opts.Auth,
			},
			"response": responseData,
			"duration": duration,
		}
		if req != nil {
			eventData["options"].(map[string]interface{})["method"] = req.Method
			eventData["options"].(map[string]interface{})["url"] = req.URL.String()
			eventData["options"].(map[string]interface{})["body"] = finalRequestBody
		}
		if executionError != nil {
			eventData["error"] = executionError.Error()
		}
		s.emitEvent("request:executed", eventData)
	}

	return responseData, executionError
}
