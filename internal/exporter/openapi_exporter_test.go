// Copyright 2026-present raml-dev
// SPDX-License-Identifier: AGPL-3.0-only

package exporter

import (
	"bytes"
	"solo/internal/collection"
	"testing"
)

func TestExportOpenAPI_ExportsBearerSchemeWithoutCredential(t *testing.T) {
	coll := collection.NewCollection("Bearer API")
	coll.Requests = append(coll.Requests, collection.Request{
		Name: "secured",
		Url:  "https://api.example.com/secured",
		Verb: "GET",
		Auth: &collection.AuthConfiguration{
			Mode:        collection.AuthModeBearer,
			BearerToken: "must-not-be-exported",
		},
	})

	document, err := ExportOpenAPI(&coll)
	if err != nil {
		t.Fatalf("ExportOpenAPI() error = %v", err)
	}
	for _, expected := range [][]byte{
		[]byte("bearerAuth:"),
		[]byte("scheme: bearer"),
		[]byte("security:"),
	} {
		if !bytes.Contains(document, expected) {
			t.Fatalf("OpenAPI document does not contain %q:\n%s", expected, document)
		}
	}
	if bytes.Contains(document, []byte("must-not-be-exported")) {
		t.Fatal("OpenAPI document contains the bearer credential")
	}
}
