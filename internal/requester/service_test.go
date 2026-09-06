// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package requester

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"solo/internal/auth"
	"solo/internal/collection"
	"solo/internal/configuration"
	"solo/internal/environment"
	"solo/internal/host"
	"solo/internal/script"
	"solo/internal/testutil"
	"strings"
	"testing"
)

func newTestService(t *testing.T) (*Service, *configuration.ConfigurationManager, *environment.EnvironmentManager, *script.ScriptManager) {
	t.Helper()

	tempHome := testutil.IsolateUserConfigDir(t)

	configManager, err := configuration.NewConfigurationManager()
	if err != nil {
		t.Fatalf("failed to create configuration manager: %v", err)
	}

	envManager := environment.NewEnvironmentManager(configManager)
	scriptManager := script.NewScriptManager(nil)
	hostManager := host.NewHostManager()
	authManager := auth.NewAuthManager(tempHome)

	service := NewService(configManager, envManager, scriptManager, hostManager, authManager)
	return service, configManager, envManager, scriptManager
}

func saveEnvironment(t *testing.T, envManager *environment.EnvironmentManager, name string, values map[string]string) {
	t.Helper()

	env := environment.NewEnvironment(name)
	for key, value := range values {
		env.Values[key] = environment.ValueType{Value: value, Type: "text"}
	}

	if err := envManager.UpdateEnvironment(&env); err != nil {
		t.Fatalf("failed to save environment: %v", err)
	}
}

func TestService_Execute_BearerAuthentication(t *testing.T) {
	service, _, _, _ := newTestService(t)

	t.Run("resolves and injects a transient token", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
		authConfig := &collection.AuthConfiguration{
			Mode:        collection.AuthModeBearer,
			BearerToken: "{{apiToken}}",
		}
		err := service.injectAuthorization(context.Background(), request, authConfig, resolutionContext{
			collectionVars: map[string]string{"apiToken": "resolved-token"},
		})
		if err != nil {
			t.Fatalf("injectAuthorization() error = %v", err)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer resolved-token" {
			t.Fatalf("Authorization = %q, want %q", got, "Bearer resolved-token")
		}
	})

	t.Run("resolves an encrypted stored token from the selected environment scope", func(t *testing.T) {
		tokenID, err := service.authManager.StoreBearerToken("", "{{apiToken}}")
		if err != nil {
			t.Fatalf("StoreBearerToken() error = %v", err)
		}
		request := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
		err = service.injectAuthorization(context.Background(), request, &collection.AuthConfiguration{
			Mode:          collection.AuthModeBearer,
			BearerTokenID: tokenID,
		}, resolutionContext{
			envVars: map[string]string{"apiToken": "environment-token"},
		})
		if err != nil {
			t.Fatalf("injectAuthorization() error = %v", err)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer environment-token" {
			t.Fatalf("Authorization = %q, want %q", got, "Bearer environment-token")
		}
	})

	t.Run("uses the standard variable scope precedence", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
		err := service.injectAuthorization(context.Background(), request, &collection.AuthConfiguration{
			Mode:        collection.AuthModeBearer,
			BearerToken: "{{apiToken}}",
		}, resolutionContext{
			sessionVars:    map[string]string{"apiToken": "session-token"},
			envVars:        map[string]string{"apiToken": "environment-token"},
			collectionVars: map[string]string{"apiToken": "collection-token"},
		})
		if err != nil {
			t.Fatalf("injectAuthorization() error = %v", err)
		}
		if got := request.Header.Get("Authorization"); got != "Bearer session-token" {
			t.Fatalf("Authorization = %q, want %q", got, "Bearer session-token")
		}
	})

	t.Run("manual Authorization header wins", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
		request.Header.Set("Authorization", "Custom credentials")
		err := service.injectAuthorization(context.Background(), request, &collection.AuthConfiguration{
			Mode:          collection.AuthModeBearer,
			BearerTokenID: "missing-token",
		}, resolutionContext{})
		if err != nil {
			t.Fatalf("injectAuthorization() error = %v", err)
		}
		if got := request.Header.Get("Authorization"); got != "Custom credentials" {
			t.Fatalf("Authorization = %q, want manual value", got)
		}
	})

	t.Run("rejects unresolved tokens", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "https://example.com", nil)
		err := service.injectAuthorization(context.Background(), request, &collection.AuthConfiguration{
			Mode:        collection.AuthModeBearer,
			BearerToken: "{{missing}}",
		}, resolutionContext{})
		if err == nil || !strings.Contains(err.Error(), "unresolved variables") {
			t.Fatalf("injectAuthorization() error = %v, want unresolved variables error", err)
		}
	})
}

func TestService_Execute_ContentTypeHandling(goTest *testing.T) {
	s, _, _, _ := newTestService(goTest)

	tests := []struct {
		name       string
		opts       ExecutionOptions
		wantHeader string
	}{
		{
			name: "do not inject Content-Type when body is present and header is missing",
			opts: ExecutionOptions{
				Method:  "POST",
				URL:     "http://example.com",
				Body:    `{"foo":"bar"}`,
				Headers: map[string]any{},
			},
			wantHeader: "",
		},
		{
			name: "do not override existing Content-Type",
			opts: ExecutionOptions{
				Method: "POST",
				URL:    "http://example.com",
				Body:   `{"foo":"bar"}`,
				Headers: map[string]any{
					"Content-Type": "application/json",
				},
			},
			wantHeader: "application/json",
		},
		{
			name: "keep Content-Type empty when body is empty",
			opts: ExecutionOptions{
				Method:  "POST",
				URL:     "http://example.com",
				Body:    "",
				Headers: map[string]any{},
			},
			wantHeader: "",
		},
		{
			name: "allow pre-request script to set Content-Type",
			opts: ExecutionOptions{
				Method:           "POST",
				URL:              "http://example.com",
				Body:             `{"foo":"bar"}`,
				Headers:          map[string]any{},
				PreRequestScript: `request.headers["Content-Type"] = "application/lua"`,
			},
			wantHeader: "application/lua",
		},
	}

	for _, tt := range tests {
		goTest.Run(tt.name, func(t *testing.T) {
			result, err := s.Execute(context.TODO(), tt.opts)
			if err != nil && result == nil {
				t.Fatalf("expected request to be prepared even on transport error: %v", err)
			}
			if result == nil || result.request == nil {
				t.Fatalf("expected request object to be returned")
			}

			gotHeader := result.request.Header.Get("Content-Type")
			if gotHeader != tt.wantHeader {
				t.Errorf("Content-Type = %q, want %q", gotHeader, tt.wantHeader)
			}
		})
	}
}

func TestService_Execute_ResolvesRequestAgainAfterPreScript(t *testing.T) {
	service, configManager, envManager, scriptManager := newTestService(t)

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if got := r.URL.Path; got != "/items/42" {
			t.Errorf("path = %q, want %q", got, "/items/42")
		}
		if got := string(body); got != "payload-from-pre-script" {
			t.Errorf("body = %q, want %q", got, "payload-from-pre-script")
		}
		if got := r.Header.Get("X-Trace"); got != "trace-42" {
			t.Errorf("X-Trace = %q, want %q", got, "trace-42")
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer apiServer.Close()

	saveEnvironment(t, envManager, "dev", map[string]string{
		"base_url": apiServer.URL,
	})
	if err := configManager.SetSelectedEnvironment("dev"); err != nil {
		t.Fatalf("failed to select environment: %v", err)
	}

	resp, err := service.ExecuteRequest(context.TODO(), ExecutionOptions{
		Method: "POST",
		URL:    "{{base_url}}/{{resource_path}}",
		Body:   "{{payload}}",
		Headers: map[string]any{
			"X-Trace": "{{trace_id}}",
		},
		PreRequestScript: `
			env.set("resource_path", "items/42")
			env.set("payload", "payload-from-pre-script")
			env.set("trace_id", "trace-42")
		`,
	})
	if err != nil {
		t.Fatalf("ExecuteRequest returned error: %v", err)
	}
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
	}

	sessionVars := scriptManager.GetSessionVars()
	if got := sessionVars["resource_path"]; got != "items/42" {
		t.Fatalf("session var resource_path = %q, want %q", got, "items/42")
	}
}

func TestService_Execute_UsesSelectedEnvironmentInsideLua(t *testing.T) {
	service, configManager, envManager, _ := newTestService(t)

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/from-env" {
			t.Errorf("path = %q, want %q", got, "/from-env")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	saveEnvironment(t, envManager, "dev", map[string]string{
		"base_url":     apiServer.URL,
		"resource_key": "from-env",
	})
	if err := configManager.SetSelectedEnvironment("dev"); err != nil {
		t.Fatalf("failed to select environment: %v", err)
	}

	_, err := service.ExecuteRequest(context.TODO(), ExecutionOptions{
		Method:           "GET",
		URL:              "{{base_url}}/placeholder",
		Headers:          map[string]any{},
		PreRequestScript: `request.url = env.get("base_url") .. "/" .. env.get("resource_key")`,
	})
	if err != nil {
		t.Fatalf("ExecuteRequest returned error: %v", err)
	}
}

func TestService_Execute_UsesCollectionVariablesWhenEnvironmentLacksKey(t *testing.T) {
	service, configManager, envManager, _ := newTestService(t)

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/from-collection" {
			t.Errorf("path = %q, want %q", got, "/from-collection")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	saveEnvironment(t, envManager, "dev", map[string]string{})
	if err := configManager.SetSelectedEnvironment("dev"); err != nil {
		t.Fatalf("failed to select environment: %v", err)
	}

	_, err := service.ExecuteRequest(context.TODO(), ExecutionOptions{
		Method:  "GET",
		URL:     "{{base_url}}/from-collection",
		Headers: map[string]any{},
		CollectionVariables: map[string]string{
			"base_url": apiServer.URL,
		},
	})
	if err != nil {
		t.Fatalf("ExecuteRequest returned error: %v", err)
	}
}

func TestService_Execute_EnvironmentWinsOverCollectionWhenValued(t *testing.T) {
	service, configManager, envManager, _ := newTestService(t)

	collectionServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("collection server should not be used when environment is valued")
	}))
	defer collectionServer.Close()

	envServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/from-env" {
			t.Errorf("path = %q, want %q", got, "/from-env")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer envServer.Close()

	saveEnvironment(t, envManager, "dev", map[string]string{
		"base_url": envServer.URL,
	})
	if err := configManager.SetSelectedEnvironment("dev"); err != nil {
		t.Fatalf("failed to select environment: %v", err)
	}

	_, err := service.ExecuteRequest(context.TODO(), ExecutionOptions{
		Method:  "GET",
		URL:     "{{base_url}}/from-env",
		Headers: map[string]any{},
		CollectionVariables: map[string]string{
			"base_url": collectionServer.URL,
		},
	})
	if err != nil {
		t.Fatalf("ExecuteRequest returned error: %v", err)
	}
}

func TestService_Execute_UsesCollectionVariablesInsideLuaFallback(t *testing.T) {
	service, configManager, envManager, _ := newTestService(t)

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/from-lua-collection" {
			t.Errorf("path = %q, want %q", got, "/from-lua-collection")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	saveEnvironment(t, envManager, "dev", map[string]string{
		"base_url": "",
	})
	if err := configManager.SetSelectedEnvironment("dev"); err != nil {
		t.Fatalf("failed to select environment: %v", err)
	}

	_, err := service.ExecuteRequest(context.TODO(), ExecutionOptions{
		Method:           "GET",
		URL:              "http://placeholder.invalid",
		Headers:          map[string]any{},
		PreRequestScript: `request.url = env.get("base_url") .. "/from-lua-collection"`,
		CollectionVariables: map[string]string{
			"base_url": apiServer.URL,
		},
	})
	if err != nil {
		t.Fatalf("ExecuteRequest returned error: %v", err)
	}
}

func TestService_Execute_ResolvesAuthAfterPreScript(t *testing.T) {
	service, configManager, envManager, _ := newTestService(t)

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("failed to parse auth form: %v", err)
		}
		clientID := r.Form.Get("client_id")
		if clientID != "client-from-pre-script" {
			t.Fatalf("client_id = %q, want %q", clientID, "client-from-pre-script")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"token-from-` + clientID + `"}`))
	}))
	defer tokenServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer token-from-client-from-pre-script" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer token-from-client-from-pre-script")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	saveEnvironment(t, envManager, "dev", map[string]string{
		"base_url": apiServer.URL,
		"auth_url": tokenServer.URL,
	})
	if err := configManager.SetSelectedEnvironment("dev"); err != nil {
		t.Fatalf("failed to select environment: %v", err)
	}

	_, err := service.ExecuteRequest(context.TODO(), ExecutionOptions{
		Method:  "GET",
		URL:     "{{base_url}}/secure",
		Headers: map[string]any{},
		Auth: &collection.AuthConfiguration{
			Enabled:   true,
			TokenURL:  "{{auth_url}}",
			TokenPath: "access_token",
			Template: map[string]string{
				"client_id": "{{client_id}}",
			},
		},
		PreRequestScript: `env.set("client_id", "client-from-pre-script")`,
	})
	if err != nil {
		t.Fatalf("ExecuteRequest returned error: %v", err)
	}
}

func TestService_Execute_PreRequestAuthorizationWinsOverOAuth(t *testing.T) {
	service, configManager, envManager, _ := newTestService(t)

	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("token endpoint should not be called when Authorization is already set")
	}))
	defer tokenServer.Close()

	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer manual-token" {
			t.Errorf("Authorization = %q, want %q", got, "Bearer manual-token")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer apiServer.Close()

	saveEnvironment(t, envManager, "dev", map[string]string{
		"base_url": apiServer.URL,
		"auth_url": tokenServer.URL,
	})
	if err := configManager.SetSelectedEnvironment("dev"); err != nil {
		t.Fatalf("failed to select environment: %v", err)
	}

	_, err := service.ExecuteRequest(context.TODO(), ExecutionOptions{
		Method:  "GET",
		URL:     "{{base_url}}/manual-auth",
		Headers: map[string]any{},
		Auth: &collection.AuthConfiguration{
			Enabled:   true,
			TokenURL:  "{{auth_url}}",
			TokenPath: "access_token",
			Template: map[string]string{
				"client_id": "unused",
			},
		},
		PreRequestScript: `request.headers["Authorization"] = "Bearer manual-token"`,
	})
	if err != nil {
		t.Fatalf("ExecuteRequest returned error: %v", err)
	}
}

func TestService_Execute_LeavesUnresolvedPlaceholdersUntouched(t *testing.T) {
	service, _, _, _ := newTestService(t)

	result, err := service.Execute(context.TODO(), ExecutionOptions{
		Method:  "GET",
		URL:     "http://example.com/{{missing}}",
		Headers: map[string]any{},
	})
	if err == nil && result == nil {
		t.Fatalf("expected a prepared request or a transport error")
	}
	if result == nil || result.request == nil {
		t.Fatalf("expected prepared request to inspect unresolved placeholder")
	}
	if got := result.request.URL.String(); got != "http://example.com/%7B%7Bmissing%7D%7D" {
		t.Fatalf("url = %q, want %q", got, "http://example.com/%7B%7Bmissing%7D%7D")
	}
}
