// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package script

import (
	"bytes"
	"io"
	"net/http"
	"net/url"

	lua "github.com/yuin/gopher-lua"
)

// RequestToLua converts an http.Request to a Lua table using the provided ScriptManager's state
func RequestToLua(sm *ScriptManager, req *http.Request) *lua.LTable {
	L := sm.GetState()
	t := L.NewTable()

	// Method
	L.SetField(t, "method", lua.LString(req.Method))

	// URL
	L.SetField(t, "url", lua.LString(req.URL.String()))

	// Body
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body.Close()
		req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body for Go
	}
	L.SetField(t, "body", lua.LString(string(bodyBytes)))

	// Headers
	headersTable := L.NewTable()
	for k, v := range req.Header {
		if len(v) > 0 {
			L.SetField(headersTable, k, lua.LString(v[0]))
		}
	}
	L.SetField(t, "headers", headersTable)

	return t
}

// LuaToRequest updates an http.Request based on a Lua table using the provided ScriptManager's state
func LuaToRequest(sm *ScriptManager, t *lua.LTable, req *http.Request) {
	L := sm.GetState()

	// Method
	if val := L.GetField(t, "method"); val.Type() == lua.LTString {
		req.Method = val.String()
	}

	// URL
	if val := L.GetField(t, "url"); val.Type() == lua.LTString {
		if u, err := url.Parse(val.String()); err == nil {
			req.URL = u
		}
	}

	// Body
	if val := L.GetField(t, "body"); val.Type() == lua.LTString {
		bodyStr := val.String()
		req.Body = io.NopCloser(bytes.NewBufferString(bodyStr))
		req.ContentLength = int64(len(bodyStr))
	}

	// Headers
	if val := L.GetField(t, "headers"); val.Type() == lua.LTTable {
		headersTable := val.(*lua.LTable)
		headersTable.ForEach(func(k lua.LValue, v lua.LValue) {
			if k.Type() == lua.LTString && v.Type() == lua.LTString {
				req.Header.Set(k.String(), v.String())
			}
		})
	}
}

// ResponseToLua converts an http.Response to a Lua table using the provided ScriptManager's state
func ResponseToLua(sm *ScriptManager, resp *http.Response, body string, responseTime int64) *lua.LTable {
	L := sm.GetState()
	t := L.NewTable()

	// Status
	L.SetField(t, "status", lua.LNumber(resp.StatusCode))

	// Body
	L.SetField(t, "body", lua.LString(body))

	// Time
	L.SetField(t, "time", lua.LNumber(responseTime))

	// Headers
	headersTable := L.NewTable()
	for k, v := range resp.Header {
		if len(v) > 0 {
			L.SetField(headersTable, k, lua.LString(v[0]))
		}
	}
	L.SetField(t, "headers", headersTable)

	return t
}
