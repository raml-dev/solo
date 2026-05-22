// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package requester

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// ---- resolveParams ----------------------------------------------------------

func TestResolveParams_ResolvesKeysAndValues(t *testing.T) {
	ctx := resolutionContext{
		envVars: map[string]string{
			"env_key":   "resolved-key",
			"env_value": "resolved-value",
		},
	}

	params := map[string]string{
		"{{env_key}}":   "{{env_value}}",
		"static-key":    "static-value",
		"{{missing}}":   "literal",
	}

	got := resolveParams(params, ctx)

	cases := []struct {
		key  string
		want string
	}{
		{"resolved-key", "resolved-value"},
		{"static-key", "static-value"},
		{"{{missing}}", "literal"}, // unresolved key stays as-is
	}

	for _, tc := range cases {
		if v, ok := got[tc.key]; !ok {
			t.Errorf("key %q not found in resolved params", tc.key)
		} else if v != tc.want {
			t.Errorf("params[%q] = %q, want %q", tc.key, v, tc.want)
		}
	}
}

func TestResolveParams_NilMap(t *testing.T) {
	got := resolveParams(nil, resolutionContext{})
	if len(got) != 0 {
		t.Errorf("expected empty map for nil input, got %v", got)
	}
}

func TestResolveParams_EmptyMap(t *testing.T) {
	got := resolveParams(map[string]string{}, resolutionContext{})
	if len(got) != 0 {
		t.Errorf("expected empty map for empty input, got %v", got)
	}
}

func TestResolveParams_SessionVarsWinOverEnvVars(t *testing.T) {
	ctx := resolutionContext{
		sessionVars: map[string]string{"key": "from-session"},
		envVars:     map[string]string{"key": "from-env"},
	}

	got := resolveParams(map[string]string{"p": "{{key}}"}, ctx)
	if v := got["p"]; v != "from-session" {
		t.Errorf("expected session var to win, got %q", v)
	}
}

// ---- resolve (end-to-end via ExecutionOptions) ------------------------------

func TestResolve_ParamsAreResolved(t *testing.T) {
	ctx := resolutionContext{
		envVars: map[string]string{"page": "2", "limit": "50"},
	}
	opts := ExecutionOptions{
		Method: "GET",
		URL:    "http://example.com",
		Params: map[string]string{
			"page":  "{{page}}",
			"limit": "{{limit}}",
		},
	}

	resolved := opts.resolve(ctx)

	if v := resolved.Params["page"]; v != "2" {
		t.Errorf("page = %q, want %q", v, "2")
	}
	if v := resolved.Params["limit"]; v != "50" {
		t.Errorf("limit = %q, want %q", v, "50")
	}
}

// ---- buildRequestFromOptions ------------------------------------------------

func TestBuildRequestFromOptions_AppendsParams(t *testing.T) {
	opts := ExecutionOptions{
		Method: "GET",
		URL:    "http://example.com/items",
		Params: map[string]string{
			"foo": "bar",
			"baz": "qux",
		},
	}

	req, err := buildRequestFromOptions(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	q, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		t.Fatalf("failed to parse query: %v", err)
	}

	if v := q.Get("foo"); v != "bar" {
		t.Errorf("param foo = %q, want %q", v, "bar")
	}
	if v := q.Get("baz"); v != "qux" {
		t.Errorf("param baz = %q, want %q", v, "qux")
	}
}

func TestBuildRequestFromOptions_EmptyValueParamIncluded(t *testing.T) {
	opts := ExecutionOptions{
		Method: "GET",
		URL:    "http://example.com/items",
		Params: map[string]string{
			"emptykey": "",
		},
	}

	req, err := buildRequestFromOptions(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// url.Values.Has returns true even for empty-value keys
	q, _ := url.ParseQuery(req.URL.RawQuery)
	if !q.Has("emptykey") {
		t.Errorf("expected emptykey to be present in query string, got RawQuery=%q", req.URL.RawQuery)
	}
	if v := q.Get("emptykey"); v != "" {
		t.Errorf("emptykey value = %q, want empty string", v)
	}
}

func TestBuildRequestFromOptions_NilParamsLeavesURLUntouched(t *testing.T) {
	opts := ExecutionOptions{
		Method: "GET",
		URL:    "http://example.com/items?existing=1",
	}

	req, err := buildRequestFromOptions(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.URL.RawQuery != "existing=1" {
		t.Errorf("RawQuery = %q, want %q", req.URL.RawQuery, "existing=1")
	}
}

func TestBuildRequestFromOptions_ParamsMergeWithExistingQueryString(t *testing.T) {
	opts := ExecutionOptions{
		Method: "GET",
		URL:    "http://example.com/items?existing=1",
		Params: map[string]string{
			"new": "2",
		},
	}

	req, err := buildRequestFromOptions(context.Background(), opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	q, _ := url.ParseQuery(req.URL.RawQuery)
	if v := q.Get("existing"); v != "1" {
		t.Errorf("existing param = %q, want %q", v, "1")
	}
	if v := q.Get("new"); v != "2" {
		t.Errorf("new param = %q, want %q", v, "2")
	}
}

// ---- integration: Execute appends params to the actual HTTP request ---------

func TestService_Execute_ParamsAppendedToURL(t *testing.T) {
	var capturedQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc, _, _, _ := newTestService(t)

	_, err := svc.ExecuteRequest(context.Background(), ExecutionOptions{
		Method: "GET",
		URL:    server.URL + "/search",
		Params: map[string]string{
			"q":    "hello world",
			"page": "1",
		},
	})
	if err != nil {
		t.Fatalf("ExecuteRequest returned error: %v", err)
	}

	if v := capturedQuery.Get("q"); v != "hello world" {
		t.Errorf("q = %q, want %q", v, "hello world")
	}
	if v := capturedQuery.Get("page"); v != "1" {
		t.Errorf("page = %q, want %q", v, "1")
	}
}

func TestService_Execute_ParamsPlaceholdersResolvedFromEnv(t *testing.T) {
	var capturedQuery url.Values
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc, configManager, envManager, _ := newTestService(t)
	saveEnvironment(t, envManager, "dev", map[string]string{
		"base_url": server.URL,
		"size":     "100",
	})
	if err := configManager.SetSelectedEnvironment("dev"); err != nil {
		t.Fatalf("failed to select environment: %v", err)
	}

	_, err := svc.ExecuteRequest(context.Background(), ExecutionOptions{
		Method: "GET",
		URL:    "{{base_url}}/data",
		Params: map[string]string{
			"size": "{{size}}",
		},
	})
	if err != nil {
		t.Fatalf("ExecuteRequest returned error: %v", err)
	}

	if v := capturedQuery.Get("size"); v != "100" {
		t.Errorf("size = %q, want %q", v, "100")
	}
}
