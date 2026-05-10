package main

import (
	"encoding/json"
	"fmt"
	"github.com/alecthomas/kong"
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
	if c.Stdin {
		if c.Name == "" {
			return fmt.Errorf("--stdin requires --name")
		}
		if _, e := ctx.Store.Put(sid, c.Name, os.Stdin, c.Interactive); e != nil {
			return e
		}
	}
	if c.Content != "" {
		if c.Name == "" {
			return fmt.Errorf("--content requires --name")
		}
		if _, e := ctx.Store.Put(sid, c.Name, strings.NewReader(c.Content), c.Interactive); e != nil {
			return e
		}
	}
	for _, p := range c.Paths {
		if e := putPath(ctx.Store, sid, p, c.Interactive); e != nil {
			return e
		}
	}
	if !c.Stdin && c.Content == "" && len(c.Paths) == 0 {
		return fmt.Errorf("provide --stdin, --content, or paths")
	}
	b, e := publicBaseURL(c.Host, c.Port, c.Tailscale, c.BaseURL)
	if e != nil {
		return e
	}
	u := store.URL(b, sid, "index.html", tok)
	if c.JSON {
		printJSON(map[string]string{"session": sid, "url": u})
	} else {
		fmt.Printf("session: %s\nurl: %s\n", sid, u)
	}
	return nil
}
func putPath(st *store.Store, sid, p string, interactive bool) error {
	info, e := os.Lstat(p)
	if e != nil {
		return e
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("rejecting symlink: %s", p)
	}
	if info.IsDir() {
		return filepath.WalkDir(p, func(path string, d os.DirEntry, e error) error {
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
			_, putErr := st.Put(sid, rel, f, interactive)
			closeErr := f.Close()
			if putErr != nil {
				return putErr
			}
			return closeErr
		})
	}
	f, e := os.Open(p)
	if e != nil {
		return e
	}
	defer f.Close()
	name := filepath.Base(p)
	_, e = st.Put(sid, name, io.Reader(f), interactive)
	return e
}
