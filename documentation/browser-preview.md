# Offline browser UI preview (mock data)

The HTML served at **`/`** (when `Accept` includes `text/html`) lives in **`root_html.go`** as `rootHTMLTemplate`. Editing that template is easier if you can view it without running the Go server or needing real GeoLite MMDBs.

## Files

| Path | Purpose |
|------|---------|
| **`dev/mock.json.example`** | Sample **`/json`**-shaped payload committed to the repo. Copy to **`dev/mock.json`** if you want your own data (see below). |
| **`dev/mock.json`** | Optional local file (**gitignored**) — same shape as **`mock.json.example`**. **`gen_preview.py`** prefers this file when present. |
| **`dev/mock-blocklists.json.example`** | Sample **`/blocklists/json`** payload (prefix **`blocklists`** + **`dnsbl`**). Copy to **`dev/mock-blocklists.json`** to customize. |
| **`dev/mock-blocklists.json`** | Optional local file (**gitignored**). Preferred over **`mock-blocklists.json.example`** when present. |
| **`dev/preview.html`** | Generated static snapshot for **`/`** (also **gitignored**). |
| **`dev/preview-blocklists.html`** | Generated static snapshot for **`/blocklists`** (also **gitignored**). |
| **`dev/gen_preview.py`** | Parses templates from **`root_html.go`** and **`blocklists_html.go`**, injects **`window.__GETIP_MOCK__`** / **`window.__GETIP_BLOCKLISTS_MOCK__`**, writes both previews (use **`--only root`** or **`--only blocklists`** for one). |

## Setup

1. Optional: copy mocks for local edits:
   - **`cp dev/mock.json.example dev/mock.json`** (home page **`/`**)
   - **`cp dev/mock-blocklists.json.example dev/mock-blocklists.json`** (**`/blocklists`**)

2. Edit **`root_html.go`** / **`blocklists_html.go`** as needed.

3. Regenerate previews:

   ```bash
   python dev/gen_preview.py
   ```

   One page only: **`python dev/gen_preview.py --only root`** or **`--only blocklists`**.

4. Open **`dev/preview.html`** (home) and/or **`dev/preview-blocklists.html`** (blocklists) in Chrome, Edge, Firefox, etc.  
   - **Windows:** double‑click in Explorer or drag into a browser.

If **`dev/mock.json`** or **`dev/mock-blocklists.json`** is missing, the script falls back to the matching **`.example`** file so a fresh clone works without copies.

## Limitations

- **Network:** Leaflet and map tiles load from **unpkg** and **OpenStreetMap**. Offline/air‑gapped machines won’t show the map unless you vendor those assets locally (not part of this repo).
- **Links:** **Plain report** and **Raw JSON** point to **`/all`**, **`/json`**, **`/blocklists/all`**, **`/blocklists/json`** on whatever origin the browser uses (`file://` has no server). Use a running **get-ip** instance for real endpoints.

## Related

- Production HTML behavior and **`/json`** loading are described in the root **[README](../README.md)** (browser vs plain text on **`/`**).
