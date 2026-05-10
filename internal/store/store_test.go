package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSafeRelRejectsTraversal(t *testing.T) {
	bad := []string{"../x", "/x", "a\\b", "a/../../b", "a\x00b"}
	for _, p := range bad {
		if _, err := SafeRel(p); err == nil {
			t.Fatalf("accepted %q", p)
		}
	}
	got, err := SafeRel("a/./b.html")
	if err != nil || got != "a/b.html" {
		t.Fatalf("got %q %v", got, err)
	}
}
func TestTokenHash(t *testing.T) {
	tok, hash, err := NewToken()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(tok, "psr_") || !CheckToken(tok, hash) || CheckToken(tok+"x", hash) {
		t.Fatal("token check failed")
	}
}
func TestCreatePutLoad(t *testing.T) {
	st, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m, tok, err := st.Create("demo", "index.html", false)
	if err != nil || tok == "" {
		t.Fatal(err)
	}
	if _, err = st.Put(m.ID, "index.html", strings.NewReader("<h1>x</h1>"), false); err != nil {
		t.Fatal(err)
	}
	m, err = st.Load(m.ID)
	if err != nil || len(m.Files) != 1 || m.Files[0].Path != "index.html" {
		t.Fatalf("bad manifest %#v %v", m, err)
	}
}

func TestReadTokenValidatesSessionID(t *testing.T) {
	st, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := st.ReadToken("../evil"); err == nil {
		t.Fatal("expected invalid session id")
	}
	if err := os.MkdirAll(filepath.Join(st.Root, "sessions", "ok"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(st.Root, "sessions", "ok", "read_token"), []byte("tok"), 0600); err != nil {
		t.Fatal(err)
	}
	got, err := st.ReadToken("ok")
	if err != nil || got != "tok" {
		t.Fatalf("got %q err %v", got, err)
	}
}

func TestPutReplacementDoesNotConsumeFileLimit(t *testing.T) {
	st, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m, _, err := st.Create("demo", "demo", false)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := st.Load(m.ID)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < MaxSessionFiles; i++ {
		manifest.Files = append(manifest.Files, File{Path: fmt.Sprintf("f%04d.txt", i), MIME: "text/plain"})
	}
	manifest.Files[0].Path = "index.html"
	if err := st.save(manifest); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Put(m.ID, "index.html", strings.NewReader("replacement"), false); err != nil {
		t.Fatal(err)
	}
	if _, err := st.Put(m.ID, "new.txt", strings.NewReader("new"), false); err == nil {
		t.Fatal("expected file limit for new file")
	}
}
