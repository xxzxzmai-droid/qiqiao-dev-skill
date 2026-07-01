# qiqiao-dev skill

Codex skill for Qiqiao / 七巧 / 道一云低代码 development.

It helps build, debug, and package:

- official user-manual guidance for app, form, workflow, report, PC/mobile page, permission, integration, automation, and AI configuration
- low-code scripts, function-library usage, custom styles, page JS event extensions, and custom form/page components
- Qiqiao custom pages using injected `index.html`, `index.css`, and `index.js`
- server-side custom function code with `var API = { ... }`
- `REST.API` / `applyApi` bridges
- form/table/workflow/common OpenAPI integration patterns
- event push, workflow push, portal todo, unified message push, webhook receiver, and zero-code integration routing
- 数智彩虹 intranet OpenAPI token and form CRUD test tooling
- Linux/UOS OpenAPI probe and local CRUD web server executable source
- self-hosted API integration guidance

## Install

Clone this repository into the Codex skills directory as `qiqiao-dev`:

```bash
mkdir -p ~/.codex/skills
git clone https://github.com/xxzxzmai-droid/qiqiao-dev-skill.git ~/.codex/skills/qiqiao-dev
```

Restart or refresh Codex after installation if the skill list does not update immediately.

## Use

Ask Codex to use the Qiqiao skill, for example:

```text
用 qiqiao-dev skill 做一个七巧自定义页面三文件版
```

or:

```text
用七巧开发 skill 排查 index.css/index.js 没有注入的问题
```

## Contents

- `SKILL.md`: main skill instructions and trigger description
- `references/`: official-manual routing, Qiqiao custom page, backend API, form/table/component, OpenAPI, push integration, and delivery notes
- `assets/custom-page-injection/`: injected custom page template with frontend/backend diagnostics
- `assets/openapi-crud-custom-page/`: Qiqiao-deployed form CRUD test page with no durable frontend secret
- `assets/openapi-go-tool/`: Go source for UOS/Linux OpenAPI probe and local CRUD test web UI
- `scripts/check_qiqiao_page.py`: validates Qiqiao three-file delivery rules
- `scripts/make_injection_harness.py`: creates a local preview harness by injecting CSS/JS
- `scripts/check_public_skill.py`: validates public-skill concision and deployment-neutrality rules

## Validate

```bash
python3 scripts/check_qiqiao_page.py assets/custom-page-injection
python3 scripts/check_qiqiao_page.py assets/openapi-crud-custom-page
python3 scripts/make_injection_harness.py assets/custom-page-injection /tmp/qiqiao-harness.html
python3 scripts/check_public_skill.py
cd assets/openapi-go-tool && go test ./...
```

The generated harness is only for local browser testing. In Qiqiao IDE, paste/upload the original `index.html`, `index.css`, `index.js`, and optional server code separately.

Build the UOS test executable from `assets/openapi-go-tool`:

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o qiqiao-openapi-tool-uos-amd64 .
```

No OpenAPI base URL is hardcoded. Provide `baseUrl` in a private config file or pass `--base-url https://<qiqiao-host>/qiqiao/runtime/api/v1/bpms-integration`. Use `--insecure-skip-verify` only when the local certificate chain is unavailable.

Run a broader test against an existing test form:

```bash
./qiqiao-openapi-tool-uos-amd64 --config /path/to/qiqiao-config.local.json --mode full-smoke --page-size 5 --delete-after=true
```
