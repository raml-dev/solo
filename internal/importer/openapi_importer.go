// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package importer

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"solo/internal/collection"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// ImportResult carries the imported collection, metadata extracted from the
// source document, and any non-fatal warnings (e.g. skipped security schemes)
// to be surfaced to the user.
type ImportResult struct {
	Collection *collection.Collection
	BasePath   string
	Servers    []string
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

	now := time.Now()
	coll := &collection.Collection{
		Id:                  generateUUID(),
		Name:                doc.Info.Title,
		Requests:            []collection.Request{},
		Variables:           map[string]collection.ValueType{},
		Folders:             []collection.Folder{},
		CreationTimestamp:   now,
		LastUpdateTimestamp: now,
	}
	if coll.Name == "" {
		coll.Name = "Imported API"
	}
	if defaultBaseURL := resolveOpenAPIDefaultBaseURL(doc, version); defaultBaseURL != "" {
		coll.Variables["baseUrl"] = collection.ValueType{
			Value: defaultBaseURL,
			Type:  "text",
		}
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
			req := o.buildRequest(method, path, op, item.Parameters, version, doc)
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

	return ImportResult{
		Collection: coll,
		BasePath:   doc.BasePath,
		Servers:    collectServerURLs(doc.Servers),
		Warnings:   warnings,
	}, nil
}

// extractExample returns the first non-nil example value following priority:
// inline example > first named example > schema example > schema default > recursive properties example.
func (o *OpenAPIImporter) extractExample(example interface{}, examples map[string]openAPIExample, schema map[string]interface{}, doc unifiedAPIDocument, depth int, hint string) interface{} {
	if depth > 10 { // Prevent infinite recursion
		return nil
	}

	if example != nil {
		return example
	}
	for _, ex := range examples {
		if ex.Value != nil {
			return ex.Value
		}
	}

	if schema == nil {
		return nil
	}

	// Resolve $ref if present
	if ref, ok := schema["$ref"].(string); ok {
		resolved := o.resolveRef(ref, doc)
		if resolved != nil {
			return o.extractExample(nil, nil, resolved, doc, depth+1, hint)
		}
	}

	if ex, ok := schema["example"]; ok && ex != nil {
		return ex
	}
	if def, ok := schema["default"]; ok && def != nil {
		return def
	}

	// Recursive: if it's an object with properties, try to build an example from them
	if props, ok := schema["properties"].(map[string]interface{}); ok {
		objExample := make(map[string]interface{})
		for name, prop := range props {
			if propMap, ok := prop.(map[string]interface{}); ok {
				if val := o.extractExample(nil, nil, propMap, doc, depth+1, name); val != nil {
					objExample[name] = val
				}
			}
		}
		if len(objExample) > 0 {
			return objExample
		}
	}

	// Recursive: if it's an array, try to build an example from items
	if items, ok := schema["items"].(map[string]interface{}); ok {
		if val := o.extractExample(nil, nil, items, doc, depth+1, hint); val != nil {
			return []interface{}{val}
		}
	}

	// Fallback for primitive types to ensure something is generated
	if typeVal, ok := schema["type"].(string); ok {
		// Check for binary format first
		if format, ok := schema["format"].(string); ok && (format == "binary" || format == "base64") {
			return "[BINARY_FILE_CONTENT]"
		}

		switch typeVal {
		case "string":
			if hint != "" {
				return "{{" + hint + "}}"
			}
			return "string"
		case "integer", "number":
			return 0
		case "boolean":
			return false
		}
	}

	return nil
}

func (o *OpenAPIImporter) resolveRequestBodyRef(ref string, doc unifiedAPIDocument) *openAPIRequestBody {
	if !strings.HasPrefix(ref, "#/") {
		return nil
	}
	parts := strings.Split(ref, "/")
	if len(parts) >= 4 && parts[1] == "components" && parts[2] == "requestBodies" && doc.Components.RequestBodies != nil {
		if raw, ok := doc.Components.RequestBodies[parts[3]]; ok {
			// Since it's interface{}, we need to handle potential map conversion
			// Or we can just unmarshal it again into the struct for safety
			b, _ := json.Marshal(raw)
			var rb openAPIRequestBody
			if err := json.Unmarshal(b, &rb); err == nil {
				return &rb
			}
		}
	}
	return nil
}

func (o *OpenAPIImporter) resolveRef(ref string, doc unifiedAPIDocument) map[string]interface{} {
	if !strings.HasPrefix(ref, "#/") {
		return nil
	}
	parts := strings.Split(ref, "/")
	if len(parts) < 3 {
		return nil
	}

	// Handle Swagger 2.0 #/definitions/...
	if parts[1] == "definitions" && doc.Definitions != nil {
		if schema, ok := doc.Definitions[parts[2]]; ok {
			return schema
		}
	}

	// Handle OpenAPI 3.x #/components/schemas/...
	if len(parts) >= 4 && parts[1] == "components" && parts[2] == "schemas" && doc.Components.Schemas != nil {
		if schema, ok := doc.Components.Schemas[parts[3]].(map[string]interface{}); ok {
			return schema
		}
	}

	return nil
}

// exampleToJSONString serialises a value to a JSON string.
// Returns "{}" when the value is nil or serialisation fails.
func exampleToJSONString(v interface{}) string {
	if v == nil {
		return "{}"
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "{}"
	}
	return string(b)
}

// buildRequest constructs a collection.Request from a single OpenAPI operation.
func (o *OpenAPIImporter) buildRequest(method, path string, op *openAPIOperation, pathParams []openAPIParameter, version string, doc unifiedAPIDocument) collection.Request {
	normalizedPath := normalizeOpenAPIPathPlaceholders(path)
	req := collection.Request{
		Id:                  generateUUID(),
		Verb:                strings.ToUpper(method),
		Url:                 buildOpenAPIRequestURL(normalizedPath),
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

	// Merge parameters: operation params override path params
	allParams := make([]openAPIParameter, 0, len(pathParams)+len(op.Parameters))
	paramMap := make(map[string]int) // key: "in:name", value: index in allParams

	for _, p := range pathParams {
		key := p.In + ":" + p.Name
		paramMap[key] = len(allParams)
		allParams = append(allParams, p)
	}
	for _, p := range op.Parameters {
		key := p.In + ":" + p.Name
		if idx, ok := paramMap[key]; ok {
			allParams[idx] = p
		} else {
			allParams = append(allParams, p)
		}
	}

	var queryParts []string

	for _, p := range allParams {
		switch p.In {
		case "header":
			v := o.extractExample(p.Example, p.Examples, p.Schema, doc, 0, p.Name)
			if v != nil {
				req.Headers[p.Name] = fmt.Sprintf("%v", v)
			} else {
				req.Headers[p.Name] = ""
			}
		case "query":
			v := o.extractExample(p.Example, p.Examples, p.Schema, doc, 0, p.Name)
			if v != nil {
				queryParts = append(queryParts, fmt.Sprintf("%s=%v", p.Name, v))
			} else {
				queryParts = append(queryParts, p.Name+"=")
			}
		}
	}

	if len(queryParts) > 0 {
		req.Url += "?" + strings.Join(queryParts, "&")
	}

	// Body — format-specific
	if version == "3.x" {
		if op.RequestBody != nil {
			rb := op.RequestBody
			if rb.Ref != "" {
				if resolved := o.resolveRequestBodyRef(rb.Ref, doc); resolved != nil {
					rb = resolved
				}
			}

			if media, ok := rb.Content["application/json"]; ok {
				req.Body = exampleToJSONString(o.extractExample(media.Example, media.Examples, media.Schema, doc, 0, ""))
				req.BodyType = "json"
			} else if media, ok := rb.Content["application/x-www-form-urlencoded"]; ok {
				obj := o.extractExample(media.Example, media.Examples, media.Schema, doc, 0, "")
				req.Body = objectToQueryString(obj)
				req.BodyType = "text"
				req.Headers["Content-Type"] = "application/x-www-form-urlencoded"
			} else if media, ok := rb.Content["multipart/form-data"]; ok {
				obj := o.extractExample(media.Example, media.Examples, media.Schema, doc, 0, "")
				boundary := "solo-boundary"
				req.Body = objectToMultipartString(obj, boundary)
				req.BodyType = "text"
				req.Headers["Content-Type"] = "multipart/form-data; boundary=" + boundary
			}
		}
	} else {
		// Swagger 2.x: body is a parameter with in=="body" or in=="formData"
		effectiveConsumes := op.Consumes
		if len(effectiveConsumes) == 0 {
			effectiveConsumes = doc.Consumes
		}

		formDataMap := make(map[string]interface{})
		hasFormData := false

		for _, p := range allParams {
			if p.In == "body" {
				isJSON := len(effectiveConsumes) == 0
				for _, ct := range effectiveConsumes {
					if strings.Contains(ct, "application/json") {
						isJSON = true
						break
					}
				}
				if isJSON {
					req.Body = exampleToJSONString(o.extractExample(p.Example, p.Examples, p.Schema, doc, 0, p.Name))
					req.BodyType = "json"
				}
				break
			} else if p.In == "formData" {
				hasFormData = true
				formDataMap[p.Name] = o.extractExample(p.Example, p.Examples, p.Schema, doc, 0, p.Name)
			}
		}

		if hasFormData {
			isMultipart := false
			for _, ct := range effectiveConsumes {
				if strings.Contains(ct, "multipart/form-data") {
					isMultipart = true
					break
				}
			}

			if isMultipart {
				boundary := "solo-boundary"
				req.Body = objectToMultipartString(formDataMap, boundary)
				req.BodyType = "text"
				req.Headers["Content-Type"] = "multipart/form-data; boundary=" + boundary
			} else {
				req.Body = objectToQueryString(formDataMap)
				req.BodyType = "text"
				req.Headers["Content-Type"] = "application/x-www-form-urlencoded"
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
		Schemas         map[string]interface{} `json:"schemas"         yaml:"schemas"`
		RequestBodies   map[string]interface{} `json:"requestBodies"   yaml:"requestBodies"`
	} `json:"components" yaml:"components"`

	// Swagger 2.x definitions
	Definitions map[string]map[string]interface{} `json:"definitions" yaml:"definitions"`

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
	Get        *openAPIOperation  `json:"get"        yaml:"get"`
	Post       *openAPIOperation  `json:"post"       yaml:"post"`
	Put        *openAPIOperation  `json:"put"        yaml:"put"`
	Patch      *openAPIOperation  `json:"patch"      yaml:"patch"`
	Delete     *openAPIOperation  `json:"delete"     yaml:"delete"`
	Head       *openAPIOperation  `json:"head"       yaml:"head"`
	Options    *openAPIOperation  `json:"options"    yaml:"options"`
	Parameters []openAPIParameter `json:"parameters" yaml:"parameters"`
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
	Name     string                    `json:"name"     yaml:"name"`
	In       string                    `json:"in"       yaml:"in"`
	Schema   map[string]interface{}    `json:"schema"   yaml:"schema"`
	Example  interface{}               `json:"example"  yaml:"example"`
	Examples map[string]openAPIExample `json:"examples" yaml:"examples"`
}

type openAPIExample struct {
	Value interface{} `json:"value" yaml:"value"`
}

type openAPIRequestBody struct {
	Ref     string                      `json:"$ref"    yaml:"$ref"`
	Content map[string]openAPIMediaType `json:"content" yaml:"content"`
}

type openAPIMediaType struct {
	Schema   map[string]interface{}    `json:"schema"   yaml:"schema"`
	Example  interface{}               `json:"example"  yaml:"example"`
	Examples map[string]openAPIExample `json:"examples" yaml:"examples"`
}

var openAPIPathParamPattern = regexp.MustCompile(`\{([A-Za-z0-9_.-]+)\}`)

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

func normalizeOpenAPIPathPlaceholders(path string) string {
	return openAPIPathParamPattern.ReplaceAllString(path, "{{$1}}")
}

func buildOpenAPIRequestURL(path string) string {
	if strings.HasPrefix(path, "/") {
		return "{{baseUrl}}" + path
	}
	return "{{baseUrl}}/" + path
}

func collectServerURLs(servers []openAPIServer) []string {
	if len(servers) == 0 {
		return nil
	}

	urls := make([]string, 0, len(servers))
	for _, server := range servers {
		if strings.TrimSpace(server.URL) == "" {
			continue
		}
		urls = append(urls, strings.TrimRight(server.URL, "/"))
	}
	if len(urls) == 0 {
		return nil
	}
	return urls
}

func resolveOpenAPIDefaultBaseURL(doc unifiedAPIDocument, version string) string {
	basePath := strings.TrimSpace(doc.BasePath)
	baseURL := ""

	if version == "3.x" {
		servers := collectServerURLs(doc.Servers)
		if len(servers) > 0 {
			baseURL = servers[0]
		}
	} else {
		host := strings.TrimSpace(doc.Host)

		if host != "" {
			scheme := "https"

			if len(doc.Schemes) > 0 && strings.TrimSpace(doc.Schemes[0]) != "" {
				scheme = strings.TrimSpace(doc.Schemes[0])
			}

			baseURL = strings.TrimRight(fmt.Sprintf("%s://%s", scheme, strings.TrimRight(host, "/")), "/")
		}
	}

	if basePath == "" || basePath == "/" {
		return baseURL
	}
	if !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	return strings.TrimRight(baseURL+strings.TrimRight(basePath, "/"), "/")
}
