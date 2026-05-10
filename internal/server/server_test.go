package server

import (
	"github.com/serizawa/pageshelf/internal/store"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestArtifactAuthAndHeaders(t *testing.T) {
	st, _ := store.New(t.TempDir())
	m, tok, err := st.Create("", "artifact", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = st.Put(m.ID, "index.html", strings.NewReader("ok"), false); err != nil {
		t.Fatal(err)
	}
	h := Server{Store: st}.Handler()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/"+m.ID+"/index.html?t=bad", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("want 401 got %d", rr.Code)
	}
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/a/"+m.ID+"/index.html?t="+tok, nil)
	h.ServeHTTP(rr, req)
	b, _ := io.ReadAll(rr.Body)
	if rr.Code != 200 || string(b) != "ok" {
		t.Fatalf("%d %q", rr.Code, string(b))
	}
	if rr.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("missing nosniff")
	}
	if !strings.Contains(rr.Header().Get("Content-Security-Policy"), "script-src 'none'") {
		t.Fatal(rr.Header().Get("Content-Security-Policy"))
	}
}
