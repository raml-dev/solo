package troubleshooting

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

var ErrNoLogFiles = errors.New("no log files found")

// BuildLogsZip reads all regular files in logsDir and returns a ZIP archive with those files.
func BuildLogsZip(logsDir string) ([]byte, []string, error) {
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		return nil, nil, fmt.Errorf("read logs directory: %w", err)
	}

	files := make([]os.DirEntry, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		files = append(files, entry)
	}

	if len(files) == 0 {
		return nil, nil, ErrNoLogFiles
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	included := make([]string, 0, len(files))

	for _, entry := range files {
		name := entry.Name()
		fullPath := filepath.Join(logsDir, name)

		data, err := os.ReadFile(fullPath)
		if err != nil {
			_ = zw.Close()
			return nil, nil, fmt.Errorf("read log file %q: %w", name, err)
		}

		w, err := zw.Create(name)
		if err != nil {
			_ = zw.Close()
			return nil, nil, fmt.Errorf("zip entry %q: %w", name, err)
		}

		if _, err := w.Write(data); err != nil {
			_ = zw.Close()
			return nil, nil, fmt.Errorf("write zip entry %q: %w", name, err)
		}

		included = append(included, name)
	}

	if err := zw.Close(); err != nil {
		return nil, nil, fmt.Errorf("close zip writer: %w", err)
	}

	return buf.Bytes(), included, nil
}
