// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only
package appinfo

import "encoding/json"

type AppInfo struct {
	CompanyName    string `json:"companyName"`
	ProductName    string `json:"productName"`
	ProductVersion string `json:"productVersion"`
	License        string `json:"license"`
	DocsLink       string `json:"docsLink"`
	GHLink         string `json:"ghLink"`
	OrgLink        string `json:"orgLink"`
}

func GetAppInfo(wailsJSON []byte) AppInfo {
	var cfg struct {
		Info AppInfo `json:"info"`
	}
	json.Unmarshal(wailsJSON, &cfg)
	return cfg.Info
}
