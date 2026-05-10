package markdown

import (
	"strings"
	"testing"
)

func TestRenderStripsFrontmatterAndUsesMetadataTitle(t *testing.T) {
	src := []byte("---\ntitle: \"Frontmatter Title\"\nauthor: \"@fcoury\"\ndate: 2026-05-01\nsource: https://example.com/x\n---\n\n# Body Title\n\nHello **world**.\n")
	out, err := Render(src, "note.md")
	if err != nil {
		t.Fatal(err)
	}
	html := string(out)
	if strings.Contains(html, "---") || strings.Contains(html, "title:") || strings.Contains(html, "author:") {
		t.Fatalf("frontmatter leaked into rendered HTML:\n%s", html)
	}
	if !strings.Contains(html, "<title>Frontmatter Title</title>") {
		t.Fatalf("metadata title not used:\n%s", html)
	}
	if !strings.Contains(html, "<h1 id=\"body-title\">Body Title</h1>") || !strings.Contains(html, "<strong>world</strong>") {
		t.Fatalf("markdown body not rendered:\n%s", html)
	}
	if !strings.Contains(html, "<dt>author</dt><dd>@fcoury</dd>") || !strings.Contains(html, "<dt>source</dt><dd>https://example.com/x</dd>") {
		t.Fatalf("metadata summary missing:\n%s", html)
	}
}

func TestRenderWithoutFrontmatterStillWorks(t *testing.T) {
	out, err := Render([]byte("# Plain\n\ntext"), "plain.md")
	if err != nil {
		t.Fatal(err)
	}
	html := string(out)
	if !strings.Contains(html, "<title>Plain</title>") || strings.Contains(html, "<aside class=\"metadata\"") {
		t.Fatalf("unexpected render output:\n%s", html)
	}
}
