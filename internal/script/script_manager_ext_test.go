// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only

package script

import (
	"net/http"
	"testing"
)

func TestScriptManager_CryptoAPI(t *testing.T) {
	sm := NewScriptManager(nil)
	req, _ := http.NewRequest("GET", "http://dummy", nil)

	script := `
        local data = "hello world"
        local m = crypto.md5(data)
        if m ~= "5eb63bbbe01eeed093cb22bb8f5acdc3" then error("md5 mismatch") end

        local s1 = crypto.sha1(data)
        if s1 ~= "2aae6c35c94fcfb415dbe95f408b9ce91ee846ed" then error("sha1 mismatch") end

        local s256 = crypto.sha256(data)
        if s256 ~= "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9" then error("sha256 mismatch") end

        local b64 = crypto.base64_encode(data)
        if b64 ~= "aGVsbG8gd29ybGQ=" then error("base64_encode mismatch") end

        local decoded = crypto.base64_decode(b64)
        if decoded ~= data then error("base64_decode mismatch") end

        local _, err = crypto.base64_decode("invalid!!!")
        if err == nil then error("base64_decode should fail for invalid input") end
    `

	if err := sm.ExecutePreRequest(script, req); err != nil {
		t.Fatalf("CryptoAPI test failed: %v", err)
	}
}

func TestScriptManager_UuidAPI(t *testing.T) {
	sm := NewScriptManager(nil)
	req, _ := http.NewRequest("GET", "http://dummy", nil)

	script := `
        local id = uuid.v4()
        if string.len(id) ~= 36 then error("uuid length mismatch") end
        env.id = id
    `

	if err := sm.ExecutePreRequest(script, req); err != nil {
		t.Fatalf("UuidAPI test failed: %v", err)
	}

	vars := sm.GetSessionVars()
	if len(vars["id"]) != 36 {
		t.Errorf("id length = %v, want 36", len(vars["id"]))
	}
}

func TestScriptManager_TimeAPI(t *testing.T) {
	sm := NewScriptManager(nil)
	req, _ := http.NewRequest("GET", "http://dummy", nil)

	script := `
        local n = time.now()
        if n <= 0 then error("time.now invalid") end

        local m = time.ms()
        if m <= 0 then error("time.ms invalid") end
        if m < n * 1000 then error("time.ms should be >= time.now * 1000") end

        local f = time.format(1712160000, "2006-01-02")
        if f ~= "2024-04-03" then error("time.format mismatch: got " .. f) end
    `

	if err := sm.ExecutePreRequest(script, req); err != nil {
		t.Fatalf("TimeAPI test failed: %v", err)
	}
}

func TestScriptManager_YamlAPI(t *testing.T) {
	sm := NewScriptManager(nil)
	req, _ := http.NewRequest("GET", "http://dummy", nil)

	script := `
        local yaml_str = "id: 123\nname: test\ntags:\n  - a\n  - b"
        local data = yaml.parse(yaml_str)
        if data.id ~= 123 then error("yaml id mismatch") end
        if data.name ~= "test" then error("yaml name mismatch") end
        if data.tags[1] ~= "a" then error("yaml tags[1] mismatch") end

        local back = yaml.stringify(data)
        if not string.find(back, "name: test", 1, true) then error("yaml stringify mismatch") end
    `

	if err := sm.ExecutePreRequest(script, req); err != nil {
		t.Fatalf("YamlAPI test failed: %v", err)
	}
}

func TestScriptManager_UtilsAPI(t *testing.T) {
	sm := NewScriptManager(nil)
	req, _ := http.NewRequest("GET", "http://dummy", nil)

	script := `
        local raw = "hello world!"
        local enc = utils.url_encode(raw)
        if enc ~= "hello+world%21" then error("url_encode mismatch: got " .. enc) end

        local dec = utils.url_decode(enc)
        if dec ~= raw then error("url_decode mismatch") end

        local _, err = utils.url_decode("%invalid")
        if err == nil then error("url_decode should fail for invalid input") end
    `

	if err := sm.ExecutePreRequest(script, req); err != nil {
		t.Fatalf("UtilsAPI test failed: %v", err)
	}
}
