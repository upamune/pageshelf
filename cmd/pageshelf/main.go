// Command pageshelf manages and serves local HTML artifact sessions.
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	mdrender "github.com/serizawa/pageshelf/internal/markdown"
	"github.com/serizawa/pageshelf/internal/secret"
	"github.com/serizawa/pageshelf/internal/server"
	"github.com/serizawa/pageshelf/internal/store"
	"github.com/serizawa/pageshelf/internal/tailscale"
	pageshelfskill "github.com/serizawa/pageshelf/skills/pageshelf"
)

type CLI struct {
	DataDir string     `help:"Storage directory." env:"PAGESHELF_DATA_DIR"`
	Serve   ServeCmd   `cmd:""`
	Put     PutCmd     `cmd:""`
	Session SessionCmd `cmd:""`
	List    ListCmd    `cmd:""`
	Files   FilesCmd   `cmd:""`
	URL     URLCmd     `cmd:""`
	GC      GCCmd      `cmd:"" help:"Remove expired sessions."`
	Skill   SkillCmd   `cmd:"" help:"Print or install the bundled agent skill."`
	Version VersionCmd `cmd:"" help:"Print version information."`
}
type (
	Ctx      struct{ Store *store.Store }
	ServeCmd struct {
		Tailscale        bool   `help:"Bind to detected Tailscale IP."`
		Host             string `default:"127.0.0.1"`
		Port             int    `default:"8787"`
		UnsafePublicBind bool
	}
)

type PutCmd struct {
	Session       string `short:"s"`
	Stdin         bool
	Name          string
	Content       string
	Interactive   bool     `help:"Deprecated no-op: HTML artifacts are interactive by default."`
	NoAnnotations bool     `name:"no-annotations" help:"Disable the default HTML annotation runtime for this session."`
	Raw           bool     `help:"Store Markdown files as-is instead of rendering .md/.markdown to HTML."`
	NoSecretScan  bool     `name:"no-secret-scan" help:"Disable default secret scanning before storing content."`
	Tag           []string `name:"tag" short:"t" help:"Tag for the session. Repeat or use comma-separated values."`
	TTL           string   `default:"14d" help:"Session retention duration, e.g. 14d, 48h, 0 for no expiry."`
	ExpiresAt     string   `name:"expires-at" help:"Explicit expiry timestamp (RFC3339) or date (YYYY-MM-DD)."`
	JSON          bool
	Host          string   `default:"127.0.0.1" help:"Host to use when printing the artifact URL."`
	Port          int      `default:"8787" help:"Port to use when printing the artifact URL."`
	Tailscale     bool     `help:"Use detected Tailscale IP when printing the artifact URL."`
	BaseURL       string   `name:"base-url" help:"Base URL to use when printing the artifact URL."`
	Paths         []string `arg:"" optional:"" name:"paths"`
}
type SessionCmd struct {
	Create SessionCreateCmd `cmd:""`
	Info   SessionInfoCmd   `cmd:""`
	Meta   SessionMetaCmd   `cmd:"" help:"Update session metadata."`
	Rm     SessionRmCmd     `cmd:""`
}
type SessionCreateCmd struct {
	Name      string   `arg:"" optional:""`
	Tag       []string `name:"tag" short:"t" help:"Tag for the session. Repeat or use comma-separated values."`
	TTL       string   `default:"14d" help:"Session retention duration, e.g. 14d, 48h, 0 for no expiry."`
	ExpiresAt string   `name:"expires-at" help:"Explicit expiry timestamp (RFC3339) or date (YYYY-MM-DD)."`
	JSON      bool
}
type SessionInfoCmd struct {
	Session string `arg:""`
	JSON    bool
}

type SessionMetaCmd struct {
	Session   string   `arg:""`
	Tag       []string `name:"tag" short:"t" help:"Replace tags. Repeat or use comma-separated values."`
	ClearTags bool     `name:"clear-tags" help:"Remove all tags."`
	TTL       string   `help:"Set expiry relative to now, e.g. 14d, 48h, 0 for no expiry."`
	ExpiresAt string   `name:"expires-at" help:"Set explicit expiry timestamp (RFC3339) or date (YYYY-MM-DD)."`
	JSON      bool
}
type SessionRmCmd struct {
	Session string `arg:""`
}
type (
	ListCmd  struct{ JSON bool }
	FilesCmd struct {
		Session string `arg:""`
		JSON    bool
	}
)

type URLCmd struct {
	Session   string `arg:""`
	Path      string `arg:"" optional:"" default:"index.html"`
	JSON      bool
	Host      string `default:"127.0.0.1" help:"Host to use in the URL."`
	Port      int    `default:"8787" help:"Port to use in the URL."`
	Tailscale bool   `help:"Use detected Tailscale IP in the URL."`
	BaseURL   string `name:"base-url" help:"Base URL to use instead of host/port."`
}
type GCCmd struct {
	DryRun bool `name:"dry-run" help:"List expired sessions without deleting them."`
	JSON   bool
}

type SkillCmd struct {
	Help    SkillHelpCmd    `cmd:"" default:"1" help:"Show bundled skill command help."`
	Show    SkillShowCmd    `cmd:"" help:"Print the bundled Pageshelf agent skill."`
	Install SkillInstallCmd `cmd:"" help:"Write the bundled Pageshelf agent skill to a directory."`
}

type SkillHelpCmd struct{}

type SkillShowCmd struct{}

type SkillInstallCmd struct {
	Dir string `arg:"" help:"Destination directory. Writes SKILL.md inside it."`
}

type VersionCmd struct{}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println(store.Version)
		return
	}
	var cli CLI
	k := kong.Parse(
		&cli,
		kong.Name("pageshelf"),
		kong.Description("agent-generated HTML artifact shelf v"+store.Version),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{Compact: true}),
		kong.Vars{"version": store.Version},
	)
	st, err := store.New(cli.DataDir)
	k.FatalIfErrorf(err)
	err = k.Run(&Ctx{Store: st})
	k.FatalIfErrorf(err)
}
func printJSON(v any) { b, _ := json.MarshalIndent(v, "", "  "); fmt.Println(string(b)) }

func printSkillHelp() {
	fmt.Print(strings.Join([]string{
		"Usage: pageshelf skill <command>",
		"",
		"Commands:",
		"  pageshelf skill show           print the bundled Pageshelf agent skill",
		"  pageshelf skill install <dir>  write SKILL.md into a skill directory",
		"  pageshelf skill help           show this help",
		"",
	}, "\n"))
}

func (c *SkillHelpCmd) Run(_ *Ctx) error {
	printSkillHelp()
	return nil
}

func (c *SkillShowCmd) Run(_ *Ctx) error {
	_, err := os.Stdout.Write(pageshelfskill.Content())
	return err
}

func (c *SkillInstallCmd) Run(_ *Ctx) error {
	if err := os.MkdirAll(c.Dir, 0o700); err != nil {
		return fmt.Errorf("create skill directory: %w", err)
	}
	path := filepath.Join(c.Dir, "SKILL.md")
	if err := os.WriteFile(path, pageshelfskill.Content(), 0o600); err != nil {
		return fmt.Errorf("write skill: %w", err)
	}
	fmt.Println(path)
	return nil
}

func (c *VersionCmd) Run(_ *Ctx) error {
	fmt.Println(store.Version)
	return nil
}

func baseURL(host string, port int) string {
	return fmt.Sprintf("http://%s", net.JoinHostPort(host, fmt.Sprint(port)))
}

func publicBaseURL(host string, port int, useTS bool, explicit string) (string, error) {
	if explicit != "" {
		return strings.TrimRight(explicit, "/"), nil
	}
	if useTS {
		if name, err := tailscale.MagicDNSName(); err == nil {
			host = name
			return baseURL(host, port), nil
		}
		ip, err := tailscale.IP()
		if err != nil {
			return "", err
		}
		host = ip
	}
	return baseURL(host, port), nil
}

func isPublicBind(h string) bool {
	h = strings.TrimSpace(strings.Trim(h, "[]"))
	return h == "" || h == "0.0.0.0" || h == "::"
}

func healthCheck(baseURL string) bool {
	client := http.Client{Timeout: 750 * time.Millisecond}
	resp, err := client.Get(strings.TrimRight(baseURL, "/") + "/healthz")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

func (c *ServeCmd) Run(ctx *Ctx) error {
	h := c.Host
	if c.Tailscale {
		ip, err := tailscale.IP()
		if err != nil {
			return err
		}
		h = ip
	}
	if isPublicBind(h) && !c.UnsafePublicBind {
		return fmt.Errorf("refusing public bind %q without --unsafe-public-bind", h)
	}
	addr := net.JoinHostPort(h, fmt.Sprint(c.Port))
	url := "http://" + addr
	if healthCheck(url) {
		fmt.Printf("pageshelf already serving at %s\n", url)
		return nil
	}
	return server.ListenAndServe(addr, ctx.Store)
}

func (c *SessionCreateCmd) Run(ctx *Ctx) error {
	ttl, expiresAt, e := parseRetention(c.TTL, c.ExpiresAt)
	if e != nil {
		return e
	}
	m, _, e := ctx.Store.CreateWithOptions(store.CreateOptions{Name: c.Name, Slug: c.Name, Tags: c.Tag, TTL: ttl, ExpiresAt: expiresAt})
	if e != nil {
		return e
	}
	if c.JSON {
		printJSON(map[string]any{"session": m})
	} else {
		fmt.Printf("%s\n", m.ID)
	}
	return nil
}

func (c *SessionInfoCmd) Run(ctx *Ctx) error {
	m, e := ctx.Store.Load(c.Session)
	if e != nil {
		return e
	}
	if c.JSON {
		printJSON(m)
	} else {
		fmt.Printf("%s (%d files)\ttags=%s\texpires_at=%s\n", m.ID, len(m.Files), strings.Join(m.Tags, ","), formatTime(m.ExpiresAt))
	}
	return nil
}

func (c *SessionMetaCmd) Run(ctx *Ctx) error {
	var tags []string
	var tagPtr []string
	if c.ClearTags {
		tagPtr = []string{}
	} else if len(c.Tag) > 0 {
		tags = c.Tag
		tagPtr = tags
	}
	var expiresPtr *time.Time
	if c.TTL != "" || c.ExpiresAt != "" {
		ttl, expiresAt, e := parseRetention(c.TTL, c.ExpiresAt)
		if e != nil {
			return e
		}
		if expiresAt.IsZero() {
			if ttl < 0 {
				expiresAt = time.Time{}
			} else {
				expiresAt = time.Now().Add(ttl)
			}
		}
		expiresPtr = &expiresAt
	}
	m, e := ctx.Store.UpdateMetadata(c.Session, tagPtr, expiresPtr)
	if e != nil {
		return e
	}
	if c.JSON {
		printJSON(m)
	} else {
		fmt.Printf("%s\ttags=%s\texpires_at=%s\n", m.ID, strings.Join(m.Tags, ","), formatTime(m.ExpiresAt))
	}
	return nil
}
func (c *SessionRmCmd) Run(ctx *Ctx) error { return ctx.Store.Remove(c.Session) }
func (c *ListCmd) Run(ctx *Ctx) error {
	xs, e := ctx.Store.List()
	if e != nil {
		return e
	}
	if c.JSON {
		printJSON(xs)
	} else {
		for _, m := range xs {
			fmt.Printf("%s\t%d files\ttags=%s\texpires_at=%s\n", m.ID, len(m.Files), strings.Join(m.Tags, ","), formatTime(m.ExpiresAt))
		}
	}
	return nil
}

func (c *FilesCmd) Run(ctx *Ctx) error {
	m, e := ctx.Store.Load(c.Session)
	if e != nil {
		return e
	}
	if c.JSON {
		printJSON(m.Files)
	} else {
		for _, f := range m.Files {
			fmt.Printf("%s\t%d\t%s\n", f.Path, f.Size, f.MIME)
		}
	}
	return nil
}

func (c *URLCmd) Run(ctx *Ctx) error {
	b, e := publicBaseURL(c.Host, c.Port, c.Tailscale, c.BaseURL)
	if e != nil {
		return e
	}
	u := store.URL(b, c.Session, c.Path)
	if c.JSON {
		printJSON(map[string]string{"url": u})
	} else {
		fmt.Println(u)
	}
	return nil
}

func (c *PutCmd) Run(ctx *Ctx) error {
	if !c.Stdin && c.Content == "" && len(c.Paths) == 0 {
		return fmt.Errorf("provide --stdin, --content, or paths")
	}
	items, e := c.collectPutItems()
	if e != nil {
		return e
	}

	sid := c.Session
	slug := c.Name
	if slug == "" && len(c.Paths) > 0 {
		slug = c.Paths[0]
	}
	if sid == "" {
		ttl, expiresAt, e := parseRetention(c.TTL, c.ExpiresAt)
		if e != nil {
			return e
		}
		m, _, e := ctx.Store.CreateWithOptions(store.CreateOptions{Name: "", Slug: slug, Interactive: c.Interactive, DisableAnnotations: c.NoAnnotations, Tags: c.Tag, TTL: ttl, ExpiresAt: expiresAt})
		if e != nil {
			return e
		}
		sid = m.ID
	} else {
		if c.NoAnnotations {
			if _, e := ctx.Store.SetDisableAnnotations(sid, true); e != nil {
				return e
			}
		}
	}
	added := []string{}
	for _, item := range items {
		if _, e := ctx.Store.Put(sid, item.name, bytes.NewReader(item.data), c.Interactive); e != nil {
			return e
		}
		added = append(added, item.name)
	}
	b, e := publicBaseURL(c.Host, c.Port, c.Tailscale, c.BaseURL)
	if e != nil {
		return e
	}
	u := store.URL(b, sid, preferredPutURLPath(added))
	if c.JSON {
		printJSON(map[string]string{"session": sid, "url": u})
	} else {
		fmt.Printf("session: %s\nurl: %s\n", sid, u)
	}
	return nil
}

type putItem struct {
	name string
	data []byte
}

func (c *PutCmd) collectPutItems() ([]putItem, error) {
	items := []putItem{}
	if c.Stdin {
		if c.Name == "" {
			return nil, fmt.Errorf("--stdin requires --name")
		}
		name, data, e := preparePutContentWithSecretScan(os.Stdin, c.Name, c.Raw, !c.NoSecretScan)
		if e != nil {
			return nil, e
		}
		items = append(items, putItem{name: name, data: data})
	}
	if c.Content != "" {
		if c.Name == "" {
			return nil, fmt.Errorf("--content requires --name")
		}
		name, data, e := preparePutContentWithSecretScan(strings.NewReader(c.Content), c.Name, c.Raw, !c.NoSecretScan)
		if e != nil {
			return nil, e
		}
		items = append(items, putItem{name: name, data: data})
	}
	for _, p := range c.Paths {
		pathItems, e := collectPutPathItems(p, c.Raw, !c.NoSecretScan)
		if e != nil {
			return nil, e
		}
		items = append(items, pathItems...)
	}
	return items, nil
}

func (c *GCCmd) Run(ctx *Ctx) error {
	res, e := ctx.Store.GC(time.Now(), c.DryRun)
	if e != nil {
		return e
	}
	if c.JSON {
		printJSON(res)
	} else {
		verb := "removed"
		if c.DryRun {
			verb = "would remove"
		}
		fmt.Printf("%s %d expired sessions\n", verb, len(res.Removed))
		for _, id := range res.Removed {
			fmt.Println(id)
		}
	}
	return nil
}

func parseRetention(ttlText, expiresText string) (time.Duration, time.Time, error) {
	if expiresText != "" {
		if t, err := time.Parse(time.RFC3339, expiresText); err == nil {
			return 0, t, nil
		}
		if t, err := time.Parse("2006-01-02", expiresText); err == nil {
			return 0, t, nil
		}
		return 0, time.Time{}, fmt.Errorf("invalid --expires-at %q", expiresText)
	}
	if ttlText == "" {
		return 0, time.Time{}, nil
	}
	if ttlText == "0" || ttlText == "none" || ttlText == "never" {
		return -1, time.Time{}, nil
	}
	if strings.HasSuffix(ttlText, "d") {
		daysText := strings.TrimSuffix(ttlText, "d")
		days, err := time.ParseDuration(daysText + "h")
		if err != nil {
			return 0, time.Time{}, err
		}
		return days * 24, time.Time{}, nil
	}
	d, err := time.ParseDuration(ttlText)
	return d, time.Time{}, err
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	return t.Format(time.RFC3339)
}

func preferredPutURLPath(paths []string) string {
	if len(paths) == 0 {
		return "index.html"
	}
	for _, p := range paths {
		if filepath.ToSlash(p) == "index.html" {
			return p
		}
	}
	return paths[0]
}

func preparePutContent(r io.Reader, name string, raw bool) (string, []byte, error) {
	return preparePutContentWithSecretScan(r, name, raw, false)
}

func preparePutContentWithSecretScan(r io.Reader, name string, raw bool, scan bool) (string, []byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, store.MaxFileSize+1))
	if err != nil {
		return "", nil, err
	}
	if int64(len(data)) > store.MaxFileSize {
		return "", nil, fmt.Errorf("file too large")
	}
	name = filepath.ToSlash(name)
	if scan {
		if err := secret.Check(name, data); err != nil {
			return "", nil, err
		}
	}
	if !raw && mdrender.IsMarkdownPath(name) {
		rendered, err := mdrender.Render(data, name)
		if err != nil {
			return "", nil, err
		}
		if scan {
			if err := secret.Check(mdrender.HTMLPath(name), rendered); err != nil {
				return "", nil, err
			}
		}
		return mdrender.HTMLPath(name), rendered, nil
	}
	return name, data, nil
}

func putPath(st *store.Store, sid, p string, interactive bool, raw bool) ([]string, error) {
	items, err := collectPutPathItems(p, raw, false)
	if err != nil {
		return nil, err
	}
	added := []string{}
	for _, item := range items {
		if _, err := st.Put(sid, item.name, bytes.NewReader(item.data), interactive); err != nil {
			return nil, err
		}
		added = append(added, item.name)
	}
	return added, nil
}

func collectPutPathItems(p string, raw bool, scan bool) ([]putItem, error) {
	info, e := os.Lstat(p)
	if e != nil {
		return nil, e
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("rejecting symlink: %s", p)
	}
	if info.IsDir() {
		items := []putItem{}
		err := filepath.WalkDir(p, func(path string, d os.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if d.Type()&os.ModeSymlink != 0 {
				return fmt.Errorf("rejecting symlink: %s", path)
			}
			if d.IsDir() {
				return nil
			}
			f, e := os.Open(path)
			if e != nil {
				return e
			}
			rel, _ := filepath.Rel(p, path)
			rel = filepath.ToSlash(rel)
			name, data, prepErr := preparePutContentWithSecretScan(f, rel, raw, scan)
			closeErr := f.Close()
			if prepErr != nil {
				return prepErr
			}
			items = append(items, putItem{name: name, data: data})
			return closeErr
		})
		return items, err
	}
	f, e := os.Open(p)
	if e != nil {
		return nil, e
	}
	name := filepath.Base(p)
	name, data, prepErr := preparePutContentWithSecretScan(f, name, raw, scan)
	closeErr := f.Close()
	if prepErr != nil {
		return nil, prepErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	return []putItem{{name: name, data: data}}, nil
}
