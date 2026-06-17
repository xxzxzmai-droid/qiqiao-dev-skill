# qiqiao-dev skill

Codex skill for Qiqiao / 七巧 / 道一云低代码 development.

It helps build, debug, and package:

- Qiqiao custom pages using injected `index.html`, `index.css`, and `index.js`
- server-side custom function code with `var API = { ... }`
- `REST.API` / `applyApi` bridges
- page JS event extensions
- form/table/OpenAPI integration patterns
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
- `references/`: Qiqiao custom page, backend API, form/table, and delivery notes
- `assets/custom-page-injection/`: injected custom page template with frontend/backend diagnostics
- `scripts/check_qiqiao_page.py`: validates Qiqiao three-file delivery rules
- `scripts/make_injection_harness.py`: creates a local preview harness by injecting CSS/JS

## Validate

```bash
python3 scripts/check_qiqiao_page.py assets/custom-page-injection
python3 scripts/make_injection_harness.py assets/custom-page-injection /tmp/qiqiao-harness.html
```

The generated harness is only for local browser testing. In Qiqiao IDE, paste/upload the original `index.html`, `index.css`, `index.js`, and optional server code separately.
