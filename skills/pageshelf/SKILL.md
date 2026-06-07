---
name: pageshelf
description: Use when publishing agent-generated HTML artifacts, long plans, research reports, PR explainers, or interactive review pages with the pageshelf CLI. Covers private local/Tailscale serving, default interactive review/annotation UX, session handling, URL generation, and common pitfalls.
version: 1.0.0
author: Hermes Agent
license: MIT
metadata:
  hermes:
    tags: [html-artifacts, tailscale, cli, reports, agent-output, security]
    related_skills: [writing-plans, requesting-code-review]
---

# Pageshelf

## Overview

Pageshelf is a small Go CLI and private static artifact server for agent-generated HTML. Use it when a response would be too long or visually dense for chat, Markdown, Telegram, or a PR comment. Instead of dumping a giant plan into the conversation, create an HTML artifact, put it into a pageshelf session, and return the generated URL.

The default workflow is intentionally low-friction for agents:

```bash
pageshelf put report.html
```

If no session is supplied, pageshelf creates one automatically and prints a clean `/a/<session>/index.html` URL. For sharing across the user's private network, run the server with Tailscale integration:

```bash
pageshelf serve --tailscale
```

Pageshelf is not a general public hosting platform. Treat it as a private artifact shelf for local or tailnet use.

## Development Note

Pageshelf's built-in annotation UI is a normal in-repo Vite/TypeScript web app at `internal/runtime/annotation`, managed with aube. The app source is `index.html`, `src/main.ts`, supporting `src/*.ts` modules, `src/light.css`, and `src/shadow.css`; `make runtime-build` runs the Vite build into `internal/runtime/annotation/dist/` and regenerates bundled Go asset constants in `internal/server/annotation_runtime_generated.go`. Run it before testing or committing runtime changes so the binary serves the current `/_pageshelf/runtime/annotation/annotation.js` and `annotation.css` assets.

## When to Use

Use pageshelf when:

- A plan, report, diff explanation, architecture review, or research synthesis is too long for chat.
- The user needs tables, SVG diagrams, color-coded sections, tabs, interactive controls, or mobile-friendly review/annotation UI.
- Telegram or another messaging surface cannot render the information well.
- A coding/review agent needs to attach a browsable HTML explainer to a task or PR.
- You want the user to skim a visual artifact and keep the chat response short.

Do not use pageshelf when:

- The output is short enough to read directly in chat.
- The artifact contains secrets, private tokens, credentials, or raw sensitive data.
- The user needs a long-lived public website.
- You need server-side application behavior. Pageshelf serves static files only.

## Quick Start

### 1. Ensure a server is running

Prefer reusing the existing Pageshelf server. Do **not** start a fresh `pageshelf serve` every time you publish an artifact.

First check the expected endpoint:

```bash
curl -fsS http://127.0.0.1:8787/healthz >/dev/null || pageshelf serve
```

Default bind:

```text
127.0.0.1:8787
```

For Tailscale sharing, check the tailnet endpoint before starting another server:

```bash
TSIP=$(tailscale ip -4 | head -1)
curl -fsS "http://$TSIP:8787/healthz" >/dev/null || pageshelf serve --tailscale
```

If the server is already managed by systemd, prefer checking/restarting the service rather than spawning ad-hoc processes.

### 2. Publish one HTML file

```bash
pageshelf put index.html
```

Typical output:

```text
session: 20260510-0945-artifact-ifw1wa
url: http://127.0.0.1:8787/a/20260510-0945-artifact-ifw1wa/index.html
```

### 3. Use Tailscale for private network sharing

Start the server on the detected Tailscale IP:

```bash
pageshelf serve --tailscale
```

Generate Tailscale URLs from `put`:

```bash
pageshelf put --tailscale index.html
```

Or later:

```bash
pageshelf url --tailscale 20260510-0945-artifact-ifw1wa
```

## Core Commands

### `serve`

```bash
pageshelf serve
pageshelf serve --tailscale
pageshelf serve --tailscale --allow-image-src https://i.gyazo.com
pageshelf serve --host 127.0.0.1 --port 8787
pageshelf serve --host 0.0.0.0 --unsafe-public-bind
```

Behavior:

- Defaults to `127.0.0.1:8787`.
- `--tailscale` binds to a detected `100.64.0.0/10` Tailscale address.
- `--allow-image-src <origin>` appends explicit image origins to the served artifact CSP, for example `--allow-image-src https://i.gyazo.com`. Repeat the flag for multiple origins.
- Public wildcard binds such as `0.0.0.0` and `::` require `--unsafe-public-bind`.
- Server has read/write/header timeouts and graceful shutdown.

Prefer `--tailscale` over `0.0.0.0` for private sharing.

### `put`

Publish files into a session. If `--session` is omitted, pageshelf creates a new session.

```bash
pageshelf put index.html
pageshelf put report.md
pageshelf put --raw source.md
pageshelf put index.html assets/
pageshelf put -s pr-review index.html
pageshelf put -s pr-review ./artifact/
```

From stdin:

```bash
cat report.html | pageshelf put --stdin --name index.html
```

Inline content:

```bash
pageshelf put --content '<!doctype html><h1>Report</h1>' --name index.html
```

Machine-readable output:

```bash
pageshelf put --json index.html
```

Generate URL for a non-default server endpoint:

```bash
pageshelf put --host 100.88.12.34 --port 8787 index.html
pageshelf put --tailscale index.html
pageshelf put --base-url http://agent-box:8787 index.html
```

With `--tailscale`, pageshelf prefers the local MagicDNS name when available, e.g. `http://omarchy-1.tailaf73.ts.net:8787/...`, and falls back to the raw `100.x.y.z` Tailscale IP.

URL selection from `put`:

- A single file returns that file's URL, e.g. `pageshelf put foo.html` returns `/foo.html`.
- Markdown files render to HTML by default: `report.md` returns `/report.html`, `index.md` returns `/index.html`.
- Use `--raw` when you need to serve the Markdown file itself, e.g. `pageshelf put --raw report.md` returns `/report.md`.
- `--stdin --name index.html` and `--content ... --name index.html` return `/index.html`.
- `--stdin --name report.md` and `--content ... --name report.md` render Markdown unless `--raw` is supplied.
- Multi-file/directory uploads prefer `index.html` if present; otherwise they return the first added file.

Interactive HTML:

```bash
pageshelf put index.html
```

HTML artifacts are interactive by default. `--interactive` is deprecated and only accepted as a compatibility no-op; do not add it to new examples. Served HTML gets Pageshelf's built-in annotation runtime by default; use `pageshelf put --no-annotations ...` when a session should opt out. The runtime provides a mobile-friendly Review button, Shadow DOM bottom sheet/desktop panel, annotations stored in same-origin `localStorage` for the current URL path, text-selection anchors with quote/prefix/suffix/heading/path metadata, element picking, edit/delete/resolve controls, **Copy annotations**, **Copy JSON**, and **Export JSON**. Treat same-origin artifacts in a session as trusted because browser storage is same-origin. There is no automatic sending; users explicitly copy/paste notes when they want to hand them off.

### `session`

```bash
pageshelf session create rate-limiter
pageshelf session create rate-limiter --tag plan --tag backend --ttl 14d
pageshelf session create rate-limiter --expires-at 2026-06-01 --json
pageshelf session info rate-limiter
pageshelf session info rate-limiter --json
pageshelf session meta rate-limiter --tag pr-review --ttl 7d
pageshelf session meta rate-limiter --clear-tags --ttl 0
pageshelf session rm rate-limiter
pageshelf gc --dry-run
pageshelf gc
```

Most agent workflows do not need `session create`; `put` can auto-create sessions. New sessions get `created_at`, `updated_at`, and `expires_at` metadata automatically; default retention is 14 days. Add tags during creation with `--tag`, replace them later with `session meta --tag ...`, and remove expired sessions with `pageshelf gc`. Use `--ttl 0` only when the artifact should not expire automatically.

### `list`, `files`, and `url`

```bash
pageshelf list
pageshelf list --json

pageshelf files <session>
pageshelf files <session> --json

pageshelf url <session>
pageshelf url <session> diff.html
pageshelf url --tailscale <session>
pageshelf url --base-url http://agent-box:8787 <session>
```

Use `url` when you already added files and need to regenerate a share link.

## Agent Workflow Patterns

### Long planning output

1. Write a concise chat summary.
2. Generate `index.html` with the full plan, diagrams, code snippets, and risk sections.
3. Publish it:

```bash
pageshelf put --tailscale index.html
```

4. Return the URL and a 3-5 bullet summary in chat.


### Mobile-first review/annotation artifacts

For plans, PR explainers, and annotated diffs, make the HTML feel like a selection-first review surface:

- single-column readable layout on phones
- large tap targets for findings, checklist rows, and code/diff locations
- visible annotation/review panel by default when the artifact is for review
- generic **Copy annotations**, **Copy JSON**, and **Export JSON** actions for manual handoff or tool ingestion
- no automatic submission back to Hermes or another agent; the user explicitly pastes copied annotations

### PR explainer

Create a self-contained HTML file with:

- high-level summary
- annotated diff snippets
- severity-colored findings
- data/control-flow diagrams
- test checklist

Then publish:

```bash
pageshelf put --tailscale pr-review.html
```

If the filename is not `index.html`, either name it explicitly while writing or call:

```bash
pageshelf url --tailscale <session> pr-review.html
```

### Multi-file artifact

If HTML references local assets, put the whole directory:

```bash
pageshelf put --tailscale ./artifact/
```

Inside the HTML, use relative links:

```html
<img src="./assets/flow.svg" alt="Flow diagram">
<link rel="stylesheet" href="./style.css">
```

Pageshelf serves them under the same session route.

### Existing session append

When `put` creates a session, capture the returned session ID. Add follow-up files with `-s`:

```bash
pageshelf put --json index.html
pageshelf put -s 20260510-0945-artifact-ifw1wa data.json diagram.svg
```

Avoid relying on hidden current-session state. Pageshelf intentionally makes session choice explicit after the first auto-create.

## Security Model

Pageshelf is designed for private local/tailnet artifact serving, not public internet hosting.

Key protections:

- default localhost bind
- explicit Tailscale bind
- wildcard bind requires `--unsafe-public-bind`
- path traversal prevention
- symlink rejection on input
- tight private-artifact CSP that allows local inline annotation UI but blocks network sends
- `Cache-Control: no-store`
- `X-Content-Type-Options: nosniff`
- `Referrer-Policy: no-referrer`
- `X-Frame-Options: DENY`

### CSP for Private Interactive Artifacts

Pageshelf assumes artifacts are generated by the user or their agent for private review, but network sends remain blocked. Default served HTML CSP:

```http
default-src 'none'
script-src 'self' 'unsafe-inline'
style-src 'self' 'unsafe-inline'
img-src 'self' data: blob:
font-src 'self' data:
connect-src 'none'
object-src 'none'
base-uri 'none'
frame-ancestors 'none'
form-action 'none'
worker-src 'none'
child-src 'none'
frame-src 'none'
```

With `--no-annotations`, Pageshelf skips runtime injection and serves HTML with `script-src 'none'` and `connect-src 'none'`. Use `pageshelf serve --allow-image-src <origin>` to append explicit remote image origins to `img-src`, for example `https://i.gyazo.com`. Values must be origins; paths, query strings, fragments, wildcards, whitespace, and semicolons are rejected. This only relaxes image loading; `connect-src 'none'` remains unchanged. Artifact URLs are capability-free; private serving depends on localhost/Tailscale binding, path traversal checks, symlink rejection, no-store headers, safe CSP, and secret scanning.

### Sensitive data rule

Do not publish secrets, credentials, raw tokens, private keys, production `.env` files, or confidential dumps. Pageshelf URLs do not contain read tokens; local/Tailscale serving is not a replacement for data classification.

## HTML Authoring Guidance

For agent-created pages:

- Make the HTML self-contained where practical.
- Prefer inline CSS for portability.
- Use SVG for architecture and sequence diagrams.
- Use relative asset paths for multi-file artifacts.
- Include a short executive summary at the top.
- Add a sticky table of contents for long reports.
- Include copy buttons freely; HTML is interactive by default. For review pages, add a generic **Copy annotations** button that copies selected locations plus comments, not Hermes-specific prompts.
- Avoid external CDNs; CSP and tailnet usage make local assets more reliable.
- Do not embed third-party analytics, trackers, fonts, or remote scripts.

For Telegram handoff, keep the chat reply short:

```text
詳細HTML作った:
<url>

中身:
- architecture diagram
- security checklist
- implementation tasks
```

## Common Pitfalls

1. **Starting a new server for every `put`.**
   `put` only writes into the data directory and prints a URL; it does not require a fresh server. Check `/healthz` on the intended local/Tailscale endpoint first and reuse the running systemd/service process when available. Spawning ad-hoc servers leads to port drift (`8791`, `8792`, `8796`, ...), stale links, and confusion.

2. **Forgetting the server endpoint in generated URLs.**
   If the server runs with `--tailscale`, generate URLs with `pageshelf put --tailscale ...` or `pageshelf url --tailscale ...`.

3. **Assuming the URL itself is secret.**
   Artifact URLs are clean and capability-free. Rely on localhost/Tailscale/private bind assumptions and do not publish sensitive content.

4. **Putting a directory with absolute asset links.**
   HTML should use relative paths like `./assets/diagram.svg`, not `/tmp/artifact/assets/diagram.svg`.

5. **Treating `--interactive` as required.**
   It is deprecated/no-op. Publish normal HTML and design the artifact to be interactive by default when useful.

6. **Publishing secrets because the server is “only Tailscale.”**
   Tailnet access is still access. Redact or summarize sensitive content first.

7. **Binding to `0.0.0.0` out of habit.**
   Use `--tailscale` for private sharing. Only use `--unsafe-public-bind` when you understand the exposure.

8. **Losing the session ID from auto-create output.**
   Use `--json` when an agent needs to parse and reuse the session ID.

9. **Expecting pageshelf to host an app backend.**
   It serves static files. Interactive artifacts must run fully in the browser.

## Verification Checklist

After publishing an artifact:

- [ ] `pageshelf serve` or `pageshelf serve --tailscale` is running.
- [ ] The URL path uses clean `/a/<session>/<file>` form, with no token query.
- [ ] The artifact opens successfully in a browser.
- [ ] Mobile review controls are reachable and readable.
- [ ] Copy annotations is generic and copies locations plus comments for manual paste.
- [ ] No secrets or raw credentials are present in the HTML or asset files.
- [ ] Multi-file assets load via relative paths.
- [ ] Chat response includes only the short summary plus URL.

## One-Shot Recipes

### Publish a long report to Tailscale

```bash
pageshelf serve --tailscale
pageshelf put --tailscale report.html
```

### Publish a generated HTML string from an agent

```bash
cat <<'HTML' | pageshelf put --stdin --name index.html --tailscale
<!doctype html>
<html>
<head><meta charset="utf-8"><title>Artifact</title></head>
<body><h1>Artifact</h1></body>
</html>
HTML
```

### Publish an interactive tuning UI

```bash
pageshelf put --tailscale index.html assets/
```

### Regenerate a link for a different server address

```bash
pageshelf url --base-url http://100.88.12.34:8787 <session> index.html
```

### Inspect and clean up

```bash
pageshelf list
pageshelf files <session>
pageshelf session rm <session>
```
