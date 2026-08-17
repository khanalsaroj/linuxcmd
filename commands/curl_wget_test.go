package commands

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCurlGetsBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello from server"))
	}))
	defer srv.Close()

	code, out, errOut := run(t, "curl", srv.URL)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if out != "hello from server" {
		t.Errorf("curl output = %q, want %q", out, "hello from server")
	}
}

func TestCurlSavesToFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("file content"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	out := filepath.Join(dir, "out.txt")
	code, _, errOut := run(t, "curl", "-o", out, srv.URL)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "file content" {
		t.Errorf("saved content = %q, want %q", got, "file content")
	}
}

func TestCurlHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	code, _, _ := run(t, "curl", srv.URL)
	if code == 0 {
		t.Error("expected nonzero exit for a 404 response")
	}
}

func TestWgetDownloadsToFile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("wget content"))
	}))
	defer srv.Close()

	dir := t.TempDir()
	out := filepath.Join(dir, "page.html")
	code, _, errOut := runIn(t, dir, "wget", "-O", out, srv.URL)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, errOut)
	}
	if !strings.Contains(errOut, "saved") {
		t.Errorf("expected a 'saved' confirmation, got %q", errOut)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "wget content" {
		t.Errorf("downloaded content = %q, want %q", got, "wget content")
	}
}
