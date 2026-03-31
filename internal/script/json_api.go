// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only

package script

import (
	"encoding/json"

	"github.com/tidwall/gjson"
	lua "github.com/yuin/gopher-lua"
)

type JsonAPI struct{}

func NewJsonAPI() *JsonAPI {
	return &JsonAPI{}
}

// Register initializes the 'json' global table and its functions
func (api *JsonAPI) Register(L *lua.LState) {
	jsonTable := L.NewTable()
	L.SetField(jsonTable, "parse", L.NewFunction(api.Parse))
	L.SetField(jsonTable, "stringify", L.NewFunction(api.Stringify))
	L.SetGlobal("json", jsonTable)
}

// Parse implements json.parse(string) -> table
func (api *JsonAPI) Parse(L *lua.LState) int {
	input := L.CheckString(1)
	res := gjson.Parse(input)
	L.Push(ValueToLua(L, res.Value()))
	return 1
}

// Stringify implements json.stringify(table) -> string
func (api *JsonAPI) Stringify(L *lua.LState) int {
	val := L.CheckAny(1)
	goVal := LuaToValue(val)
	data, err := json.Marshal(goVal)
	if err != nil {
		L.Error(lua.LString(err.Error()), 1)
		return 0
	}
	L.Push(lua.LString(string(data)))
	return 1
}

// ValueToLua converts generic Go values (maps, slices, primitives) to Lua values
func ValueToLua(L *lua.LState, val interface{}) lua.LValue {
	switch v := val.(type) {
	case string:
		return lua.LString(v)
	case float64:
		return lua.LNumber(v)
	case float32:
		return lua.LNumber(v)
	case int:
		return lua.LNumber(v)
	case int32:
		return lua.LNumber(v)
	case int64:
		return lua.LNumber(v)
	case bool:
		return lua.LBool(v)
	case nil:
		return lua.LNil
	case map[string]interface{}:
		table := L.NewTable()
		for k, v := range v {
			L.SetField(table, k, ValueToLua(L, v))
		}
		return table
	case []interface{}:
		table := L.NewTable()
		for _, v := range v {
			table.Append(ValueToLua(L, v))
		}
		return table
	default:
		return lua.LNil
	}
}

// LuaToValue converts Lua values back to generic Go values
func LuaToValue(lval lua.LValue) interface{} {
	switch v := lval.(type) {
	case lua.LString:
		return string(v)
	case lua.LNumber:
		return float64(v)
	case lua.LBool:
		return bool(v)
	case *lua.LTable:
		// Improved heuristic: if it has any string keys, it's a map.
		// If it only has contiguous integer keys starting from 1, it's an array.
		isArr := true
		maxn := v.Len()
		
		// If it's empty, we default to map (common in Lua->JSON)
		if maxn == 0 {
			// Check if it's REALLY empty or if it has string keys
			hasStringKeys := false
			v.ForEach(func(k, _ lua.LValue) {
				if _, ok := k.(lua.LString); ok {
					hasStringKeys = true
				}
			})
			if !hasStringKeys {
				// It's an empty table, could be [] or {}. Default to {} for now.
				return make(map[string]interface{})
			}
			isArr = false
		} else {
			// It has an array part, but does it also have string keys?
			v.ForEach(func(k, _ lua.LValue) {
				if _, ok := k.(lua.LString); ok {
					isArr = false
				}
			})
		}

		if isArr {
			arr := make([]interface{}, 0, maxn)
			for i := 1; i <= maxn; i++ {
				arr = append(arr, LuaToValue(v.RawGetInt(i)))
			}
			return arr
		}

		// Otherwise treat as map
		m := make(map[string]interface{})
		v.ForEach(func(k, val lua.LValue) {
			m[k.String()] = LuaToValue(val)
		})
		return m
	default:
		return nil
	}
}
