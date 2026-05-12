package server

import (
	"net/http"
	"regexp"
	"strings"
)

const annotationRuntimePath = "/_pageshelf/runtime/annotation/"

type runtimeAsset struct {
	contentType string
	body        string
}

const annotationRuntimeMarkup = `<div data-pageshelf-annotation></div><link rel="stylesheet" href="/_pageshelf/runtime/annotation/annotation.css"><script defer src="/_pageshelf/runtime/annotation/annotation.js" id="pageshelf-annotation-script"></script>`

func serveAnnotationRuntimeAsset(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, annotationRuntimePath)
	asset, ok := annotationRuntimeAssets[name]
	if !ok || strings.Contains(name, "/") {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", asset.contentType)
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	if r.Method != http.MethodHead {
		_, _ = w.Write([]byte(asset.body))
	}
}

func injectAnnotationRuntime(html []byte) []byte {
	s := string(html)
	if hasAnnotationRuntime(s) {
		return html
	}
	s = stripMetaCSP(s)
	lower := strings.ToLower(s)
	idx := strings.LastIndex(lower, "</body>")
	if idx < 0 {
		return append([]byte(s), []byte(annotationRuntimeMarkup)...)
	}
	out := make([]byte, 0, len(s)+len(annotationRuntimeMarkup))
	out = append(out, s[:idx]...)
	out = append(out, annotationRuntimeMarkup...)
	out = append(out, s[idx:]...)
	return out
}

func hasAnnotationRuntime(s string) bool {
	return strings.Contains(s, `id="pageshelf-annotation-script"`) || strings.Contains(s, `id='pageshelf-annotation-script'`)
}

var metaTagRe = regexp.MustCompile(`(?is)<meta\b[^>]*>`)

func stripMetaCSP(s string) string {
	return metaTagRe.ReplaceAllStringFunc(s, func(tag string) string {
		lower := strings.ToLower(tag)
		if strings.Contains(lower, "http-equiv") && strings.Contains(lower, "content-security-policy") {
			return ""
		}
		return tag
	})
}

func isHTML(mime, path string) bool {
	mime = strings.ToLower(strings.TrimSpace(strings.Split(mime, ";")[0]))
	if mime != "" && mime != "application/octet-stream" {
		return mime == "text/html"
	}
	path = strings.ToLower(path)
	return strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".htm")
}
