# pageshelf

pageshelf v0.2 is a small, security-first Go CLI for publishing agent-generated HTML artifacts (plans, research notes, PR explainers) to a local or Tailscale-only web server.

Artifact URLs intentionally avoid the product name and use `/a/:session_id/:path`:

```text
http://100.x.y.z:8787/a/20260510-0924-artifact-7k3p/index.html?t=psr_xxx
```

## Install / build

```sh
go build ./cmd/pageshelf
```

## Usage

```sh
pageshelf serve [--tailscale] [--host HOST] [--port PORT] [--unsafe-public-bind]
pageshelf put [--session SESSION] [--stdin --name NAME | --content TEXT --name NAME | paths...] [--interactive] [--raw] [--tag TAG] [--ttl 14d] [--expires-at RFC3339|YYYY-MM-DD] [--json] [--host HOST --port PORT | --tailscale | --base-url URL]
pageshelf session create [name] [--tag TAG] [--ttl 14d] [--expires-at RFC3339|YYYY-MM-DD] [--json]
pageshelf session info <session> [--json]
pageshelf session meta <session> [--tag TAG | --clear-tags] [--ttl DURATION | --expires-at RFC3339|YYYY-MM-DD] [--json]
pageshelf session rm <session>
pageshelf list [--json]
pageshelf gc [--dry-run] [--json]
pageshelf files <session> [--json]
pageshelf url <session> [path] [--json] [--host HOST --port PORT | --tailscale | --base-url URL]
```

Storage defaults to `~/.local/share/pageshelf`; override with `--data-dir` or `PAGESHELF_DATA_DIR`.

## Agent skill

This repository ships a Hermes-style agent skill at:

```text
skills/devops/pageshelf/SKILL.md
```

Install or copy that skill into an agent skill directory when you want agents to know the preferred pageshelf workflow: create/publish HTML artifacts, use Tailscale URLs, avoid leaking secrets, and return short chat summaries instead of long Markdown dumps.

## Security model

- Server binds to `127.0.0.1` by default.
- `--tailscale` binds to a detected `100.64.0.0/10` address.
- Empty host, `0.0.0.0`, and `::` public binds are rejected unless `--unsafe-public-bind` is supplied.
- `put` returns a URL for the artifact it actually added: a single `foo.html` returns `foo.html`; multi-file uploads prefer `index.html` when present, otherwise the first added file.
- `put` renders Markdown inputs by default: `report.md` becomes `report.html`, `index.md` becomes `index.html`, and the returned URL points at the generated HTML. Use `--raw` to serve `.md`/`.markdown` files as literal Markdown files instead.
- `put` and `url` print localhost URLs by default, or can generate URLs with `--host/--port`, `--tailscale`, or `--base-url`.
- With `--tailscale`, URL generation prefers the local MagicDNS name from `hostname -f` or `tailscale status --json` (for example `omarchy-1.tailaf73.ts.net`) and falls back to the raw Tailscale IP.
- Each session has a random `psr_...` read token. The manifest stores only a SHA-256 hash; the plaintext token is stored in a local `0600` token file so the CLI can later reconstruct URLs.
- Each session manifest includes metadata: `created_at`, `updated_at`, `expires_at`, and optional normalized `tags`.
- New sessions default to `--ttl 14d`. Use `--ttl 0` for no expiry, `--expires-at` for a fixed deletion-eligible timestamp, and `pageshelf gc` to delete expired sessions later.
- Artifact paths reject absolute paths, `..`, backslashes and NUL bytes.
- Symlink inputs are rejected.
- CSP safe mode uses `default-src 'none'` and `script-src 'none'`. `--interactive` permits self/inline scripts but keeps `connect-src 'none'`.
- Security headers include no-store cache control, DENY framing, nosniff, no-referrer, COOP, CORP, and Permissions-Policy.

## License

MIT
