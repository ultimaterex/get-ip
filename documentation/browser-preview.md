# Offline browser UI preview (mock data)

The HTML served at **`/`** (when `Accept` includes `text/html`) lives in **`root_html.go`** as `rootHTMLTemplate`. Editing that template is easier if you can view it without running the Go server or needing real GeoLite MMDBs.

## Files

| Path | Purpose |
|------|---------|
| **`dev/mock.json.example`** | Sample **`/json`**-shaped payload committed to the repo. Copy to **`dev/mock.json`** if you want your own data (see below). |
| **`dev/mock.json`** | Optional local file (**gitignored**) — same shape as **`mock.json.example`**. **`gen_preview.py`** prefers this file when present. |
| **`dev/preview.html`** | Generated static snapshot (also **gitignored**). Open in a desktop browser after running **`python dev/gen_preview.py`**. |
| **`dev/gen_preview.py`** | Reads **`dev/mock.json`** or falls back to **`dev/mock.json.example`**, parses **`root_html.go`**, writes **`dev/preview.html`** with **`window.__GETIP_MOCK__`** injected. |

## Setup

1. Optional: **`cp dev/mock.json.example dev/mock.json`** (Unix) or copy on Windows, then edit **`dev/mock.json`** with whatever IPs/geo you want to preview (file stays local / out of git).

2. Edit **`root_html.go`** as needed.

3. Regenerate the preview:

   ```bash
   python dev/gen_preview.py
   ```

4. Open **`dev/preview.html`** in Chrome, Edge, Firefox, etc.  
   - **Windows:** double‑click in Explorer or drag into a browser.

If **`dev/mock.json`** is missing, the script uses **`dev/mock.json.example`** so a fresh clone still works without an extra copy step.

## Limitations

- **Network:** Leaflet and map tiles load from **unpkg** and **OpenStreetMap**. Offline/air‑gapped machines won’t show the map unless you vendor those assets locally (not part of this repo).
- **Links:** **Plain report** and **Raw JSON** point to **`/all`** and **`/json`** on whatever origin the browser uses (`file://` has no server). Use a running **get-ip** instance for real endpoints.

## Related

- Production HTML behavior and **`/json`** loading are described in the root **[README](../README.md)** (browser vs plain text on **`/`**).
