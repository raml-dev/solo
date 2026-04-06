// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only

package script

import (
	"github.com/google/uuid"
	lua "github.com/yuin/gopher-lua"
)

type UuidAPI struct{}

func NewUuidAPI() *UuidAPI {
	return &UuidAPI{}
}

func (api *UuidAPI) Register(L *lua.LState) {
	uuidTable := L.NewTable()
	L.SetField(uuidTable, "v4", L.NewFunction(api.V4))
	L.SetGlobal("uuid", uuidTable)
}

func (api *UuidAPI) V4(L *lua.LState) int {
	L.Push(lua.LString(uuid.NewString()))
	return 1
}
