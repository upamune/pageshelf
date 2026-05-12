// Package server provides HTTP serving for Pageshelf artifacts.
package server

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/serizawa/pageshelf/internal/store"
)

// Server serves Pageshelf artifacts over HTTP.
type Server struct{ Store *store.Store }

// Handler returns the HTTP handler for serving artifacts and health checks.
func (s Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(annotationRuntimePath, serveAnnotationRuntimeAsset)
	mux.HandleFunc("/a/", s.artifact)
	health := func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }
	mux.HandleFunc("/-/healthz", health)
	mux.HandleFunc("/healthz", health)
	return security(mux)
}

func security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		h.Set("Cache-Control", "no-store")
		h.Set("X-Frame-Options", "DENY")
		next.ServeHTTP(w, r)
	})
}

func csp(disableAnnotations bool) string {
	if disableAnnotations {
		return "default-src 'none'; script-src 'none'; style-src 'self'; img-src 'self' data: blob:; font-src 'self' data:; connect-src 'none'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'none'; worker-src 'none'; child-src 'none'; frame-src 'none'"
	}
	return "default-src 'none'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; font-src 'self' data:; connect-src 'none'; object-src 'none'; base-uri 'none'; frame-ancestors 'none'; form-action 'none'; worker-src 'none'; child-src 'none'; frame-src 'none'"
}

func (s Server) artifact(w http.ResponseWriter, r *http.Request) {
	parts := strings.SplitN(strings.TrimPrefix(r.URL.Path, "/a/"), "/", 2)
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}
	id, path := parts[0], parts[1]
	f, meta, m, err := s.Store.Open(id, path)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer func() { _ = f.Close() }()
	w.Header().Set("Content-Security-Policy", csp(m.DisableAnnotations))
	w.Header().Set("Content-Type", meta.MIME)
	if isHTML(meta.MIME, meta.Path) {
		b, err := io.ReadAll(f)
		if err != nil {
			http.Error(w, "read artifact", http.StatusInternalServerError)
			return
		}
		if !m.DisableAnnotations {
			b = injectAnnotationRuntime(b)
		}
		w.Header().Del("Content-Length")
		// HTML responses are rewritten in memory for annotation injection, so ServeContent
		// cannot set validators for us. Preserve the artifact's manifest timestamp.
		w.Header().Set("Last-Modified", meta.UpdatedAt.UTC().Format(http.TimeFormat))
		if r.Method == http.MethodHead {
			return
		}
		_, _ = w.Write(b)
		return
	}
	http.ServeContent(w, r, meta.Path, meta.UpdatedAt, f)
}

// ListenAndServe starts the HTTP server and shuts it down on interrupt signals.
func ListenAndServe(addr string, st *store.Store) error {
	log.Printf("pageshelf serving on http://%s", addr)
	srv := &http.Server{
		Addr:              addr,
		Handler:           Server{Store: st}.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
