# latex-compiler-api

[![Docker Pulls](https://img.shields.io/docker/pulls/fellipessanha/latex-compiler-api)](https://hub.docker.com/r/fellipessanha/latex-compiler-api)
[![Go](https://img.shields.io/badge/go-1.25-blue)](https://go.dev)
[![License: MIT](https://img.shields.io/badge/license-MIT-green)](./LICENCE)

A self-hosted HTTP API that compiles LaTeX documents from a git repository or an uploaded archive and returns the result as PDF, PNG, or a compressed archive.

Designed to be run as short-lived, isolated Docker containers — spin one up per job, or keep one running behind a trusted network boundary.

> [!WARNING]
> **LaTeX is inherently unsafe.** A crafted `.tex` file can read arbitrary files from the container filesystem. Enabling `shell_escape` allows full arbitrary code execution. Never expose this service to untrusted users without understanding these risks. See [Security](#security).

---

## Quick start

```bash
docker run -d \
  -p 8080:8080 \
  -e API_SECRET=your-secret-here \
  fellipessanha/latex-compiler-api
```

The API is now live at `http://localhost:8080`.

Interactive API docs (Swagger UI) are available at **[http://localhost:8080/swagger/](http://localhost:8080/swagger/)** — browse and try every endpoint directly from the browser.

---

## Configuration

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP server port |
| `AUTH_PROVIDER` | `bearer` | Auth strategy: `bearer` or `none` |
| `API_SECRET` | — | Shared secret for `Authorization: Bearer <token>`. Required when `AUTH_PROVIDER=bearer`. |
| `COMPILE_TIMEOUT_SECONDS` | `60` | Maximum seconds allowed per compilation job |

> `AUTH_PROVIDER=none` disables authentication entirely. Only use this on a trusted private network.

---

## Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Liveness check — returns `200 OK` |
| `GET` | `/swagger/` | Interactive OpenAPI documentation |
| `POST` | `/git[/{output}]` | Compile from a git repository |
| `POST` | `/blob[/{output}]` | Compile from an uploaded archive |

The `{output}` path segment selects the response format. Omitting it defaults to `pdf`.

| Format | MIME type | Description |
|---|---|---|
| `pdf` | `application/pdf` | Compiled PDF (default) |
| `png` | `image/png` | First page rendered as PNG |
| `zip` | `application/zip` | All output files in a ZIP archive |
| `tar` | `application/x-tar` | All output files in a TAR archive |
| `7zip` | `application/x-7z-compressed` | All output files in a 7z archive |

---

## Request body

### `POST /git`

```json
{
  "url": "https://github.com/user/repo",
  "token": "ghp_...",
  "ref": "main",
  "entry": "main.tex",
  "target_dir": "paper",
  "compile_options": {},
  "output_options": {}
}
```

| Field | Required | Description |
|---|---|---|
| `url` | Yes | Repository URL (HTTPS) |
| `token` | No | Auth token embedded into the clone URL (for private repos) |
| `ref` | No | Branch, tag, or full 40-char commit SHA. Defaults to the default branch. |
| `entry` | No | Entry `.tex` file relative to the repo root (or `target_dir`). Defaults to `main.tex`. |
| `target_dir` | No | Subdirectory within the repository to compile from. Must be a relative path. |
| `compile_options` | No | See [Compile options](#compile-options). |
| `output_options` | No | See [Output options](#output-options). |

### `POST /blob`

Multipart form (`multipart/form-data`) with two parts:

| Part | Type | Description |
|---|---|---|
| `file` | file | Compressed source archive (ZIP, TAR, or 7zip) |
| `options` | string (JSON) | Optional. JSON-encoded options (see below). |

```json
{
  "entry": "main.tex",
  "compile_options": {},
  "output_options": {}
}
```

### Compile options

| Field | Type | Default | Description |
|---|---|---|---|
| `engine` | string | `pdflatex` | LaTeX engine: `pdflatex`, `xelatex`, or `lualatex` |
| `shell_escape` | bool | `false` | Enable `\write18`. **Significantly increases risk with untrusted input.** |
| `extra_flags` | string[] | `[]` | Extra flags appended to the `latexmk` invocation |
| `engine_args` | string[] | `[]` | Args injected into the engine invocation string |

### Output options

| Field | Type | Default | Description |
|---|---|---|---|
| `include_png` | bool | `false` | Add a `thumbnail.png` (page 1) to archive outputs. Ignored for `pdf` and `png`. |

---

## Examples

All examples assume the server is running at `http://localhost:8080` with `API_SECRET=secret`.

### Compile from a git repository

<details>
<summary>curl</summary>

```bash
curl -X POST http://localhost:8080/git \
  -H "Authorization: Bearer secret" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://github.com/user/my-paper","entry":"main.tex"}' \
  --output document.pdf
```

</details>

<details>
<summary>Python</summary>

```python
import requests

resp = requests.post(
    "http://localhost:8080/git",
    headers={"Authorization": "Bearer secret"},
    json={"url": "https://github.com/user/my-paper", "entry": "main.tex"},
)
resp.raise_for_status()
with open("document.pdf", "wb") as f:
    f.write(resp.content)
```

</details>

<details>
<summary>Elixir</summary>

```elixir
{:ok, %{body: pdf}} =
  Req.post("http://localhost:8080/git",
    headers: [{"authorization", "Bearer secret"}],
    json: %{url: "https://github.com/user/my-paper", entry: "main.tex"}
  )

File.write!("document.pdf", pdf)
```

</details>

<details>
<summary>Go</summary>

```go
package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
)

func main() {
	body, _ := json.Marshal(map[string]string{
		"url":   "https://github.com/user/my-paper",
		"entry": "main.tex",
	})
	req, _ := http.NewRequest("POST", "http://localhost:8080/git", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("Content-Type", "application/json")

	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	out, _ := os.Create("document.pdf")
	defer out.Close()
	io.Copy(out, resp.Body)
}
```

</details>

### Compile from an uploaded archive

<details>
<summary>curl</summary>

```bash
curl -X POST http://localhost:8080/blob/pdf \
  -H "Authorization: Bearer secret" \
  -F "file=@sources.zip" \
  -F 'options={"entry":"main.tex"}' \
  --output document.pdf
```

</details>

<details>
<summary>Python</summary>

```python
import requests

with open("sources.zip", "rb") as f:
    resp = requests.post(
        "http://localhost:8080/blob/pdf",
        headers={"Authorization": "Bearer secret"},
        files={"file": f},
        data={"options": '{"entry":"main.tex"}'},
    )
resp.raise_for_status()
with open("document.pdf", "wb") as f:
    f.write(resp.content)
```

</details>

<details>
<summary>Elixir</summary>

```elixir
pdf =
  Req.post!("http://localhost:8080/blob/pdf",
    headers: [{"authorization", "Bearer secret"}],
    form_multipart: [
      file: File.read!("sources.zip"),
      options: ~s({"entry":"main.tex"})
    ]
  ).body

File.write!("document.pdf", pdf)
```

</details>

<details>
<summary>Go</summary>

```go
package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

func main() {
	src, _ := os.ReadFile("sources.zip")

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", "sources.zip")
	fw.Write(src)
	mw.WriteField("options", `{"entry":"main.tex"}`)
	mw.Close()

	req, _ := http.NewRequest("POST", "http://localhost:8080/blob/pdf", &buf)
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("Content-Type", mw.FormDataContentType())

	resp, _ := http.DefaultClient.Do(req)
	defer resp.Body.Close()

	out, _ := os.Create("document.pdf")
	defer out.Close()
	io.Copy(out, resp.Body)
}
```

</details>

---

## Security

LaTeX is a full programming language with access to the container filesystem. A crafted document can:

- Read arbitrary files visible inside the container (e.g. `/etc/passwd`)
- With `shell_escape` enabled: execute arbitrary shell commands

**Recommended mitigations:**

- Keep `shell_escape: false` (the default) for any untrusted input.
- Set container resource limits: `--memory 1g --cpus 1`.
- Tune `COMPILE_TIMEOUT_SECONDS` to limit runaway jobs (default: 60 s).
- If you only use `/blob`, add `--network none` to block outbound connections from inside the container.
- Do not mount sensitive host paths into the container.
- Use `AUTH_PROVIDER=bearer` with a strong `API_SECRET` whenever the service is reachable over a network.

This service is designed to be **short-lived**: run it on-demand, compile, and tear it down. Avoid exposing it persistently to the public internet.

---

## Building from source

```bash
git clone https://github.com/fellipessanha/latex-compiler-api
cd latex-compiler-api
docker build -t latex-compiler-api .
```

Requires Docker. The Dockerfile handles everything — no local LaTeX installation needed.

---

## License

MIT — see [LICENCE](./LICENCE).
