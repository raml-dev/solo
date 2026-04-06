// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only

package script

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

type CryptoAPI struct{}

func NewCryptoAPI() *CryptoAPI {
	return &CryptoAPI{}
}

func (api *CryptoAPI) Register(L *lua.LState) {
	cryptoTable := L.NewTable()
	L.SetField(cryptoTable, "md5", L.NewFunction(api.Md5))
	L.SetField(cryptoTable, "sha1", L.NewFunction(api.Sha1))
	L.SetField(cryptoTable, "sha256", L.NewFunction(api.Sha256))
	L.SetField(cryptoTable, "base64_encode", L.NewFunction(api.Base64Encode))
	L.SetField(cryptoTable, "base64_decode", L.NewFunction(api.Base64Decode))
	L.SetGlobal("crypto", cryptoTable)
}

func (api *CryptoAPI) Md5(L *lua.LState) int {
	input := L.CheckString(1)
	hash := md5.Sum([]byte(input))
	L.Push(lua.LString(fmt.Sprintf("%x", hash)))
	return 1
}

func (api *CryptoAPI) Sha1(L *lua.LState) int {
	input := L.CheckString(1)
	hash := sha1.Sum([]byte(input))
	L.Push(lua.LString(fmt.Sprintf("%x", hash)))
	return 1
}

func (api *CryptoAPI) Sha256(L *lua.LState) int {
	input := L.CheckString(1)
	hash := sha256.Sum256([]byte(input))
	L.Push(lua.LString(fmt.Sprintf("%x", hash)))
	return 1
}

func (api *CryptoAPI) Base64Encode(L *lua.LState) int {
	input := L.CheckString(1)
	encoded := base64.StdEncoding.EncodeToString([]byte(input))
	L.Push(lua.LString(encoded))
	return 1
}

func (api *CryptoAPI) Base64Decode(L *lua.LState) int {
	input := L.CheckString(1)
	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}
	L.Push(lua.LString(string(decoded)))
	return 1
}
