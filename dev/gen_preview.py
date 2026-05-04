import argparse
import json
import re
from pathlib import Path

root = Path(__file__).resolve().parent.parent
dev = root / "dev"


def extract_go_template(go_file: Path, const_name: str) -> str:
    p = go_file.read_text(encoding="utf-8")
    m = re.search(rf"const {re.escape(const_name)} = `(.+)`\s*\Z", p, re.S)
    if not m:
        raise SystemExit(f"could not parse {go_file.name} ({const_name})")
    html = m.group(1)
    if not html.startswith("<!DOCTYPE"):
        raise SystemExit(f"{const_name} did not start with <!DOCTYPE: {html[:40]!r}")
    return html


def write_root_preview() -> None:
    html = extract_go_template(root / "root_html.go", "rootHTMLTemplate")

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
    html = html.replace("__GETIP_DUAL_V4_URL__", json.dumps(""))
    html = html.replace("__GETIP_DUAL_V6_URL__", json.dumps(""))
    banner = (
        "<!-- Mock preview: see documentation/browser-preview.md | regenerate: python dev/gen_preview.py -->\n"
    )
    out = dev / "preview.html"
    out.write_text(banner + html, encoding="utf-8", newline="\n")
    print(f"wrote {out.relative_to(root)} (mock from {src.relative_to(root)})")


def write_blocklists_preview() -> None:
    html = extract_go_template(root / "blocklists_html.go", "blocklistsHTMLTemplate")

    mock_path = dev / "mock-blocklists.json"
    example_path = dev / "mock-blocklists.json.example"
    if mock_path.is_file():
        src = mock_path
    elif example_path.is_file():
        src = example_path
    else:
        raise SystemExit(
            "missing dev/mock-blocklists.json.example (and no dev/mock-blocklists.json). "
            "Restore mock-blocklists.json.example from the repo."
        )

    data = json.loads(src.read_text(encoding="utf-8"))
    mock_js = json.dumps(data, ensure_ascii=False, indent=2)
    mock = f"<script>\nwindow.__GETIP_BLOCKLISTS_MOCK__ = {mock_js};\n</script>\n"

    primary = data.get("ipv4") or data.get("ipv6") or "127.0.0.1"
    if not isinstance(primary, str):
        primary = "127.0.0.1"

    html = html.replace(
        "<title>Blocklists · get-ip</title>",
        "<title>Blocklists · get-ip · mock preview</title>\n" + mock,
    )
    html = html.replace("__PRIMARY_IP__", primary)
    banner = (
        "<!-- Mock preview: blocklists — see documentation/browser-preview.md | regenerate: python dev/gen_preview.py -->\n"
    )
    out = dev / "preview-blocklists.html"
    out.write_text(banner + html, encoding="utf-8", newline="\n")
    print(f"wrote {out.relative_to(root)} (mock from {src.relative_to(root)})")


def main() -> None:
    ap = argparse.ArgumentParser(description="Generate static HTML previews with injected JSON mocks.")
    ap.add_argument(
        "--only",
        choices=("all", "root", "blocklists"),
        default="all",
        help="which preview to generate (default: all)",
    )
    args = ap.parse_args()

    if args.only in ("all", "root"):
        write_root_preview()
    if args.only in ("all", "blocklists"):
        write_blocklists_preview()


if __name__ == "__main__":
    main()
