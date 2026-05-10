package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/serizawa/pageshelf/internal/store"
	pageshelfskill "github.com/serizawa/pageshelf/skills/pageshelf"
)

const testSecret = "ghp_123456789012345678901234567890123456"

func newTestSession(t *testing.T) (*store.Store, string) {
	t.Helper()
	st, err := store.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m, _, err := st.Create("test", "test", false)
	if err != nil {
		t.Fatal(err)
	}
	return st, m.ID
}

func TestPreparePutContentRendersMarkdownByDefault(t *testing.T) {
	name, data, err := preparePutContent(strings.NewReader("# Demo\n\n- one\n- two\n"), "README.md", false)
	if err != nil {
		t.Fatal(err)
	}
	if name != "README.html" {
		t.Fatalf("name = %q, want README.html", name)
	}
	body := string(data)
	if !strings.Contains(body, "<title>Demo</title>") || !strings.Contains(body, "<li>one</li>") {
		t.Fatalf("rendered HTML missing expected content:\n%s", body)
	}
}

func TestPreparePutContentRawMarkdown(t *testing.T) {
	name, data, err := preparePutContent(strings.NewReader("# Demo\n"), "README.md", true)
	if err != nil {
		t.Fatal(err)
	}
	if name != "README.md" {
		t.Fatalf("name = %q, want README.md", name)
	}
	if string(data) != "# Demo\n" {
		t.Fatalf("data = %q", string(data))
	}
}

func TestPutPathRendersMarkdownAndReturnsHTMLPath(t *testing.T) {
	st, sid := newTestSession(t)
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "report.md")
	if err := os.WriteFile(mdPath, []byte("# Report\n\nhello"), 0o600); err != nil {
		t.Fatal(err)
	}
	added, err := putPath(st, sid, mdPath, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(added) != 1 || added[0] != "report.html" {
		t.Fatalf("added = %#v", added)
	}
	m, err := st.Load(sid)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Files) != 1 || m.Files[0].Path != "report.html" || m.Files[0].MIME != "text/html; charset=utf-8" {
		t.Fatalf("manifest files = %#v", m.Files)
	}
	f, _, _, err := st.Open(sid, "report.html")
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(f.Name())
	closeErr := f.Close()
	if err != nil {
		t.Fatal(err)
	}
	if closeErr != nil {
		t.Fatal(closeErr)
	}
	if !strings.Contains(string(b), "<h1 id=\"report\">Report</h1>") {
		t.Fatalf("rendered file missing heading:\n%s", string(b))
	}
}

func TestPutPathRawMarkdownKeepsMDPath(t *testing.T) {
	st, sid := newTestSession(t)
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "report.md")
	if err := os.WriteFile(mdPath, []byte("# Report\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	added, err := putPath(st, sid, mdPath, false, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(added) != 1 || added[0] != "report.md" {
		t.Fatalf("added = %#v", added)
	}
	m, err := st.Load(sid)
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Files) != 1 || m.Files[0].Path != "report.md" {
		t.Fatalf("manifest files = %#v", m.Files)
	}
}

func TestPutSecretScanDefaultBlocksFileBeforeSessionCreate(t *testing.T) {
	st, err := store.New(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	secretPath := filepath.Join(dir, "secret.txt")
	if err := os.WriteFile(secretPath, []byte("key="+testSecret+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	err = (&PutCmd{Paths: []string{secretPath}}).Run(&Ctx{Store: st})
	if err == nil || !strings.Contains(err.Error(), "secret scan blocked") || !strings.Contains(err.Error(), "gitleaks detected") {
		t.Fatalf("err = %v, want gitleaks secret scan blocked", err)
	}
	if !strings.Contains(err.Error(), "secret.txt:1") {
		t.Fatalf("err = %v, want file and line in error", err)
	}
	sessions, err := st.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(sessions) != 0 {
		t.Fatalf("sessions = %#v, want no empty session", sessions)
	}
}

func TestPutNoSecretScanAllowsSecretFile(t *testing.T) {
	st, sid := newTestSession(t)
	dir := t.TempDir()
	secretPath := filepath.Join(dir, "secret.txt")
	if err := os.WriteFile(secretPath, []byte("key="+testSecret+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	items, err := (&PutCmd{NoSecretScan: true, Paths: []string{secretPath}}).collectPutItems()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].name != "secret.txt" {
		t.Fatalf("items = %#v", items)
	}
	if _, err := st.Put(sid, items[0].name, strings.NewReader(string(items[0].data)), false); err != nil {
		t.Fatal(err)
	}
}

func TestPutSecretScanBlocksNestedDirectorySecret(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "nested")
	if err := os.Mkdir(nested, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "secret.md"), []byte("# Leak\n\n"+testSecret+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := (&PutCmd{Paths: []string{dir}}).collectPutItems()
	if err == nil || !strings.Contains(err.Error(), "secret scan blocked nested/secret.md") {
		t.Fatalf("err = %v, want nested secret blocked", err)
	}
}

func TestPutSecretScanBlocksContent(t *testing.T) {
	_, err := (&PutCmd{Name: "note.txt", Content: "aws_access_key_id = " + testSecret}).collectPutItems()
	if err == nil || !strings.Contains(err.Error(), "secret scan blocked") {
		t.Fatalf("err = %v, want content secret blocked", err)
	}
}

func TestPutSecretScanBlocksStdin(t *testing.T) {
	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()
	if _, err := w.WriteString("aws_access_key_id = " + testSecret); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	_, err = (&PutCmd{Name: "stdin.txt", Stdin: true}).collectPutItems()
	if err == nil || !strings.Contains(err.Error(), "secret scan blocked") {
		t.Fatalf("err = %v, want stdin secret blocked", err)
	}
}

func TestSkillShowCommandPrintsBundledSkill(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	err = (&SkillShowCmd{}).Run(&Ctx{})
	if closeErr := w.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	os.Stdout = oldStdout
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != string(pageshelfskill.Content()) {
		t.Fatalf("skill show output did not match bundled skill")
	}
}

func TestSkillHelpCommandPrintsHelp(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	err = (&SkillHelpCmd{}).Run(&Ctx{})
	if closeErr := w.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	os.Stdout = oldStdout
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	out := string(b)
	if !strings.Contains(out, "Usage: pageshelf skill <command>") || !strings.Contains(out, "pageshelf skill show") {
		t.Fatalf("skill help output = %q", out)
	}
}

func TestSkillInstallCommandWritesBundledSkill(t *testing.T) {
	dir := t.TempDir()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	err = (&SkillInstallCmd{Dir: dir}).Run(&Ctx{})
	if closeErr := w.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	os.Stdout = oldStdout
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	path := strings.TrimSpace(string(b))
	if path != filepath.Join(dir, "SKILL.md") {
		t.Fatalf("install output path = %q", path)
	}
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(written) != string(pageshelfskill.Content()) {
		t.Fatalf("installed skill did not match bundled skill")
	}
}

func TestVersionCommandPrintsStoreVersion(t *testing.T) {
	var out strings.Builder
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	err = (&VersionCmd{}).Run(&Ctx{})
	if closeErr := w.Close(); closeErr != nil {
		t.Fatal(closeErr)
	}
	os.Stdout = oldStdout
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	out.Write(b)
	if strings.TrimSpace(out.String()) != store.Version {
		t.Fatalf("version output = %q, want %q", out.String(), store.Version)
	}
}
