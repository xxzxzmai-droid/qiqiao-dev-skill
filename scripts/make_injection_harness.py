#!/usr/bin/env python3
import argparse
from pathlib import Path


def read_text(path):
    return path.read_text(encoding="utf-8", errors="replace") if path.exists() else ""


def main():
    parser = argparse.ArgumentParser(description="Create a local preview harness by injecting index.css and index.js into index.html.")
    parser.add_argument("folder", help="Folder containing index.html, index.css, and index.js")
    parser.add_argument("output", help="Output HTML path")
    args = parser.parse_args()

    root = Path(args.folder).expanduser().resolve()
    output = Path(args.output).expanduser().resolve()
    html = read_text(root / "index.html")
    css = read_text(root / "index.css")
    js = read_text(root / "index.js")

    if not html:
        raise SystemExit("index.html is missing or empty")

    style = "\n<style data-qiqiao-harness=\"index.css\">\n" + css + "\n</style>\n"
    script = "\n<script data-qiqiao-harness=\"index.js\">\n" + js.replace("</script>", "<\\/script>") + "\n</script>\n"

    if "</head>" in html:
        html = html.replace("</head>", style + "</head>", 1)
    else:
        html = style + html

    if "</body>" in html:
        html = html.replace("</body>", script + "</body>", 1)
    else:
        html = html + script

    output.parent.mkdir(parents=True, exist_ok=True)
    output.write_text(html, encoding="utf-8")
    print(str(output))


if __name__ == "__main__":
    raise SystemExit(main())
