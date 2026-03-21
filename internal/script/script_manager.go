// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only

package script

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"
)

type ScriptManager struct {
	state       *lua.LState
	mutex       sync.Mutex
	sessionVars map[string]string // vars modified/created by Lua (ephemeral and in-memory only)
	currentEnv  map[string]string // vars of the selected env (Lua is read-only on these)
	wailsCtx    context.Context
}

func NewScriptManager(ctx context.Context) *ScriptManager {
	runCtx := ctx
	if runCtx == nil {
		runCtx = context.Background()
	}

	L := lua.NewState(lua.Options{SkipOpenLibs: true})

	// Load allowed libraries
	for _, pair := range []struct {
		n string
		f lua.LGFunction
	}{
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.StringLibName, lua.OpenString},
		{lua.MathLibName, lua.OpenMath},
	} {
		if err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(pair.f),
			NRet:    0,
			Protect: true,
		}, lua.LString(pair.n)); err != nil {
			panic(err)
		}
	}

	// Remove dangerous functions from base lib
	for _, name := range []string{
		"dofile",
		"loadfile",
		"load",
		"rawget",
		"rawset",
		"rawequal",
		"rawlen",
	} {
		L.SetGlobal(name, lua.LNil)
	}

	// Instruction limit to prevent CPU exhaustion / infinite loops
	L.SetMx(10_000_000)

	// Cancel the Lua state if the context is done
	go func() {
		<-runCtx.Done()
		L.Close()
	}()

	sm := &ScriptManager{
		state:       L,
		mutex:       sync.Mutex{},
		sessionVars: make(map[string]string),
		currentEnv:  make(map[string]string),
		wailsCtx:    ctx,
	}

	// Register global 'env' table and API
	NewEnvAPI(sm).Register(L)

	return sm
}

// GetState safely returns the Lua state instance.
// Note: The caller is responsible for thread safety if accessing the state directly outside of ScriptManager's methods.
func (sm *ScriptManager) GetState() *lua.LState {
	return sm.state
}

// ExecutePreRequest runs a Lua script before a request is sent.
// It exposes a global 'request' object which can be modified.
func (sm *ScriptManager) ExecutePreRequest(script string, req *http.Request) (err error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("lua execution panic: %v", r)
			slog.Error("Lua Panic recovered", "panic", r)
		}
	}()

	if script == "" {
		return nil
	}

	// 500ms timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	sm.state.SetContext(ctx)
	defer sm.state.SetContext(context.TODO())

	reqTable := RequestToLua(sm, req)
	sm.state.SetGlobal("request", reqTable)
	defer sm.state.SetGlobal("request", lua.LNil) // Cleanup

	if err := sm.state.DoString(script); err != nil {
		return err
	}

	// Update request with potential changes from Lua
	newReqVal := sm.state.GetGlobal("request")
	if newReqTable, ok := newReqVal.(*lua.LTable); ok {
		LuaToRequest(sm, newReqTable, req)
	}

	return nil
}

// ExecutePostResponse runs a Lua script after a response is received.
// It exposes global 'request' and 'response' objects (read-only context mostly).
func (sm *ScriptManager) ExecutePostResponse(script string, req *http.Request, resp *http.Response, body string, responseTime int64) (err error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	defer func() {
		// Force Garbage Collection to clean up large response bodies
		sm.state.DoString("collectgarbage()")

		if r := recover(); r != nil {
			err = fmt.Errorf("lua execution panic: %v", r)
			slog.Error("Lua Panic recovered", "panic", r)
		}
	}()

	if script == "" {
		return nil
	}

	// 500ms timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond) //TODO: does make sense to have this param as configurable?
	defer cancel()
	sm.state.SetContext(ctx)
	defer sm.state.SetContext(context.TODO())

	// Expose request (read-only context)
	reqTable := RequestToLua(sm, req)
	sm.state.SetGlobal("request", reqTable)
	defer sm.state.SetGlobal("request", lua.LNil)

	// Expose response
	respTable := ResponseToLua(sm, resp, body, responseTime)
	sm.state.SetGlobal("response", respTable)
	defer sm.state.SetGlobal("response", lua.LNil)

	return sm.state.DoString(script)
}

// SetEnvironment updates the current environment variables (read-only for Lua).
func (sm *ScriptManager) SetEnvironment(env map[string]string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.currentEnv = env
}

// SetContext injects the Wails runtime context so scripts can emit events.
// Called from app.startup() after the context is available.
func (sm *ScriptManager) SetContext(ctx context.Context) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.wailsCtx = ctx
}

// GetSessionVars returns a copy of the current session variables.
func (sm *ScriptManager) GetSessionVars() map[string]string {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	copyVars := make(map[string]string, len(sm.sessionVars))
	for k, v := range sm.sessionVars {
		copyVars[k] = v
	}
	return copyVars
}

// RemoveSessionVar removes a single session variable by key.
func (sm *ScriptManager) RemoveSessionVar(key string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	delete(sm.sessionVars, key)
}

// ClearSessionVars removes all session variables.
func (sm *ScriptManager) ClearSessionVars() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	sm.sessionVars = make(map[string]string)
}
