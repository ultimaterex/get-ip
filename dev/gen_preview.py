import json
import re
from pathlib import Path

root = Path(__file__).resolve().parent.parent
dev = root / "dev"

p = (root / "root_html.go").read_text(encoding="utf-8")
m = re.search(r"const rootHTMLTemplate = `(.+)`\s*\Z", p, re.S)
if not m:
    raise SystemExit("could not parse root_html.go")
html = m.group(1)
assert html.startswith("<!DOCTYPE"), html[:40]

mock_path = dev / "mock.json"
example_path = dev / "mock.json.example"
if mock_path.is_file():
    src = mock_path
elif example_path.is_file():
    src = example_path
else:
    raise SystemExit(
        "missing dev/mock.json.example (and no dev/mock.json). "
        "Restore mock.json.example from the repo."
    )

data = json.loads(src.read_text(encoding="utf-8"))
mock_js = json.dumps(data, ensure_ascii=False, indent=2)
mock = f"<script>\nwindow.__GETIP_MOCK__ = {mock_js};\n</script>\n"

primary = data.get("ipv4") or data.get("ipv6") or "127.0.0.1"
if not isinstance(primary, str):
    primary = "127.0.0.1"

html = html.replace(
    "<title>Your IP</title>",
    "<title>Your IP · mock preview</title>\n" + mock,
)
html = html.replace("__PRIMARY_IP__", primary)
banner = (
    "<!-- Mock preview: see documentation/browser-preview.md | regenerate: python dev/gen_preview.py -->\n"
)
out = dev / "preview.html"
out.write_text(banner + html, encoding="utf-8", newline="\n")
print(f"wrote {out.relative_to(root)} (mock from {src.relative_to(root)})")
