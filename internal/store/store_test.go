package store

import (
	"fmt"
	"strings"
	"testing"
	"time"
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

func TestCreatePutLoad(t *testing.T) {
	st, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m, tok, err := st.Create("demo", "index.html", false)
	if err != nil {
		t.Fatal(err)
	}
	if tok != "" {
		t.Fatalf("expected no read token, got %q", tok)
	}
	if _, err = st.Put(m.ID, "index.html", strings.NewReader("<h1>x</h1>"), false); err != nil {
		t.Fatal(err)
	}
	m, err = st.Load(m.ID)
	if err != nil || len(m.Files) != 1 || m.Files[0].Path != "index.html" {
		t.Fatalf("bad manifest %#v %v", m, err)
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

func TestCreateWithOptionsAddsMetadataAndDefaultExpiry(t *testing.T) {
	st, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	before := time.Now().Add(DefaultTTL - time.Minute)
	m, tok, err := st.CreateWithOptions(CreateOptions{Name: "demo", Slug: "demo", Tags: []string{"Plan", "agent,report", "plan"}})
	if err != nil {
		t.Fatal(err)
	}
	if tok != "" {
		t.Fatalf("expected no read token, got %q", tok)
	}
	if got := strings.Join(m.Tags, ","); got != "agent,plan,report" {
		t.Fatalf("tags = %q", got)
	}
	if m.ExpiresAt.Before(before) || m.ExpiresAt.After(time.Now().Add(DefaultTTL+time.Minute)) {
		t.Fatalf("unexpected expires_at %s", m.ExpiresAt)
	}
}

func TestUpdateMetadataAndGC(t *testing.T) {
	st, err := New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	expiredAt := time.Now().Add(-time.Hour)
	m, _, err := st.CreateWithOptions(CreateOptions{Name: "old", Slug: "old", Tags: []string{"old"}, ExpiresAt: expiredAt})
	if err != nil {
		t.Fatal(err)
	}
	future := time.Now().Add(time.Hour)
	updated, err := st.UpdateMetadata(m.ID, []string{"keep", "review"}, &future)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(updated.Tags, ",") != "keep,review" || !updated.ExpiresAt.Equal(future) {
		t.Fatalf("bad metadata: %#v", updated)
	}
	dry, err := st.GC(time.Now(), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(dry.Removed) != 0 || dry.Kept != 1 {
		t.Fatalf("dry gc = %#v", dry)
	}
	past := time.Now().Add(-time.Minute)
	if _, err := st.UpdateMetadata(m.ID, nil, &past); err != nil {
		t.Fatal(err)
	}
	res, err := st.GC(time.Now(), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Removed) != 1 || res.Removed[0] != m.ID {
		t.Fatalf("gc = %#v", res)
	}
	if _, err := st.Load(m.ID); err == nil {
		t.Fatal("expected expired session to be removed")
	}
}
