// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package script

import (
	"net/url"

	lua "github.com/yuin/gopher-lua"
)

type UtilsAPI struct{}

func NewUtilsAPI() *UtilsAPI {
	return &UtilsAPI{}
}

func (api *UtilsAPI) Register(L *lua.LState) {
	utilsTable := L.NewTable()
	L.SetField(utilsTable, "url_encode", L.NewFunction(api.UrlEncode))
	L.SetField(utilsTable, "url_decode", L.NewFunction(api.UrlDecode))
	L.SetGlobal("utils", utilsTable)
}

func (api *UtilsAPI) UrlEncode(L *lua.LState) int {
	input := L.CheckString(1)
	L.Push(lua.LString(url.QueryEscape(input)))
	return 1
}

func (api *UtilsAPI) UrlDecode(L *lua.LState) int {
	input := L.CheckString(1)
	decoded, err := url.QueryUnescape(input)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(decoded))
	return 1
}
