package geolookup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// MaxMind GeoLite2 direct downloads (HTTPS + Basic auth). See:
// https://dev.maxmind.com/geoip/updating-databases/
const (
	CityDownloadURL = "https://download.maxmind.com/geoip/databases/GeoLite2-City/download?suffix=tar.gz"
	ASNDownloadURL  = "https://download.maxmind.com/geoip/databases/GeoLite2-ASN/download?suffix=tar.gz"
	tarEntryCity    = "GeoLite2-City.mmdb"
	tarEntryASN     = "GeoLite2-ASN.mmdb"
)

func maxmindHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 15 * time.Minute,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			// MaxMind redirects to an R2 presigned URL. Basic auth must apply only to the
			// first request; forwarding Authorization breaks R2 (e.g. Missing x-amz-content-sha256).
			req.Header.Del("Authorization")
			return nil
		},
	}
}

// DownloadMMDB downloads a GeoLite2 tar.gz from MaxMind and extracts the named .mmdb into destPath.
func DownloadMMDB(ctx context.Context, downloadURL, destPath, mmdbEntryName, accountID, licenseKey string, logf func(string, ...interface{})) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadURL, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(accountID, licenseKey)

	resp, err := maxmindHTTPClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("maxmind http %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(destPath), "geolite-*.mmdb")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	ok := false
	defer func() {
		if !ok {
			_ = tmp.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	gr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gr.Close()

	tr := tar.NewReader(gr)
	var found bool
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if h.Typeflag != tar.TypeReg {
			continue
		}
		if !strings.HasSuffix(h.Name, mmdbEntryName) {
			continue
		}
		if _, err := io.Copy(tmp, tr); err != nil {
			return err
		}
		found = true
		break
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("tar.gz did not contain %s", mmdbEntryName)
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		return err
	}
	ok = true
	if logf != nil {
		logf("geolite: wrote %s", destPath)
	}
	return nil
}
