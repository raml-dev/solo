// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package script

import (
	"encoding/xml"
	"io"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

type XmlAPI struct{}

func NewXmlAPI() *XmlAPI {
	return &XmlAPI{}
}

// Register initializes the 'xml' global table and its functions
func (api *XmlAPI) Register(L *lua.LState) {
	xmlTable := L.NewTable()
	L.SetField(xmlTable, "parse", L.NewFunction(api.Parse))
	L.SetGlobal("xml", xmlTable)
}

// Node represents a simplified XML node structure for Lua
type Node struct {
	XMLName  xml.Name
	Attrs    []xml.Attr `xml:",any,attr"`
	Content  string     `xml:",chardata"`
	Children []Node     `xml:",any"`
}

// Parse implements xml.parse(string) -> table
func (api *XmlAPI) Parse(L *lua.LState) int {
	input := L.CheckString(1)

	decoder := xml.NewDecoder(strings.NewReader(input))
	var root Node
	err := decoder.Decode(&root)
	if err != nil && err != io.EOF {
		L.Push(lua.LNil)
		L.Push(lua.LString(err.Error()))
		return 2
	}

	L.Push(api.nodeToLua(L, root))
	return 1
}

// nodeToLua converts an XML node to a Lua table
func (api *XmlAPI) nodeToLua(L *lua.LState, node Node) lua.LValue {
	t := L.NewTable()
	L.SetField(t, "name", lua.LString(node.XMLName.Local))

	// Attributes
	if len(node.Attrs) > 0 {
		attrTable := L.NewTable()
		for _, attr := range node.Attrs {
			L.SetField(attrTable, attr.Name.Local, lua.LString(attr.Value))
		}
		L.SetField(t, "attrs", attrTable)
	}

	// Content
	content := strings.TrimSpace(node.Content)
	if content != "" {
		L.SetField(t, "content", lua.LString(content))
	}

	// Children
	if len(node.Children) > 0 {
		childTable := L.NewTable()
		for _, child := range node.Children {
			childTable.Append(api.nodeToLua(L, child))
		}
		L.SetField(t, "children", childTable)

		// Also provide direct access by tag name for convenience
		for _, child := range node.Children {
			L.SetField(t, child.XMLName.Local, api.nodeToLua(L, child))
		}
	}

	return t
}
