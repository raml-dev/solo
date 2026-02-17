package script

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func TestScriptManager_SessionPersistence(t *testing.T) {
	sm := NewScriptManager(nil)
	req, _ := http.NewRequest("GET", "http://dummy", nil)

	// Script 1: Set a variable
	err := sm.ExecutePreRequest(`env.foo = "bar"`, req)
	if err != nil {
		t.Fatalf("Script 1 failed: %v", err)
	}

	// Script 2: Read the variable
	err = sm.ExecutePreRequest(`
		if env.foo ~= "bar" then
			error("env.foo is not bar")
		end
	`, req)
	if err != nil {
		t.Fatalf("Script 2 failed: %v", err)
	}

	// Verify in Go map via getter
	vars := sm.GetSessionVars()
	if val, ok := vars["foo"]; !ok || val != "bar" {
		t.Errorf("sessionVars['foo'] = %v, want 'bar'", val)
	}
}

func TestScriptManager_ExecutePreRequest(t *testing.T) {
	sm := NewScriptManager(nil)
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	req.Header.Set("Original", "True")

	script := `
		request.headers["Added-By-Lua"] = "Yes"
		request.method = "POST"
        request.body = "modified body"
	`

	err := sm.ExecutePreRequest(script, req)
	if err != nil {
		t.Fatalf("ExecutePreRequest failed: %v", err)
	}

	if req.Header.Get("Added-By-Lua") != "Yes" {
		t.Error("Header 'Added-By-Lua' not set by script")
	}
	if req.Method != "POST" {
		t.Errorf("Method = %s, want POST", req.Method)
	}

	if req.Body == nil {
		t.Error("Body is nil, expected modified body")
	} else {
		bodyBytes, _ := io.ReadAll(req.Body)
		if string(bodyBytes) != "modified body" {
			t.Errorf("Body = %s, want 'modified body'", string(bodyBytes))
		}
	}
}

func TestScriptManager_ExecutePostResponse(t *testing.T) {
	sm := NewScriptManager(nil)
	req, _ := http.NewRequest("GET", "http://example.com", nil)
	resp := &http.Response{
		StatusCode: 201,
		Header:     make(http.Header),
	}
	resp.Header.Set("Content-Type", "application/json")
	body := `{"id": 123}`

	script := `
		if response.status ~= 201 then
			error("Status is not 201")
		end
		if response.headers["Content-Type"] ~= "application/json" then
			error("Content-Type mismatch")
		end
		env.extractedId = "123" 
	`

	err := sm.ExecutePostResponse(script, req, resp, body, 100)
	if err != nil {
		t.Fatalf("ExecutePostResponse failed: %v", err)
	}

	// Access internal map via getter for verification
	vars := sm.GetSessionVars()
	if val := vars["extractedId"]; val != "123" {
		t.Errorf("extractedId = %v, want '123'", val)
	}
}

func TestScriptManager_EnvironmentFallback(t *testing.T) {
	sm := NewScriptManager(nil)
	sm.SetEnvironment(map[string]string{"baseUrl": "http://api.dev"})

	req, _ := http.NewRequest("GET", "http://dummy", nil)

	// Script reads env.baseUrl
	script := `
        if env.baseUrl ~= "http://api.dev" then
            error("baseUrl mismatch")
        end
        -- Override in session
        env.baseUrl = "http://override.com"
    `

	err := sm.ExecutePreRequest(script, req)
	if err != nil {
		t.Fatalf("Script failed: %v", err)
	}

	// Check that session var took precedence for subsequent reads
	vars := sm.GetSessionVars()
	if vars["baseUrl"] != "http://override.com" {
		t.Errorf("Session var override failed")
	}

	script2 := `
        if env.baseUrl ~= "http://override.com" then
             error("Session override not persisting")
        end
    `
	err = sm.ExecutePreRequest(script2, req)
	if err != nil {
		t.Fatalf("Script 2 failed: %v", err)
	}
}

func TestScriptManager_Concurrency(t *testing.T) {
	sm := NewScriptManager(nil)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("GET", "http://example.com", nil)
			sm.ExecutePreRequest(`env.counter = "updated"`, req)
		}()
	}

	wg.Wait()
	vars := sm.GetSessionVars()
	if vars["counter"] != "updated" {
		t.Error("Concurrency test failed to update sessionVars")
	}
}

func TestScriptManager_EnvAPI(t *testing.T) {
	sm := NewScriptManager(nil)
	req, _ := http.NewRequest("GET", "http://dummy", nil)

	// Test env.set, env.get, env.log
	script := `
        env.set("api_key", "secret123")
        local val = env.get("api_key")
        if val ~= "secret123" then
            error("env.get returned " .. tostring(val))
        end
        env.log("Test log message")
    `
	if err := sm.ExecutePreRequest(script, req); err != nil {
		t.Fatalf("EnvAPI test failed: %v", err)
	}

	// Verify side effect
	vars := sm.GetSessionVars()
	if vars["api_key"] != "secret123" {
		t.Errorf("env.set did not update sessionVars")
	}
}

func TestScriptManager_PanicRecovery(t *testing.T) {
	sm := NewScriptManager(nil)
	req, _ := http.NewRequest("GET", "http://dummy", nil)

	// Register a function that panics
	L := sm.GetState()
	L.SetGlobal("panic_me", L.NewFunction(func(L *lua.LState) int {
		panic("boom")
	}))

	err := sm.ExecutePreRequest("panic_me()", req)
	if err == nil {
		t.Error("Expected error from panic, got nil")
	} else {
		// Check error message format
		// "lua execution panic: boom"
		if !strings.Contains(err.Error(), "boom") {
			t.Errorf("Expected error containing 'boom', got '%v'", err)
		}
	}
}
