package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/serizawa/pageshelf/internal/store"
)

func newTestSession(t *testing.T) (*store.Store, string) {
	t.Helper()
	st, err := store.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m, _, err := st.Create("test", "test", false)
	if err != nil {
		t.Fatal(err)
	}
	return st, m.ID
}

func TestPreparePutContentRendersMarkdownByDefault(t *testing.T) {
	name, data, err := preparePutContent(strings.NewReader("# Demo\n\n- one\n- two\n"), "README.md", false)
	if err != nil {
		t.Fatal(err)
	}
	if name != "README.html" {
		t.Fatalf("name = %q, want README.html", name)
	}
	body := string(data)
	if !strings.Contains(body, "<title>Demo</title>") || !strings.Contains(body, "<li>one</li>") {
		t.Fatalf("rendered HTML missing expected content:\n%s", body)
	}
}

func TestPreparePutContentRawMarkdown(t *testing.T) {
	name, data, err := preparePutContent(strings.NewReader("# Demo\n"), "README.md", true)
	if err != nil {
		t.Fatal(err)
	}
	if name != "README.md" {
		t.Fatalf("name = %q, want README.md", name)
	}
	if string(data) != "# Demo\n" {
		t.Fatalf("data = %q", string(data))
	}
}

func TestPutPathRendersMarkdownAndReturnsHTMLPath(t *testing.T) {
	st, sid := newTestSession(t)
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "report.md")
	if err := os.WriteFile(mdPath, []byte("# Report\n\nhello"), 0600); err != nil {
		t.Fatal(err)
	}
	added, err := putPath(st, sid, mdPath, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(added) != 1 || added[0] != "report.html" {
		t.Fatalf("added = %#v", added)
	}
	m, err := st.Load(sid)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Files) != 1 || m.Files[0].Path != "report.html" || m.Files[0].MIME != "text/html; charset=utf-8" {
		t.Fatalf("manifest files = %#v", m.Files)
	}
	f, _, _, err := st.Open(sid, "report.html")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	b, err := os.ReadFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "<h1 id=\"report\">Report</h1>") {
		t.Fatalf("rendered file missing heading:\n%s", string(b))
	}
}

func TestPutPathRawMarkdownKeepsMDPath(t *testing.T) {
	st, sid := newTestSession(t)
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "report.md")
	if err := os.WriteFile(mdPath, []byte("# Report\n"), 0600); err != nil {
		t.Fatal(err)
	}
	added, err := putPath(st, sid, mdPath, false, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(added) != 1 || added[0] != "report.md" {
		t.Fatalf("added = %#v", added)
	}
	m, err := st.Load(sid)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Files) != 1 || m.Files[0].Path != "report.md" {
		t.Fatalf("manifest files = %#v", m.Files)
	}
}
