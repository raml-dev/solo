// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package script

import (
	"time"

	lua "github.com/yuin/gopher-lua"
)

type TimeAPI struct{}

func NewTimeAPI() *TimeAPI {
	return &TimeAPI{}
}

func (api *TimeAPI) Register(L *lua.LState) {
	timeTable := L.NewTable()
	L.SetField(timeTable, "now", L.NewFunction(api.Now))
	L.SetField(timeTable, "ms", L.NewFunction(api.Ms))
	L.SetField(timeTable, "format", L.NewFunction(api.Format))
	L.SetGlobal("time", timeTable)
}

func (api *TimeAPI) Now(L *lua.LState) int {
	L.Push(lua.LNumber(time.Now().Unix()))
	return 1
}

func (api *TimeAPI) Ms(L *lua.LState) int {
	L.Push(lua.LNumber(time.Now().UnixMilli()))
	return 1
}

func (api *TimeAPI) Format(L *lua.LState) int {
	timestamp := int64(L.CheckNumber(1))
	layout := L.CheckString(2)
	t := time.Unix(timestamp, 0)
	L.Push(lua.LString(t.Format(layout)))
	return 1
}
