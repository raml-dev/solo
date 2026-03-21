// Copyright 2026-present raml-dev
// SPDX-License-Identifier: GPL-3.0-only

package troubleshooting

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestBuildLogsZip_IncludesAllRegularFiles(t *testing.T) {
	logsDir := t.TempDir()

	files := map[string]string{
		"solo.log":      "main-log-content",
		"solo.log.1":    "rotated-log-content",
		"solo.log.2.gz": "compressed-by-lumberjack",
	}

	for name, content := range files {
		if err := os.WriteFile(filepath.Join(logsDir, name), []byte(content), 0o600); err != nil {
			t.Fatalf("failed to create fixture %s: %v", name, err)
		}
	}

	zipBytes, included, err := BuildLogsZip(logsDir)
	if err != nil {
		t.Fatalf("BuildLogsZip failed: %v", err)
	}

	if len(included) != len(files) {
		t.Fatalf("expected %d included files, got %d", len(files), len(included))
	}

	zr, err := zip.NewReader(bytes.NewReader(zipBytes), int64(len(zipBytes)))
	if err != nil {
		t.Fatalf("invalid zip archive: %v", err)
	}

	if len(zr.File) != len(files) {
		t.Fatalf("expected %d zip entries, got %d", len(files), len(zr.File))
	}

	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("failed opening zip entry %s: %v", f.Name, err)
		}
		payload, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("failed reading zip entry %s: %v", f.Name, err)
		}

		expected, ok := files[f.Name]
		if !ok {
			t.Fatalf("unexpected file in zip: %s", f.Name)
		}
		if string(payload) != expected {
			t.Fatalf("unexpected payload for %s: got %q want %q", f.Name, string(payload), expected)
		}
	}
}

func TestBuildLogsZip_EmptyDirectory(t *testing.T) {
	logsDir := t.TempDir()

	_, _, err := BuildLogsZip(logsDir)
	if !errors.Is(err, ErrNoLogFiles) {
		t.Fatalf("expected ErrNoLogFiles, got %v", err)
	}
}
