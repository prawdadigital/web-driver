package zip

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	dir := t.TempDir()

	files := map[string]string{
		"top.txt":            "top level contents",
		"prefs.js":           "user_pref(\"x\", true);",
		"sub/nested.txt":     "nested contents",
		"sub/deep/leaf.json": "{\"k\":\"v\"}",
	}
	for name, content := range files {
		full := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
			t.Fatalf("MkdirAll(%q): %v", full, err)
		}
		if err := os.WriteFile(full, []byte(content), 0644); err != nil {
			t.Fatalf("WriteFile(%q): %v", full, err)
		}
	}

	buf, err := New(dir)
	if err != nil {
		t.Fatalf("New(%q): %v", dir, err)
	}

	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("zip.NewReader: %v", err)
	}

	got := make(map[string]string)
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatalf("open %q: %v", f.Name, err)
		}
		b, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("read %q: %v", f.Name, err)
		}
		// zip entry names always use forward slashes.
		got[filepath.ToSlash(f.Name)] = string(b)
	}

	if len(got) != len(files) {
		t.Fatalf("zip contains %d entries %v, want %d %v", len(got), got, len(files), files)
	}
	for name, want := range files {
		if got[name] != want {
			t.Errorf("entry %q = %q, want %q", name, got[name], want)
		}
	}
}

func TestNewNotADirectory(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "afile")
	if err := os.WriteFile(file, []byte("data"), 0644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if _, err := New(file); err == nil {
		t.Fatal("New(file) = nil error, want error for non-directory")
	}
}

func TestNewMissingPath(t *testing.T) {
	if _, err := New(filepath.Join(t.TempDir(), "does-not-exist")); err == nil {
		t.Fatal("New(missing) = nil error, want error")
	}
}
