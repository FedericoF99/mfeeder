package cmd

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	previous := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	os.Stdout = w
	closedWriter := false
	closedReader := false
	defer func() {
		os.Stdout = previous
		if !closedWriter {
			_ = w.Close()
		}
		if !closedReader {
			_ = r.Close()
		}
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	closedWriter = true
	os.Stdout = previous

	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}
	closedReader = true

	return string(out)
}

func withConfigFile(t *testing.T, content string) {
	t.Helper()

	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.WriteFile(filepath.Join(dir, "mfeeder.conf"), []byte(content), 0644); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	t.Setenv("MFEEDER_DATA_DIR", dir)
}
