package requester

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
	"yapla/internal/configuration"
	"yapla/internal/host"
)

// ExecutionOptions encapsulates all parameters required to execute a request
type ExecutionOptions struct {
	Method   string
	URL      string
	Body     string
	Headers  map[string]any
	Cookies  map[string]any
	Settings *configuration.RequestSettingsOverride
}

type Service struct {
	hostManager   *host.HostManager
	configManager *configuration.ConfigurationManager
}

func NewService(cm *configuration.ConfigurationManager) *Service {
	return &Service{
		hostManager:   host.NewHostManager(),
		configManager: cm,
	}
}

func (s *Service) Execute(opts ExecutionOptions) (*http.Response, error) {
	slog.Debug("Preparing HTTP request",
		"method", opts.Method,
		"url", opts.URL,
		"headers_count", len(opts.Headers),
		"body_length", len(opts.Body))

	// Get base client from HostManager (handles custom Certs/Transport caching)
	client, err := s.hostManager.GetClientForUrl(opts.URL)
	if err != nil {
		slog.Error("Failed to get HTTP client", "url", opts.URL, "error", err)
		return nil, err
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
		return nil, err
	}

	// Set Default User-Agent if not overridden
	if s.configManager != nil {
		cfg := s.configManager.Get()
		if cfg.Request.DefaultUserAgent != "" {
			request.Header.Set("User-Agent", cfg.Request.DefaultUserAgent)
		}
	}

	for k, v := range opts.Headers {
		v, ok := v.(string)

		if !ok {
			return nil, fmt.Errorf("header %s  with value %v is not parsable as string", k, v)
		} else {
			request.Header.Set(k, v) // Use Set to override default UA if present in headers
		}

	}

	for k, v := range opts.Cookies {
		v, ok := v.(string)

		if !ok {
			return nil, fmt.Errorf("cookie %s  with value %v is not parsable as string", k, v)
		} else {
			request.AddCookie(&http.Cookie{Name: k, Value: v})
		}

	}

	response, err := client.Do(request)

	if err != nil {
		slog.Error("Error occurred in HTTP request", "method", opts.Method, "url", opts.URL, "error", err)
		return nil, err
	}

	return response, err

}

type ResponseData struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Duration   int64             `json:"duration"`
}

func (s *Service) ExecuteRequest(opts ExecutionOptions) (*ResponseData, error) {
	start := time.Now()
	resp, err := s.Execute(opts)
	duration := time.Since(start).Milliseconds()

	if err != nil {
		slog.Error("HTTP request failed",
			"method", opts.Method,
			"url", opts.URL,
			"duration_ms", duration,
			"error", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read response body", "url", opts.URL, "error", err)
		return nil, err
	}

	// Convert headers
	respHeaders := make(map[string]string, len(resp.Header))
	for k, v := range resp.Header {
		respHeaders[k] = v[0]
	}

	slog.Info("HTTP request completed",
		"method", opts.Method,
		"url", opts.URL,
		"status", resp.StatusCode,
		"duration_ms", duration)

	slog.Debug("Response details",
		"headers_count", len(respHeaders),
		"body_length", len(bodyBytes))

	return &ResponseData{
		StatusCode: resp.StatusCode,
		Headers:    respHeaders,
		Body:       string(bodyBytes),
		Duration:   duration,
	}, nil
}
