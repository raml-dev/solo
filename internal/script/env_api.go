// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package script

import (
	"log/slog"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	lua "github.com/yuin/gopher-lua"
)

type EnvAPI struct {
	sm *ScriptManager
}

func NewEnvAPI(sm *ScriptManager) *EnvAPI {
	return &EnvAPI{sm: sm}
}

// Register initializes the 'env' global table and its functions
func (api *EnvAPI) Register(L *lua.LState) {
	envTable := L.NewTable()
	mt := L.NewTable()

	// Register explicit methods: env.get, env.set, env.log
	L.SetField(envTable, "get", L.NewFunction(api.Get))
	L.SetField(envTable, "set", L.NewFunction(api.Set))
	L.SetField(envTable, "log", L.NewFunction(api.Log))

	// Configure metatable for dynamic access (env.var)
	L.SetField(mt, "__index", L.NewFunction(api.Index))
	L.SetField(mt, "__newindex", L.NewFunction(api.NewIndex))

	L.SetMetatable(envTable, mt)
	L.SetGlobal("env", envTable)
}

// Get implements env.get(key)
func (api *EnvAPI) Get(L *lua.LState) int {
	key := L.CheckString(1)

	// 1. Check sessionVars
	if val, ok := api.sm.sessionVars[key]; ok {
		L.Push(lua.LString(val))
		return 1
	}

	// 2. Check currentEnv/currentCollection fallback chain
	if val, ok := api.sm.resolveScopedValueLocked(key); ok {
		L.Push(lua.LString(val))
		return 1
	}

	// 3. Return nil
	L.Push(lua.LNil)
	return 1
}

// Set implements env.set(key, value)
func (api *EnvAPI) Set(L *lua.LState) int {
	key := L.CheckString(1)
	val := L.CheckString(2) // value must be string as per spec

	api.sm.sessionVars[key] = val

	// Create copy for event
	varsCopy := make(map[string]string, len(api.sm.sessionVars))
	for k, v := range api.sm.sessionVars {
		varsCopy[k] = v
	}

	// Emit event if context is available
	if api.sm.wailsCtx != nil {
		runtime.EventsEmit(api.sm.wailsCtx, "session_vars_updated", varsCopy)
	}

	return 0
}

// Log implements env.log(message)
func (api *EnvAPI) Log(L *lua.LState) int {
	msg := L.CheckString(1)
	slog.Info("[LUA] " + msg)
	return 0
}

// Index implements __index metamethod for env.var access
func (api *EnvAPI) Index(L *lua.LState) int {
	key := L.CheckString(2)

	if val, ok := api.sm.sessionVars[key]; ok {
		L.Push(lua.LString(val))
		return 1
	}
	if val, ok := api.sm.resolveScopedValueLocked(key); ok {
		L.Push(lua.LString(val))
		return 1
	}
	L.Push(lua.LNil)
	return 1
}

// NewIndex implements __newindex metamethod for env.var = val assignment
func (api *EnvAPI) NewIndex(L *lua.LState) int {
	key := L.CheckString(2)
	val := L.CheckAny(3)

	api.sm.sessionVars[key] = val.String()

	// Create copy for event
	varsCopy := make(map[string]string, len(api.sm.sessionVars))
	for k, v := range api.sm.sessionVars {
		varsCopy[k] = v
	}

	if api.sm.wailsCtx != nil {
		runtime.EventsEmit(api.sm.wailsCtx, "session_vars_updated", varsCopy)
	}

	return 0
}
