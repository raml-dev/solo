package requester

import (
	"bytes"
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
	"solo/internal/host"
	"solo/internal/script"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ExecutionOptions encapsulates all parameters required to execute a request
type ExecutionOptions struct {
	Method             string                                 `json:"method"`
	URL                string                                 `json:"url"`
	Body               string                                 `json:"body"`
	Headers            map[string]any                         `json:"headers"`
	Cookies            map[string]any                         `json:"cookies"`
	Settings           *configuration.RequestSettingsOverride `json:"settings,omitempty"`
	Auth               *collection.AuthConfiguration          `json:"auth,omitempty"`
	PreRequestScript   string                                 `json:"preRequestScript,omitempty"`
	PostResponseScript string                                 `json:"postResponseScript,omitempty"`
}

type Service struct {
	hostManager   *host.HostManager
	configManager *configuration.ConfigurationManager
	scriptManager *script.ScriptManager
	authManager   *auth.AuthManager
	ctx           context.Context
}

func NewService(cm *configuration.ConfigurationManager, sm *script.ScriptManager, hm *host.HostManager, am *auth.AuthManager) *Service {
	return &Service{
		hostManager:   hm,
		configManager: cm,
		scriptManager: sm,
		authManager:   am,
	}
}

func (s *Service) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *Service) Execute(opts ExecutionOptions) (*http.Response, *http.Request, error) {
	slog.Debug("Preparing HTTP request",
		"method", opts.Method,
		"url", opts.URL,
		"headers_count", len(opts.Headers),
		"body_length", len(opts.Body))

	// Get base client from HostManager (handles custom Certs/Transport caching)
	client, err := s.hostManager.GetClientForUrl(opts.URL)
	if err != nil {
		slog.Error("Failed to get HTTP client", "url", opts.URL, "error", err)
		return nil, nil, err
	}

	// Apply Global Configuration with Local Overrides
	if s.configManager != nil {
		cfg := s.configManager.Get()
		overrides := opts.Settings

		// Helper to check override existence
		hasOverride := overrides != nil

		// 1. Timeout
		timeout := cfg.Request.TimeoutSeconds
		if hasOverride && overrides.TimeoutSeconds != nil {
			timeout = *overrides.TimeoutSeconds
		}
		if timeout > 0 {
			client.Timeout = time.Duration(timeout) * time.Second
		}

		// 2. Redirects
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

		// 3. Proxy (Requires Transport modification, careful with shared transport)
		proxyURL := cfg.Request.ProxyURL
		if hasOverride && overrides.ProxyURL != nil {
			proxyURL = *overrides.ProxyURL
		}

		if proxyURL != "" {
			parsedProxyUrl, err := url.Parse(proxyURL)
			if err == nil {
				// We need to clone the transport to avoid affecting other clients/threads sharing the pool
				if t, ok := client.Transport.(*http.Transport); ok {
					newTransport := t.Clone()
					newTransport.Proxy = http.ProxyURL(parsedProxyUrl)
					client.Transport = newTransport
				}
			}
		}

		// 4. SSL Validation
		validateSSL := cfg.Request.ValidateSSL
		if hasOverride && overrides.ValidateSSL != nil {
			validateSSL = *overrides.ValidateSSL
		}

		if !validateSSL {
			if t, ok := client.Transport.(*http.Transport); ok {
				newTransport := t.Clone()
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

	request, err := http.NewRequest(opts.Method, opts.URL, bytes.NewReader([]byte(opts.Body)))
	if err != nil {
		slog.Error("Failed to create HTTP request", "method", opts.Method, "url", opts.URL, "error", err)
		return nil, nil, err
	}

	// 1. Set User-Agent and other defaults
	if s.configManager != nil {
		cfg := s.configManager.Get()
		if cfg.Request.DefaultUserAgent != "" {
			request.Header.Set("User-Agent", cfg.Request.DefaultUserAgent)
		}
	}

	// 2. Apply User Defined Headers
	for k, v := range opts.Headers {
		vStr, ok := v.(string)
		if ok {
			request.Header.Set(k, vStr)
		}
	}

	// 3. Handle OAuth2 Authentication (Injected LAST to take precedence)
	if opts.Auth != nil && opts.Auth.Enabled && s.authManager != nil {
		token, refreshed, err := s.authManager.GetOrRefreshToken(*opts.Auth)
		if err != nil {
			slog.Error("Failed to get/refresh OAuth2 token", "error", err)
			return nil, nil, fmt.Errorf("authentication failed: %w", err)
		}

		if token != "" {
			request.Header.Set("Authorization", "Bearer "+token)
			slog.Debug("Injected OAuth2 Bearer token", "refreshed", refreshed)
		}
	}

	// 4. Set Cookies
	for k, v := range opts.Cookies {
		vStr, ok := v.(string)
		if ok {
			request.AddCookie(&http.Cookie{Name: k, Value: vStr})
		}
	}

	// Host Specific Cookies
	parsedUrl, err := url.Parse(opts.URL)
	if err == nil {
		if hostCfg, ok := s.hostManager.GetHost(parsedUrl.Host); ok {
			for name, value := range hostCfg.Cookies {
				request.AddCookie(&http.Cookie{Name: name, Value: value})
			}
		}
	}

	// Execute pre-request script (may mutate method, url, headers, body)
	if s.scriptManager != nil && opts.PreRequestScript != "" {
		if err := s.scriptManager.ExecutePreRequest(opts.PreRequestScript, request); err != nil {
			slog.Warn("Pre-request script error", "error", err)
			return nil, nil, fmt.Errorf("pre-request script error: %w", err)
		}
		slog.Debug("Pre-request script executed successfully")
	}

	response, err := client.Do(request)

	if err != nil {
		slog.Error("Error occurred in HTTP request", "method", opts.Method, "url", opts.URL, "error", err)
		return nil, request, err
	}

	return response, request, nil

}

type ResponseData struct {
	StatusCode     int               `json:"statusCode"`
	Headers        map[string]string `json:"headers"`
	RequestHeaders map[string]string `json:"requestHeaders"`
	Body           string            `json:"body"`
	Duration       int64             `json:"duration"`
}

func (s *Service) ExecuteRequest(opts ExecutionOptions) (*ResponseData, error) {
	start := time.Now()
	resp, req, err := s.Execute(opts)
	duration := time.Since(start).Milliseconds()

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
			slog.Error("Failed to read response body", "url", opts.URL, "error", err)
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
			if err := s.scriptManager.ExecutePostResponse(opts.PostResponseScript, req, resp, string(bodyBytes), duration); err != nil {
				slog.Warn("Post-response script error", "error", err)
				// Non-fatal: log the error but return the response anyway
			} else {
				slog.Debug("Post-response script executed successfully")
			}
		}

		slog.Info("HTTP request completed",
			"method", opts.Method,
			"url", opts.URL,
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
	if s.ctx != context.TODO() {
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
		if executionError != nil {
			eventData["error"] = executionError.Error()
		}
		runtime.EventsEmit(s.ctx, "request:executed", eventData)
	}

	return responseData, executionError
}
