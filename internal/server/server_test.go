package server

import (
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/serizawa/pageshelf/internal/store"
)

func TestArtifactServingAndHeaders(t *testing.T) {
	st, _ := store.New(t.TempDir())
	m, _, err := st.Create("", "artifact", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = st.Put(m.ID, "index.html", strings.NewReader("<!doctype html><html><body>ok</body></html>"), false); err != nil {
		t.Fatal(err)
	}
	if _, err = st.Put(m.ID, "plain.txt", strings.NewReader("ok"), false); err != nil {
		t.Fatal(err)
	}
	h := Server{Store: st}.Handler()
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/"+m.ID+"/index.html", nil)
	h.ServeHTTP(rr, req)
	b, _ := io.ReadAll(rr.Body)
	if rr.Code != 200 || !strings.Contains(string(b), "ok") {
		t.Fatalf("%d %q", rr.Code, string(b))
	}
	if !strings.Contains(string(b), "pageshelf-annotation-script") || !strings.Contains(string(b), `data-pageshelf-annotation`) {
		t.Fatalf("missing annotation runtime: %q", string(b))
	}
	for _, marker := range []string{
		`href="/_pageshelf/runtime/annotation/annotation.css"`,
		`defer src="/_pageshelf/runtime/annotation/annotation.js"`,
	} {
		if !strings.Contains(string(b), marker) {
			t.Fatalf("runtime markup missing marker %q", marker)
		}
	}
	if strings.Contains(string(b), "Copy annotations") || strings.Contains(string(b), "attachShadow") {
		t.Fatalf("runtime script was inlined into HTML: %q", string(b))
	}
	htmlCSP := rr.Header().Get("Content-Security-Policy")
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/_pageshelf/runtime/annotation/annotation.js", nil)
	h.ServeHTTP(rr, req)
	assetBody := rr.Body.String()
	if rr.Code != 200 || !strings.Contains(rr.Header().Get("Content-Type"), "text/javascript") {
		t.Fatalf("bad runtime script response: code=%d content-type=%q", rr.Code, rr.Header().Get("Content-Type"))
	}
	for _, marker := range []string{"attachShadow", "pageshelf.annotation.v1", "Copy notes", "Copy as JSON", "Download JSON", "Clear resolved", "Resolve", ":host"} {
		if !strings.Contains(assetBody, marker) {
			t.Fatalf("runtime script missing marker %q", marker)
		}
	}
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/_pageshelf/runtime/annotation/annotation.css", nil)
	h.ServeHTTP(rr, req)
	stylesheetBody := rr.Body.String()
	if rr.Code != 200 || !strings.Contains(rr.Header().Get("Content-Type"), "text/css") || !strings.Contains(stylesheetBody, ".ps-ann-highlight") || !strings.Contains(stylesheetBody, ".ps-ann-pick") {
		t.Fatalf("bad runtime stylesheet response: code=%d content-type=%q body=%q", rr.Code, rr.Header().Get("Content-Type"), rr.Body.String())
	}
	for _, marker := range []string{".sheet", ".btn", "textarea", ":host"} {
		if strings.Contains(stylesheetBody, marker) {
			t.Fatalf("runtime stylesheet leaks shadow DOM marker %q: %q", marker, stylesheetBody)
		}
	}
	if rr.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("missing nosniff")
	}
	if !strings.Contains(htmlCSP, "script-src 'self' 'unsafe-inline'") {
		t.Fatal(htmlCSP)
	}
	if !strings.Contains(htmlCSP, "connect-src 'none'") {
		t.Fatal(htmlCSP)
	}
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("HEAD", "/a/"+m.ID+"/index.html", nil)
	h.ServeHTTP(rr, req)
	if rr.Code != 200 || rr.Body.Len() != 0 || rr.Header().Get("Last-Modified") == "" {
		t.Fatalf("bad HEAD response: code=%d body=%q last-modified=%q", rr.Code, rr.Body.String(), rr.Header().Get("Last-Modified"))
	}
	rr = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/a/"+m.ID+"/plain.txt", nil)
	h.ServeHTTP(rr, req)
	b, _ = io.ReadAll(rr.Body)
	if rr.Code != 200 || string(b) != "ok" {
		t.Fatalf("non-html response changed: %d %q", rr.Code, string(b))
	}
	if strings.Contains(string(b), "pageshelf-annotation-script") {
		t.Fatal("annotation runtime injected into non-html")
	}
}

func TestAllowImageSrcCSP(t *testing.T) {
	st, _ := store.New(t.TempDir())
	m, _, err := st.Create("", "artifact", false)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = st.Put(m.ID, "index.html", strings.NewReader("<!doctype html><html><body><img src=\"https://i.gyazo.com/example.jpg\"></body></html>"), false); err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/"+m.ID+"/index.html", nil)
	Server{Store: st, AllowImageSrc: []string{"https://i.gyazo.com"}}.Handler().ServeHTTP(rr, req)
	csp := rr.Header().Get("Content-Security-Policy")
	if !strings.Contains(csp, "img-src 'self' data: blob: https://i.gyazo.com;") {
		t.Fatalf("unexpected CSP: %q", csp)
	}
	if !strings.Contains(csp, "connect-src 'none'") {
		t.Fatalf("connect-src should stay locked down: %q", csp)
	}
}

func TestInjectAnnotationRuntime(t *testing.T) {
	in := []byte("<!doctype html><html><body><h1>Hi</h1></body></html>")
	out := string(injectAnnotationRuntime(in))
	if strings.Count(out, "pageshelf-annotation-script") != 1 {
		t.Fatalf("runtime injection count mismatch: %q", out)
	}
	if !strings.Contains(out, `</script></body>`) {
		t.Fatalf("runtime was not inserted before body close: %q", out)
	}
	again := string(injectAnnotationRuntime([]byte(out)))
	if strings.Count(again, "pageshelf-annotation-script") != 1 {
		t.Fatalf("runtime injected more than once: %q", again)
	}
	withoutBody := string(injectAnnotationRuntime([]byte("<main>Hi</main>")))
	if !strings.Contains(withoutBody, "pageshelf-annotation-script") {
		t.Fatalf("runtime not appended when body close missing: %q", withoutBody)
	}
	spoof := string(injectAnnotationRuntime([]byte("<html><body>text pageshelf-annotation-script only</body></html>")))
	if strings.Count(spoof, "pageshelf-annotation-script") != 2 {
		t.Fatalf("ordinary text spoof blocked injection: %q", spoof)
	}
	withMeta := string(injectAnnotationRuntime([]byte(`<!doctype html><html><head><meta http-equiv="Content-Security-Policy" content="script-src 'none'"></head><body>ok</body></html>`)))
	if strings.Contains(strings.ToLower(withMeta), "http-equiv") || !strings.Contains(withMeta, "pageshelf-annotation-script") {
		t.Fatalf("meta CSP not stripped or runtime missing: %q", withMeta)
	}
}

func TestDisabledAnnotations(t *testing.T) {
	st, _ := store.New(t.TempDir())
	m, _, err := st.CreateWithOptions(store.CreateOptions{Slug: "artifact", DisableAnnotations: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = st.Put(m.ID, "index.html", strings.NewReader("<!doctype html><html><body>ok</body></html>"), false); err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/a/"+m.ID+"/index.html", nil)
	Server{Store: st}.Handler().ServeHTTP(rr, req)
	body := rr.Body.String()
	if strings.Contains(body, "pageshelf-annotation-script") {
		t.Fatalf("runtime injected despite disabled annotations: %q", body)
	}
	if !strings.Contains(rr.Header().Get("Content-Security-Policy"), "script-src 'none'") || !strings.Contains(rr.Header().Get("Content-Security-Policy"), "connect-src 'none'") {
		t.Fatalf("unexpected disabled CSP: %q", rr.Header().Get("Content-Security-Policy"))
	}
}

func TestIsHTML(t *testing.T) {
	if !isHTML("text/html; charset=utf-8", "") || !isHTML("", "index.htm") || !isHTML("", "INDEX.HTML") {
		t.Fatal("expected html detection")
	}
	if !isHTML("application/octet-stream", "index.html") {
		t.Fatal("expected octet-stream extension fallback")
	}
	if isHTML("text/plain", "index.html") || isHTML("application/json", "data.json") {
		t.Fatal("expected non-html to remain non-html")
	}
}
