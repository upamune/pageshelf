package markdown

import (
	"bytes"
	"html/template"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var firstH1Re = regexp.MustCompile(`(?m)^#\s+(.+?)\s*#*\s*$`)

var pageTemplate = template.Must(template.New("markdown-page").Parse(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{.Title}}</title>
<style>
:root { color-scheme: light dark; }
body { margin: 0; font: 16px/1.6 system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; background: Canvas; color: CanvasText; }
main { max-width: 48rem; margin: 0 auto; padding: 2rem 1rem 4rem; }
img, svg, video { max-width: 100%; height: auto; }
pre { overflow-x: auto; padding: 1rem; border-radius: .5rem; background: color-mix(in srgb, CanvasText 8%, Canvas); }
code { font-family: ui-monospace, SFMono-Regular, Consolas, "Liberation Mono", monospace; }
:not(pre) > code { padding: .1rem .25rem; border-radius: .25rem; background: color-mix(in srgb, CanvasText 8%, Canvas); }
blockquote { margin-left: 0; padding-left: 1rem; border-left: .25rem solid color-mix(in srgb, CanvasText 25%, Canvas); color: color-mix(in srgb, CanvasText 75%, Canvas); }
table { border-collapse: collapse; width: 100%; display: block; overflow-x: auto; }
th, td { border: 1px solid color-mix(in srgb, CanvasText 20%, Canvas); padding: .35rem .5rem; }
a { color: LinkText; }
</style>
</head>
<body>
<main>
{{.Body}}
</main>
</body>
</html>
`))

type pageData struct {
	Title string
	Body  template.HTML
}

func IsMarkdownPath(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".md" || ext == ".markdown"
}

func HTMLPath(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return path + ".html"
	}
	return strings.TrimSuffix(path, ext) + ".html"
}

func Title(source, fallbackName string) string {
	if m := firstH1Re.FindStringSubmatch(source); len(m) == 2 {
		return strings.TrimSpace(stripInlineMarkup(m[1]))
	}
	base := filepath.Base(fallbackName)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	base = strings.ReplaceAll(base, "-", " ")
	base = strings.ReplaceAll(base, "_", " ")
	base = strings.TrimSpace(base)
	if base == "" || base == "." {
		return "Markdown artifact"
	}
	return base
}

func Render(source []byte, name string) ([]byte, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithXHTML()),
	)
	var body bytes.Buffer
	if err := md.Convert(source, &body); err != nil {
		return nil, err
	}
	var page bytes.Buffer
	err := pageTemplate.Execute(&page, pageData{Title: Title(string(source), name), Body: template.HTML(body.String())})
	return page.Bytes(), err
}

func stripInlineMarkup(s string) string {
	s = strings.Trim(s, " `*_~")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "*", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "~", "")
	return s
}
