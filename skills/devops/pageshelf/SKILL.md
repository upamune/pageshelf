---
name: pageshelf
description: Use when publishing agent-generated HTML artifacts, long plans, research reports, PR explainers, or interactive review pages with the pageshelf CLI. Covers secure local/Tailscale serving, session handling, URL generation, and common pitfalls.
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

Pageshelf is a small Go CLI and secure static artifact server for agent-generated HTML. Use it when a response would be too long or visually dense for chat, Markdown, Telegram, or a PR comment. Instead of dumping a giant plan into the conversation, create an HTML artifact, put it into a pageshelf session, and return the generated URL.

The default workflow is intentionally low-friction for agents:

```bash
pageshelf put report.html
```

If no session is supplied, pageshelf creates one automatically and prints a tokenized `/a/<session>/index.html` URL. For sharing across the user's private network, run the server with Tailscale integration:

```bash
pageshelf serve --tailscale
```

Pageshelf is not a general public hosting platform. Treat it as a private artifact shelf for local or tailnet use.

## When to Use

Use pageshelf when:

- A plan, report, diff explanation, architecture review, or research synthesis is too long for chat.
- The user needs tables, SVG diagrams, color-coded sections, tabs, or interactive controls.
- Telegram or another messaging surface cannot render the information well.
- A coding/review agent needs to attach a browsable HTML explainer to a task or PR.
- You want the user to skim a visual artifact and keep the chat response short.

Do not use pageshelf when:

- The output is short enough to read directly in chat.
- The artifact contains secrets, private tokens, credentials, or raw sensitive data.
- The user needs a long-lived public website.
- You need server-side application behavior. Pageshelf serves static files only.

## Quick Start

### 1. Start the server locally

```bash
pageshelf serve
```

Default bind:

```text
127.0.0.1:8787
```

### 2. Publish one HTML file

```bash
pageshelf put index.html
```

Typical output:

```text
session: 20260510-0945-artifact-ifw1wa
url: http://127.0.0.1:8787/a/20260510-0945-artifact-ifw1wa/index.html?t=psr_...
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
pageshelf serve --host 127.0.0.1 --port 8787
pageshelf serve --host 0.0.0.0 --unsafe-public-bind
```

Behavior:

- Defaults to `127.0.0.1:8787`.
- `--tailscale` binds to a detected `100.64.0.0/10` Tailscale address.
- Public wildcard binds such as `0.0.0.0` and `::` require `--unsafe-public-bind`.
- Server has read/write/header timeouts and graceful shutdown.

Prefer `--tailscale` over `0.0.0.0` for private sharing.

### `put`

Publish files into a session. If `--session` is omitted, pageshelf creates a new session.

```bash
pageshelf put index.html
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

Interactive HTML:

```bash
pageshelf put --interactive index.html
```

Only use `--interactive` when the artifact needs local JavaScript for sliders, tabs, copy buttons, animations, or custom editors.

### `session`

```bash
pageshelf session create rate-limiter
pageshelf session create rate-limiter --json
pageshelf session info rate-limiter
pageshelf session info rate-limiter --json
pageshelf session rm rate-limiter
```

Most agent workflows do not need `session create`; `put` can auto-create sessions.

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

- tokenized read URLs
- token hash stored in manifest
- local token file stored separately
- default localhost bind
- explicit Tailscale bind
- wildcard bind requires `--unsafe-public-bind`
- path traversal prevention
- symlink rejection on input
- restrictive CSP
- `Cache-Control: no-store`
- `X-Content-Type-Options: nosniff`
- `Referrer-Policy: no-referrer`
- `X-Frame-Options: DENY`

### Safe vs Interactive CSP

Default safe mode disables scripts:

```http
script-src 'none'
connect-src 'none'
```

Interactive mode permits inline JavaScript but still blocks network fetches:

```http
script-src 'self' 'unsafe-inline'
connect-src 'none'
```

Use `--interactive` only for artifacts that need browser-side behavior.

### Sensitive data rule

Do not publish secrets, credentials, raw tokens, private keys, production `.env` files, or confidential dumps. Tokenized Tailscale URLs are convenient, not a replacement for data classification.

## HTML Authoring Guidance

For agent-created pages:

- Make the HTML self-contained where practical.
- Prefer inline CSS for portability.
- Use SVG for architecture and sequence diagrams.
- Use relative asset paths for multi-file artifacts.
- Include a short executive summary at the top.
- Add a sticky table of contents for long reports.
- Include copy buttons only when `--interactive` will be used.
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

1. **Forgetting the server endpoint in generated URLs.**
   If the server runs with `--tailscale`, generate URLs with `pageshelf put --tailscale ...` or `pageshelf url --tailscale ...`.

2. **Expecting `/a/<session>/...` to work without a token.**
   Artifact URLs require `?t=psr_...`.

3. **Putting a directory with absolute asset links.**
   HTML should use relative paths like `./assets/diagram.svg`, not `/tmp/artifact/assets/diagram.svg`.

4. **Using `--interactive` casually.**
   It allows inline JavaScript. Use safe mode unless the page needs client-side behavior.

5. **Publishing secrets because the server is “only Tailscale.”**
   Tailnet access is still access. Redact or summarize sensitive content first.

6. **Binding to `0.0.0.0` out of habit.**
   Use `--tailscale` for private sharing. Only use `--unsafe-public-bind` when you understand the exposure.

7. **Losing the session ID from auto-create output.**
   Use `--json` when an agent needs to parse and reuse the session ID.

8. **Expecting pageshelf to host an app backend.**
   It serves static files. Interactive artifacts must run fully in the browser.

## Verification Checklist

After publishing an artifact:

- [ ] `pageshelf serve` or `pageshelf serve --tailscale` is running.
- [ ] The URL path uses `/a/<session>/<file>` and includes `?t=psr_...`.
- [ ] The artifact opens successfully in a browser.
- [ ] Wrong or missing token returns `401`.
- [ ] Safe artifacts do not require JavaScript.
- [ ] Interactive artifacts were published with `--interactive` intentionally.
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
pageshelf put --interactive --tailscale index.html assets/
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
