// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package script

import (
	lua "github.com/yuin/gopher-lua"
	"gopkg.in/yaml.v3"
)

type YamlAPI struct{}

func NewYamlAPI() *YamlAPI {
	return &YamlAPI{}
}

func (api *YamlAPI) Register(L *lua.LState) {
	yamlTable := L.NewTable()
	L.SetField(yamlTable, "parse", L.NewFunction(api.Parse))
	L.SetField(yamlTable, "stringify", L.NewFunction(api.Stringify))
	L.SetGlobal("yaml", yamlTable)
}

func (api *YamlAPI) Parse(L *lua.LState) int {
	input := L.CheckString(1)
	var val interface{}
	err := yaml.Unmarshal([]byte(input), &val)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(ValueToLua(L, val))
	return 1
}

func (api *YamlAPI) Stringify(L *lua.LState) int {
	val := L.CheckAny(1)
	goVal := LuaToValue(val)
	data, err := yaml.Marshal(goVal)
	if err != nil {
		L.Error(lua.LString(err.Error()), 1)
		return 0
	}
	L.Push(lua.LString(string(data)))
	return 1
}
