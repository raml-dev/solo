// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package importer

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"solo/internal/collection"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ImportResult carries the imported collection and any non-fatal warnings
// (e.g. skipped security schemes) to be surfaced to the user.
type ImportResult struct {
	Collection *collection.Collection
	Warnings   []string
}

// OpenAPIImporter implements the Importer interface for OpenAPI 3.x and Swagger 2.x
// documents in JSON or YAML format.
type OpenAPIImporter struct{}

func NewOpenAPIImporter() *OpenAPIImporter {
	return &OpenAPIImporter{}
}

// Import parses an OpenAPI 3.x or Swagger 2.x document and returns an ImportResult.
func (o *OpenAPIImporter) Import(path string) (ImportResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return ImportResult{}, fmt.Errorf("cannot read file: %w", err)
	}

	var doc unifiedAPIDocument
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".yaml" || ext == ".yml" {
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return ImportResult{}, fmt.Errorf("cannot parse YAML: %w", err)
		}
	} else {
		if err := json.Unmarshal(data, &doc); err != nil {
			return ImportResult{}, fmt.Errorf("cannot parse JSON: %w", err)
		}
	}

	// Detect format version
	var version string
	switch {
	case strings.HasPrefix(doc.OpenAPI, "3."):
		version = "3.x"
	case doc.Swagger == "2.0":
		version = "2.0"
	default:
		return ImportResult{}, fmt.Errorf("unsupported format: expected openapi 3.x or swagger 2.0 (got openapi=%q swagger=%q)", doc.OpenAPI, doc.Swagger)
	}

	baseURL := resolveBaseURL(doc, version)

	now := time.Now()
	coll := &collection.Collection{
		Id:                  generateUUID(),
		Name:                doc.Info.Title,
		Requests:            []collection.Request{},
		Folders:             []collection.Folder{},
		CreationTimestamp:   now,
		LastUpdateTimestamp: now,
	}
	if coll.Name == "" {
		coll.Name = "Imported API"
	}

	// Iterate paths in a deterministic method order
	methods := []string{"get", "post", "put", "patch", "delete", "head", "options"}
	for path, item := range doc.Paths {
		ops := map[string]*openAPIOperation{
			"get":     item.Get,
			"post":    item.Post,
			"put":     item.Put,
			"patch":   item.Patch,
			"delete":  item.Delete,
			"head":    item.Head,
			"options": item.Options,
		}
		for _, method := range methods {
			op := ops[method]
			if op == nil {
				continue
			}
			req := buildRequest(method, path, op, baseURL, version, doc)
			addOpenAPIRequest(coll, req, op.Tags)
		}
	}

	slog.Info("OpenAPI import completed",
		"version", version,
		"collection", coll.Name,
		"requests", len(coll.Requests))

	// Detect security schemes and build warnings
	var warnings []string
	schemeNames := collectSecuritySchemeNames(doc, version)
	if len(schemeNames) > 0 {
		msg := fmt.Sprintf(
			"Security schemes detected (%s) but not imported. "+
				"Configure authentication manually via request headers or environment variables.",
			strings.Join(schemeNames, ", "),
		)
		slog.Warn("OpenAPI import: security schemes skipped", "schemes", schemeNames)
		warnings = append(warnings, msg)
	}

	return ImportResult{Collection: coll, Warnings: warnings}, nil
}

// resolveBaseURL returns the base URL for requests depending on the format version.
func resolveBaseURL(doc unifiedAPIDocument, version string) string {
	if version == "3.x" {
		if len(doc.Servers) == 0 {
			return ""
		}
		return strings.TrimRight(doc.Servers[0].URL, "/")
	}

	// Swagger 2.x
	if doc.Host == "" {
		return ""
	}
	scheme := "https"
	if len(doc.Schemes) > 0 {
		scheme = doc.Schemes[0]
	}
	return scheme + "://" + doc.Host + strings.TrimRight(doc.BasePath, "/")
}

// buildRequest constructs a collection.Request from a single OpenAPI operation.
func buildRequest(method, path string, op *openAPIOperation, baseURL, version string, doc unifiedAPIDocument) collection.Request {
	req := collection.Request{
		Id:                  generateUUID(),
		Verb:                strings.ToUpper(method),
		Url:                 baseURL + path,
		Headers:             make(map[string]string),
		Cookies:             make(map[string]string),
		CreationTimestamp:   time.Now(),
		LastUpdateTimestamp: time.Now(),
	}

	// Name: operationId > summary > "METHOD /path"
	switch {
	case op.OperationId != "":
		req.Name = op.OperationId
	case op.Summary != "":
		req.Name = op.Summary
	default:
		req.Name = strings.ToUpper(method) + " " + path
	}

	// Header parameters (both formats)
	for _, p := range op.Parameters {
		if p.In == "header" {
			req.Headers[p.Name] = ""
		}
	}

	// Body — format-specific
	if version == "3.x" {
		if op.RequestBody != nil {
			if _, ok := op.RequestBody.Content["application/json"]; ok {
				req.Body = "{}"
				req.BodyType = "json"
			}
		}
	} else {
		// Swagger 2.x: body is a parameter with in=="body"
		hasBodyParam := false
		for _, p := range op.Parameters {
			if p.In == "body" {
				hasBodyParam = true
				break
			}
		}
		if hasBodyParam {
			// consumes priority: operation-level > root-level > default JSON
			effectiveConsumes := op.Consumes
			if len(effectiveConsumes) == 0 {
				effectiveConsumes = doc.Consumes
			}
			isJSON := len(effectiveConsumes) == 0
			for _, ct := range effectiveConsumes {
				if strings.Contains(ct, "application/json") {
					isJSON = true
					break
				}
			}
			if isJSON {
				req.Body = "{}"
				req.BodyType = "json"
			}
		}
	}

	return req
}

// collectSecuritySchemeNames returns the names of all declared security schemes.
func collectSecuritySchemeNames(doc unifiedAPIDocument, version string) []string {
	var names []string
	if version == "3.x" {
		for name := range doc.Components.SecuritySchemes {
			names = append(names, name)
		}
	} else {
		for name := range doc.SecurityDefinitions {
			names = append(names, name)
		}
	}
	return names
}

// ── Internal models ───────────────────────────────────────────────────────────

// unifiedAPIDocument covers both OpenAPI 3.x and Swagger 2.x fields.
type unifiedAPIDocument struct {
	// Version discriminators
	OpenAPI string `json:"openapi" yaml:"openapi"` // OpenAPI 3.x: "3.x.y"
	Swagger string `json:"swagger" yaml:"swagger"` // Swagger 2.x: "2.0"

	Info openAPIInfo `json:"info" yaml:"info"`

	// OpenAPI 3.x base URL
	Servers []openAPIServer `json:"servers" yaml:"servers"`

	// OpenAPI 3.x security schemes (under components)
	Components struct {
		SecuritySchemes map[string]interface{} `json:"securitySchemes" yaml:"securitySchemes"`
	} `json:"components" yaml:"components"`

	// Swagger 2.x base URL components
	Host     string   `json:"host"     yaml:"host"`
	BasePath string   `json:"basePath" yaml:"basePath"`
	Schemes  []string `json:"schemes"  yaml:"schemes"`

	// Swagger 2.x global consumes (fallback for operations without their own)
	Consumes []string `json:"consumes" yaml:"consumes"`

	// Swagger 2.x security schemes
	SecurityDefinitions map[string]interface{} `json:"securityDefinitions" yaml:"securityDefinitions"`

	Paths map[string]openAPIPathItem `json:"paths" yaml:"paths"`
}

type openAPIInfo struct {
	Title string `json:"title" yaml:"title"`
}

type openAPIServer struct {
	URL string `json:"url" yaml:"url"`
}

type openAPIPathItem struct {
	Get     *openAPIOperation `json:"get"     yaml:"get"`
	Post    *openAPIOperation `json:"post"    yaml:"post"`
	Put     *openAPIOperation `json:"put"     yaml:"put"`
	Patch   *openAPIOperation `json:"patch"   yaml:"patch"`
	Delete  *openAPIOperation `json:"delete"  yaml:"delete"`
	Head    *openAPIOperation `json:"head"    yaml:"head"`
	Options *openAPIOperation `json:"options" yaml:"options"`
}

type openAPIOperation struct {
	OperationId string              `json:"operationId" yaml:"operationId"`
	Summary     string              `json:"summary"     yaml:"summary"`
	Tags        []string            `json:"tags"        yaml:"tags"`
	Parameters  []openAPIParameter  `json:"parameters"  yaml:"parameters"`
	RequestBody *openAPIRequestBody `json:"requestBody" yaml:"requestBody"` // OpenAPI 3.x only
	Consumes    []string            `json:"consumes"    yaml:"consumes"`    // Swagger 2.x only
}

type openAPIParameter struct {
	Name   string                 `json:"name"   yaml:"name"`
	In     string                 `json:"in"     yaml:"in"`
	Schema map[string]interface{} `json:"schema" yaml:"schema"`
}

type openAPIRequestBody struct {
	Content map[string]openAPIMediaType `json:"content" yaml:"content"`
}

type openAPIMediaType struct {
	Schema map[string]interface{} `json:"schema" yaml:"schema"`
}

func addOpenAPIRequest(coll *collection.Collection, req collection.Request, tags []string) {
	if len(tags) == 0 || strings.TrimSpace(tags[0]) == "" {
		coll.Requests = append(coll.Requests, req)
		return
	}

	tagName := strings.TrimSpace(tags[0])
	for i := range coll.Folders {
		if coll.Folders[i].Name == tagName {
			coll.Folders[i].Requests = append(coll.Folders[i].Requests, req)
			return
		}
	}

	folder := collection.NewFolder(tagName)
	folder.Requests = append(folder.Requests, req)
	coll.Folders = append(coll.Folders, folder)
}
