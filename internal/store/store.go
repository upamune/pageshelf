// Package store manages persisted Pageshelf sessions and artifacts.
package store

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// Version is the Pageshelf application version reported by the CLI.
// Release builds set it from the git tag via GoReleaser ldflags.
var Version = "0.2.0"

const (
	// MaxFileSize is the maximum stored artifact size in bytes.
	MaxFileSize int64 = 25 << 20
	// MaxSessionFiles is the maximum number of files in one session.
	MaxSessionFiles = 1000
	// DefaultTTL is the default session lifetime.
	DefaultTTL = 14 * 24 * time.Hour
)

var (
	idRe    = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,79}$`)
	errPath = errors.New("invalid artifact path")
)

// Store provides filesystem-backed session storage.
type Store struct{ Root string }

// CreateOptions configures session creation.
type CreateOptions struct {
	Name        string
	Slug        string
	Interactive bool
	Tags        []string
	TTL         time.Duration
	ExpiresAt   time.Time
}

// GCResult summarizes garbage collection results.
type GCResult struct {
	Removed []string `json:"removed"`
	Kept    int      `json:"kept"`
}

// Manifest describes a stored Pageshelf session.
type Manifest struct {
	ID            string    `json:"id"`
	Name          string    `json:"name,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	ExpiresAt     time.Time `json:"expires_at"`
	Tags          []string  `json:"tags,omitempty"`
	Interactive   bool      `json:"interactive"`
	ReadTokenHash string    `json:"read_token_hash"`
	Files         []File    `json:"files"`
}

// File describes a stored artifact file.
type File struct {
	Path      string    `json:"path"`
	MIME      string    `json:"mime"`
	Size      int64     `json:"size"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DefaultDir returns the default Pageshelf data directory.
func DefaultDir() string {
	if v := os.Getenv("PAGESHELF_DATA_DIR"); v != "" {
		return v
	}
	d, _ := os.UserHomeDir()
	return filepath.Join(d, ".local", "share", "pageshelf")
}

// New creates a Store rooted at root, or the default directory if root is empty.
func New(root string) (*Store, error) {
	if root == "" {
		root = DefaultDir()
	}
	s := &Store{Root: root}
	return s, os.MkdirAll(filepath.Join(root, "sessions"), 0o700)
}

// SessionDir returns the filesystem path for a session.
func (s *Store) SessionDir(id string) string { return filepath.Join(s.Root, "sessions", id) }

// ValidateSessionID reports whether id is a valid session identifier.
func ValidateSessionID(id string) error {
	if !idRe.MatchString(id) || strings.Contains(id, "..") {
		return fmt.Errorf("invalid session id")
	}
	return nil
}

// SafeRel normalizes p to a safe relative artifact path.
func SafeRel(p string) (string, error) {
	if p == "" {
		return "index.html", nil
	}
	if strings.ContainsAny(p, "\x00\\") || filepath.IsAbs(p) {
		return "", errPath
	}
	c := filepath.Clean(filepath.ToSlash(p))
	if c == "." {
		return "index.html", nil
	}
	if strings.HasPrefix(c, "../") || c == ".." {
		return "", errPath
	}
	return c, nil
}

// Slug converts s into a stable URL-safe slug.
func Slug(s string) string {
	s = strings.TrimSuffix(filepath.Base(s), filepath.Ext(s))
	if s == "" || s == "index" {
		s = "artifact"
	}
	s = strings.ToLower(s)
	var b strings.Builder
	dash := false
	for _, r := range s {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if ok {
			b.WriteRune(r)
			dash = false
		} else if !dash {
			b.WriteByte('-')
			dash = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		out = "artifact"
	}
	if len(out) > 32 {
		out = out[:32]
	}
	return out
}
func random(n int) ([]byte, error) { b := make([]byte, n); _, e := rand.Read(b); return b, e }

// NewToken returns a read token and its SHA-256 hash.
func NewToken() (string, string, error) {
	b, e := random(32)
	if e != nil {
		return "", "", e
	}
	tok := "psr_" + base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(tok))
	return tok, hex.EncodeToString(h[:]), nil
}

// CheckToken reports whether tok matches hash.
func CheckToken(tok, hash string) bool {
	h := sha256.Sum256([]byte(tok))
	got := hex.EncodeToString(h[:])
	return subtle.ConstantTimeCompare([]byte(got), []byte(hash)) == 1
}

// NormalizeTags canonicalizes, deduplicates, and sorts tags.
func NormalizeTags(tags []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, tag := range tags {
		for _, part := range strings.Split(tag, ",") {
			part = strings.ToLower(strings.TrimSpace(part))
			part = strings.Trim(part, "#")
			if part == "" || seen[part] {
				continue
			}
			seen[part] = true
			out = append(out, part)
		}
	}
	sort.Strings(out)
	return out
}

// Create creates a session with the supplied basic metadata.
func (s *Store) Create(name, slug string, interactive bool) (*Manifest, string, error) {
	return s.CreateWithOptions(CreateOptions{Name: name, Slug: slug, Interactive: interactive})
}

// CreateWithOptions creates a session using opts.
func (s *Store) CreateWithOptions(opts CreateOptions) (*Manifest, string, error) {
	name := opts.Name
	slug := opts.Slug
	if slug == "" {
		slug = name
	}
	now := time.Now()
	expiresAt := opts.ExpiresAt
	if expiresAt.IsZero() {
		ttl := opts.TTL
		if ttl == 0 {
			ttl = DefaultTTL
		}
		if ttl > 0 {
			expiresAt = now.Add(ttl)
		}
	}
	rb, _ := random(4)
	id := now.Format("20060102-1504") + "-" + Slug(slug) + "-" + strings.ToLower(base64.RawURLEncoding.EncodeToString(rb))[:6]
	tok, hash, e := NewToken()
	if e != nil {
		return nil, "", e
	}
	m := &Manifest{ID: id, Name: name, CreatedAt: now, UpdatedAt: now, ExpiresAt: expiresAt, Tags: NormalizeTags(opts.Tags), Interactive: opts.Interactive, ReadTokenHash: hash}
	if e = s.save(m); e != nil {
		return nil, "", e
	}
	e = os.WriteFile(filepath.Join(s.SessionDir(id), "read_token"), []byte(tok), 0o600)
	return m, tok, e
}

func (s *Store) save(m *Manifest) error {
	if e := ValidateSessionID(m.ID); e != nil {
		return e
	}
	dir := s.SessionDir(m.ID)
	if e := os.MkdirAll(filepath.Join(dir, "files"), 0o700); e != nil {
		return e
	}
	m.UpdatedAt = time.Now()
	data, e := json.MarshalIndent(m, "", "  ")
	if e != nil {
		return e
	}
	return os.WriteFile(filepath.Join(dir, "manifest.json"), data, 0o600)
}

// Load reads a session manifest by ID.
func (s *Store) Load(id string) (*Manifest, error) {
	if e := ValidateSessionID(id); e != nil {
		return nil, e
	}
	b, e := os.ReadFile(filepath.Join(s.SessionDir(id), "manifest.json"))
	if e != nil {
		return nil, e
	}
	var m Manifest
	e = json.Unmarshal(b, &m)
	return &m, e
}

// ReadToken reads the persisted read token for a session.
func (s *Store) ReadToken(id string) (string, error) {
	if e := ValidateSessionID(id); e != nil {
		return "", e
	}
	b, e := os.ReadFile(filepath.Join(s.SessionDir(id), "read_token"))
	return string(b), e
}

// Put writes an artifact file into a session.
func (s *Store) Put(id, rel string, r io.Reader, interactive bool) (File, error) {
	m, e := s.Load(id)
	if e != nil {
		return File{}, e
	}
	rel, e = SafeRel(rel)
	if e != nil {
		return File{}, e
	}
	replaced := false
	for i := range m.Files {
		if m.Files[i].Path == rel {
			replaced = true
			break
		}
	}
	if !replaced && len(m.Files) >= MaxSessionFiles {
		return File{}, fmt.Errorf("file limit exceeded")
	}
	dst := filepath.Join(s.SessionDir(id), "files", filepath.FromSlash(rel))
	base := filepath.Join(s.SessionDir(id), "files")
	if !strings.HasPrefix(dst, base+string(os.PathSeparator)) && dst != base {
		return File{}, errPath
	}
	if e = os.MkdirAll(filepath.Dir(dst), 0o700); e != nil {
		return File{}, e
	}
	tmp := dst + ".tmp"
	f, e := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if e != nil {
		return File{}, e
	}
	n, e := io.Copy(f, io.LimitReader(r, MaxFileSize+1))
	ce := f.Close()
	if e == nil {
		e = ce
	}
	if e != nil {
		_ = os.Remove(tmp)
		return File{}, e
	}
	if n > MaxFileSize {
		_ = os.Remove(tmp)
		return File{}, fmt.Errorf("file too large")
	}
	if e = os.Rename(tmp, dst); e != nil {
		return File{}, e
	}
	mt := mime.TypeByExtension(filepath.Ext(rel))
	if mt == "" {
		mt = "application/octet-stream"
	}
	fi := File{Path: rel, MIME: mt, Size: n, UpdatedAt: time.Now()}
	for i := range m.Files {
		if m.Files[i].Path == rel {
			m.Files[i] = fi
			replaced = true
		}
	}
	if !replaced {
		m.Files = append(m.Files, fi)
	}
	sort.Slice(m.Files, func(i, j int) bool { return m.Files[i].Path < m.Files[j].Path })
	m.Interactive = m.Interactive || interactive
	return fi, s.save(m)
}

// Open opens an artifact file and returns its metadata and manifest.
func (s *Store) Open(id, rel string) (*os.File, File, *Manifest, error) {
	m, e := s.Load(id)
	if e != nil {
		return nil, File{}, nil, e
	}
	rel, e = SafeRel(rel)
	if e != nil {
		return nil, File{}, nil, e
	}
	var meta File
	ok := false
	for _, f := range m.Files {
		if f.Path == rel {
			meta = f
			ok = true
		}
	}
	if !ok {
		return nil, File{}, nil, os.ErrNotExist
	}
	p := filepath.Join(s.SessionDir(id), "files", filepath.FromSlash(rel))
	f, e := os.Open(p)
	return f, meta, m, e
}

// List returns all session manifests ordered by creation time descending.
func (s *Store) List() ([]Manifest, error) {
	ents, e := os.ReadDir(filepath.Join(s.Root, "sessions"))
	if e != nil {
		return nil, e
	}
	var out []Manifest
	for _, en := range ents {
		if en.IsDir() {
			if m, e := s.Load(en.Name()); e == nil {
				out = append(out, *m)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

// Remove deletes a session by ID.
func (s *Store) Remove(id string) error {
	if e := ValidateSessionID(id); e != nil {
		return e
	}
	return os.RemoveAll(s.SessionDir(id))
}

// UpdateMetadata updates mutable session metadata.
func (s *Store) UpdateMetadata(id string, tags []string, expiresAt *time.Time) (*Manifest, error) {
	m, e := s.Load(id)
	if e != nil {
		return nil, e
	}
	if tags != nil {
		m.Tags = NormalizeTags(tags)
	}
	if expiresAt != nil {
		m.ExpiresAt = *expiresAt
	}
	if e := s.save(m); e != nil {
		return nil, e
	}
	return m, nil
}

// GC removes expired sessions unless dryRun is true.
func (s *Store) GC(now time.Time, dryRun bool) (GCResult, error) {
	items, e := s.List()
	if e != nil {
		return GCResult{}, e
	}
	res := GCResult{}
	for _, m := range items {
		if m.ExpiresAt.IsZero() || m.ExpiresAt.After(now) {
			res.Kept++
			continue
		}
		res.Removed = append(res.Removed, m.ID)
		if !dryRun {
			if e := s.Remove(m.ID); e != nil {
				return res, e
			}
		}
	}
	return res, nil
}

// URL builds an authenticated artifact URL.
func URL(base, session, path, token string) string {
	if path == "" {
		path = "index.html"
	}
	return strings.TrimRight(base, "/") + "/a/" + url.PathEscape(session) + "/" + strings.TrimLeft(path, "/") + "?t=" + url.QueryEscape(token)
}
