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
.metadata { margin: 0 0 1.5rem; padding: .75rem 1rem; border: 1px solid color-mix(in srgb, CanvasText 16%, Canvas); border-radius: .75rem; background: color-mix(in srgb, CanvasText 4%, Canvas); color: color-mix(in srgb, CanvasText 70%, Canvas); font-size: .9rem; }
.metadata dl { margin: 0; display: grid; grid-template-columns: max-content 1fr; gap: .25rem .75rem; }
.metadata dt { font-weight: 650; }
.metadata dd { margin: 0; overflow-wrap: anywhere; }
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
{{if .Meta}}
<aside class="metadata" aria-label="Document metadata">
<dl>{{range $key, $value := .Meta}}<dt>{{$key}}</dt><dd>{{$value}}</dd>{{end}}</dl>
</aside>
{{end}}
{{.Body}}
</main>
</body>
</html>
`))

type pageData struct {
	Title string
	Meta  map[string]string
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

func Title(source, fallbackName string, meta map[string]string) string {
	if meta != nil {
		if title := strings.TrimSpace(meta["title"]); title != "" {
			return stripQuotes(title)
		}
	}
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
	bodySource, meta := stripFrontmatter(source)
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithXHTML()),
	)
	var body bytes.Buffer
	if err := md.Convert(bodySource, &body); err != nil {
		return nil, err
	}
	var page bytes.Buffer
	err := pageTemplate.Execute(&page, pageData{Title: Title(string(bodySource), name, meta), Meta: meta, Body: template.HTML(body.String())})
	return page.Bytes(), err
}

func stripFrontmatter(source []byte) ([]byte, map[string]string) {
	text := string(source)
	text = strings.TrimPrefix(text, "\ufeff")
	if !strings.HasPrefix(text, "---\n") && !strings.HasPrefix(text, "---\r\n") {
		return source, nil
	}
	newline := "\n"
	if strings.HasPrefix(text, "---\r\n") {
		newline = "\r\n"
	}
	start := len("---" + newline)
	endMarker := newline + "---" + newline
	end := strings.Index(text[start:], endMarker)
	if end < 0 {
		return source, nil
	}
	fm := text[start : start+end]
	body := text[start+end+len(endMarker):]
	return []byte(body), parseSimpleYAML(fm)
}

func parseSimpleYAML(fm string) map[string]string {
	meta := map[string]string{}
	for _, line := range strings.Split(strings.ReplaceAll(fm, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		key := strings.ToLower(strings.TrimSpace(parts[0]))
		value := strings.TrimSpace(parts[1])
		if key == "" || value == "" {
			continue
		}
		meta[key] = stripQuotes(value)
	}
	if len(meta) == 0 {
		return nil
	}
	return meta
}

func stripInlineMarkup(s string) string {
	s = strings.Trim(s, " `*_~")
	s = strings.ReplaceAll(s, "`", "")
	s = strings.ReplaceAll(s, "*", "")
	s = strings.ReplaceAll(s, "_", "")
	s = strings.ReplaceAll(s, "~", "")
	return s
}

func stripQuotes(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
