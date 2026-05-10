package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kong"
	mdrender "github.com/serizawa/pageshelf/internal/markdown"
	"github.com/serizawa/pageshelf/internal/server"
	"github.com/serizawa/pageshelf/internal/store"
	"github.com/serizawa/pageshelf/internal/tailscale"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type CLI struct {
	DataDir string     `help:"Storage directory." env:"PAGESHELF_DATA_DIR"`
	Serve   ServeCmd   `cmd:""`
	Put     PutCmd     `cmd:""`
	Session SessionCmd `cmd:""`
	List    ListCmd    `cmd:""`
	Files   FilesCmd   `cmd:""`
	URL     URLCmd     `cmd:""`
}
type Ctx struct{ Store *store.Store }
type ServeCmd struct {
	Tailscale        bool   `help:"Bind to detected Tailscale IP."`
	Host             string `default:"127.0.0.1"`
	Port             int    `default:"8787"`
	UnsafePublicBind bool
}
type PutCmd struct {
	Session     string `short:"s"`
	Stdin       bool
	Name        string
	Content     string
	Interactive bool
	Raw         bool `help:"Store Markdown files as-is instead of rendering .md/.markdown to HTML."`
	JSON        bool
	Host        string   `default:"127.0.0.1" help:"Host to use when printing the artifact URL."`
	Port        int      `default:"8787" help:"Port to use when printing the artifact URL."`
	Tailscale   bool     `help:"Use detected Tailscale IP when printing the artifact URL."`
	BaseURL     string   `name:"base-url" help:"Base URL to use when printing the artifact URL."`
	Paths       []string `arg:"" optional:"" name:"paths"`
}
type SessionCmd struct {
	Create SessionCreateCmd `cmd:""`
	Info   SessionInfoCmd   `cmd:""`
	Rm     SessionRmCmd     `cmd:""`
}
type SessionCreateCmd struct {
	Name string `arg:"" optional:""`
	JSON bool
}
type SessionInfoCmd struct {
	Session string `arg:""`
	JSON    bool
}
type SessionRmCmd struct {
	Session string `arg:""`
}
type ListCmd struct{ JSON bool }
type FilesCmd struct {
	Session string `arg:""`
	JSON    bool
}
type URLCmd struct {
	Session   string `arg:""`
	Path      string `arg:"" optional:"" default:"index.html"`
	JSON      bool
	Host      string `default:"127.0.0.1" help:"Host to use in the URL."`
	Port      int    `default:"8787" help:"Port to use in the URL."`
	Tailscale bool   `help:"Use detected Tailscale IP in the URL."`
	BaseURL   string `name:"base-url" help:"Base URL to use instead of host/port."`
}

func main() {
	var cli CLI
	k := kong.Parse(&cli, kong.Name("pageshelf"), kong.Description("agent-generated HTML artifact shelf v"+store.Version))
	st, err := store.New(cli.DataDir)
	k.FatalIfErrorf(err)
	err = k.Run(&Ctx{Store: st})
	k.FatalIfErrorf(err)
}
func printJSON(v any) { b, _ := json.MarshalIndent(v, "", "  "); fmt.Println(string(b)) }
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
	return server.ListenAndServe(net.JoinHostPort(h, fmt.Sprint(c.Port)), ctx.Store)
}
func (c *SessionCreateCmd) Run(ctx *Ctx) error {
	m, t, e := ctx.Store.Create(c.Name, c.Name, false)
	if e != nil {
		return e
	}
	out := map[string]any{"session": m, "token": t}
	if c.JSON {
		printJSON(out)
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
		fmt.Printf("%s (%d files)\n", m.ID, len(m.Files))
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
			fmt.Printf("%s\t%d files\n", m.ID, len(m.Files))
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
	tok, e := ctx.Store.ReadToken(c.Session)
	if e != nil {
		return e
	}
	b, e := publicBaseURL(c.Host, c.Port, c.Tailscale, c.BaseURL)
	if e != nil {
		return e
	}
	u := store.URL(b, c.Session, c.Path, tok)
	if c.JSON {
		printJSON(map[string]string{"url": u})
	} else {
		fmt.Println(u)
	}
	return nil
}
func (c *PutCmd) Run(ctx *Ctx) error {
	sid := c.Session
	slug := c.Name
	if slug == "" && len(c.Paths) > 0 {
		slug = c.Paths[0]
	}
	var tok string
	if sid == "" {
		m, t, e := ctx.Store.Create("", slug, c.Interactive)
		if e != nil {
			return e
		}
		sid = m.ID
		tok = t
	} else {
		var e error
		tok, e = ctx.Store.ReadToken(sid)
		if e != nil {
			return e
		}
	}
	added := []string{}
	if c.Stdin {
		if c.Name == "" {
			return fmt.Errorf("--stdin requires --name")
		}
		name, data, e := preparePutContent(os.Stdin, c.Name, c.Raw)
		if e != nil {
			return e
		}
		if _, e := ctx.Store.Put(sid, name, bytes.NewReader(data), c.Interactive); e != nil {
			return e
		}
		added = append(added, name)
	}
	if c.Content != "" {
		if c.Name == "" {
			return fmt.Errorf("--content requires --name")
		}
		name, data, e := preparePutContent(strings.NewReader(c.Content), c.Name, c.Raw)
		if e != nil {
			return e
		}
		if _, e := ctx.Store.Put(sid, name, bytes.NewReader(data), c.Interactive); e != nil {
			return e
		}
		added = append(added, name)
	}
	for _, p := range c.Paths {
		paths, e := putPath(ctx.Store, sid, p, c.Interactive, c.Raw)
		if e != nil {
			return e
		}
		added = append(added, paths...)
	}
	if !c.Stdin && c.Content == "" && len(c.Paths) == 0 {
		return fmt.Errorf("provide --stdin, --content, or paths")
	}
	b, e := publicBaseURL(c.Host, c.Port, c.Tailscale, c.BaseURL)
	if e != nil {
		return e
	}
	u := store.URL(b, sid, preferredPutURLPath(added), tok)
	if c.JSON {
		printJSON(map[string]string{"session": sid, "url": u})
	} else {
		fmt.Printf("session: %s\nurl: %s\n", sid, u)
	}
	return nil
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
	data, err := io.ReadAll(io.LimitReader(r, store.MaxFileSize+1))
	if err != nil {
		return "", nil, err
	}
	if int64(len(data)) > store.MaxFileSize {
		return "", nil, fmt.Errorf("file too large")
	}
	name = filepath.ToSlash(name)
	if !raw && mdrender.IsMarkdownPath(name) {
		rendered, err := mdrender.Render(data, name)
		if err != nil {
			return "", nil, err
		}
		return mdrender.HTMLPath(name), rendered, nil
	}
	return name, data, nil
}

func putPath(st *store.Store, sid, p string, interactive bool, raw bool) ([]string, error) {
	info, e := os.Lstat(p)
	if e != nil {
		return nil, e
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("rejecting symlink: %s", p)
	}
	if info.IsDir() {
		added := []string{}
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
			name, data, prepErr := preparePutContent(f, rel, raw)
			closeErr := f.Close()
			if prepErr != nil {
				return prepErr
			}
			if _, putErr := st.Put(sid, name, bytes.NewReader(data), interactive); putErr != nil {
				return putErr
			}
			added = append(added, name)
			return closeErr
		})
		return added, err
	}
	f, e := os.Open(p)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	name := filepath.Base(p)
	name, data, e := preparePutContent(f, name, raw)
	if e != nil {
		return nil, e
	}
	_, e = st.Put(sid, name, bytes.NewReader(data), interactive)
	return []string{name}, e
}
