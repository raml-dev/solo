// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package exporter

import (
	"encoding/json"
	"regexp"
	"solo/internal/collection"
	"strings"

	"gopkg.in/yaml.v3"
)

var soloPlaceholderPattern = regexp.MustCompile(`\{\{([A-Za-z0-9_.\-]+)\}\}`)

// ExportOpenAPI converts a Solo collection to an OpenAPI 3.1 YAML document.
func ExportOpenAPI(coll *collection.Collection) ([]byte, error) {
	doc := buildOpenAPIDocument(coll)
	return yaml.Marshal(doc)
}

func buildOpenAPIDocument(coll *collection.Collection) map[string]interface{} {
	baseURL := "/"
	if v, ok := coll.Variables["baseUrl"]; ok && strings.TrimSpace(v.Value) != "" {
		baseURL = strings.TrimRight(strings.TrimSpace(v.Value), "/")
	}

	paths := map[string]interface{}{}
	var tags []map[string]string

	for _, req := range coll.Requests {
		addOpenAPIPath(paths, req, "", baseURL)
	}

	for _, folder := range coll.Folders {
		collectFolderPaths(paths, &tags, folder, baseURL)
	}

	doc := map[string]interface{}{
		"openapi": "3.1.0",
		"info": map[string]interface{}{
			"title":   coll.Name,
			"version": "1.0.0",
		},
		"servers": []interface{}{
			map[string]interface{}{"url": baseURL},
		},
		"paths": paths,
	}
	if len(tags) > 0 {
		doc["tags"] = tags
	}
	return doc
}

func collectFolderPaths(paths map[string]interface{}, tags *[]map[string]string, folder collection.Folder, baseURL string) {
	*tags = append(*tags, map[string]string{"name": folder.Name})
	for _, req := range folder.Requests {
		addOpenAPIPath(paths, req, folder.Name, baseURL)
	}
	for _, sub := range folder.Folders {
		collectFolderPaths(paths, tags, sub, baseURL)
	}
}

func addOpenAPIPath(paths map[string]interface{}, req collection.Request, tag string, baseURL string) {
	path := toOpenAPIPath(req.Url, baseURL)
	method := strings.ToLower(req.Verb)

	pathItem, ok := paths[path].(map[string]interface{})
	if !ok {
		pathItem = map[string]interface{}{}
		paths[path] = pathItem
	}

	operation := map[string]interface{}{
		"operationId": req.Name,
	}
	if tag != "" {
		operation["tags"] = []string{tag}
	}

	var params []interface{}
	for name, value := range req.Headers {
		p := map[string]interface{}{
			"name":     name,
			"in":       "header",
			"required": false,
		}
		if value != "" {
			p["example"] = value
		}
		params = append(params, p)
	}
	if len(params) > 0 {
		operation["parameters"] = params
	}

	if req.Body != "" && strings.ToLower(req.BodyType) == "json" {
		var parsed interface{}
		if err := json.Unmarshal([]byte(req.Body), &parsed); err == nil {
			operation["requestBody"] = map[string]interface{}{
				"content": map[string]interface{}{
					"application/json": map[string]interface{}{
						"example": parsed,
					},
				},
			}
		}
	}

	pathItem[method] = operation
}

// toOpenAPIPath strips the baseURL prefix and converts {{var}} placeholders to {var}.
func toOpenAPIPath(rawURL, baseURL string) string {
	path := rawURL
	if baseURL != "/" && baseURL != "" {
		path = strings.TrimPrefix(path, baseURL)
	}
	if path == "" {
		path = "/"
	}
	return soloPlaceholderPattern.ReplaceAllString(path, "{$1}")
}
