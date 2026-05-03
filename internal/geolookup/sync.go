package geolookup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

func maxmindCreds() (accountID, licenseKey string) {
	accountID = strings.TrimSpace(os.Getenv("MAXMIND_ACCOUNT_ID"))
	licenseKey = strings.TrimSpace(os.Getenv("MAXMIND_LICENSE_KEY"))
	return accountID, licenseKey
}

// MaxAgeFromEnv returns GEOLITE_MAX_AGE_DAYS as a duration, defaulting to 7 days.
func MaxAgeFromEnv() time.Duration {
	s := strings.TrimSpace(os.Getenv("GEOLITE_MAX_AGE_DAYS"))
	if s == "" {
		return 7 * 24 * time.Hour
	}
	var n int
	if _, err := fmt.Sscanf(s, "%d", &n); err != nil || n < 1 {
		return 7 * 24 * time.Hour
	}
	return time.Duration(n) * 24 * time.Hour
}

func shouldRefresh(path string, maxAge time.Duration) (bool, string) {
	st, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return true, "missing file"
		}
		return false, ""
	}
	if time.Since(st.ModTime()) > maxAge {
		return true, "stale"
	}
	return false, ""
}

// SyncIfConfigured downloads GeoLite MMDBs when MAXMIND_ACCOUNT_ID and MAXMIND_LICENSE_KEY
// are set and files are missing or stale. Download errors are logged; the caller may still open existing files.
func SyncIfConfigured(ctx context.Context, logf func(string, ...interface{})) {
	cityPath := CityMMDBPath()
	asnPath := ASNMMDBPath()
	maxAge := MaxAgeFromEnv()
	acc, key := maxmindCreds()

	if acc == "" || key == "" {
		logf("geolite: MAXMIND_ACCOUNT_ID / MAXMIND_LICENSE_KEY not set; only using local MMDB files if present")
		return
	}

	if need, reason := shouldRefresh(cityPath, maxAge); need {
		logf("geolite: updating city database (%s)", reason)
		if err := DownloadMMDB(ctx, CityDownloadURL, cityPath, tarEntryCity, acc, key, logf); err != nil {
			logf("geolite: city download failed: %v (using existing file if any)", err)
		}
	}
	if need, reason := shouldRefresh(asnPath, maxAge); need {
		logf("geolite: updating ASN database (%s)", reason)
		if err := DownloadMMDB(ctx, ASNDownloadURL, asnPath, tarEntryASN, acc, key, logf); err != nil {
			logf("geolite: asn download failed: %v (using existing file if any)", err)
		}
	}
}

// Fetch downloads or refreshes GeoLite MMDBs when needed. It requires MAXMIND_ACCOUNT_ID and
// MAXMIND_LICENSE_KEY. Returns a joined error if any required download fails.
func Fetch(ctx context.Context, logf func(string, ...interface{})) error {
	if logf == nil {
		logf = func(string, ...interface{}) {}
	}
	acc, key := maxmindCreds()
	if acc == "" || key == "" {
		return fmt.Errorf("MAXMIND_ACCOUNT_ID and MAXMIND_LICENSE_KEY must be set")
	}

	cityPath := CityMMDBPath()
	asnPath := ASNMMDBPath()
	maxAge := MaxAgeFromEnv()

	var errs []error
	didWork := false

	if need, reason := shouldRefresh(cityPath, maxAge); need {
		didWork = true
		logf("geolite: updating city database (%s)", reason)
		if err := DownloadMMDB(ctx, CityDownloadURL, cityPath, tarEntryCity, acc, key, logf); err != nil {
			errs = append(errs, fmt.Errorf("city download: %w", err))
		}
	}
	if need, reason := shouldRefresh(asnPath, maxAge); need {
		didWork = true
		logf("geolite: updating ASN database (%s)", reason)
		if err := DownloadMMDB(ctx, ASNDownloadURL, asnPath, tarEntryASN, acc, key, logf); err != nil {
			errs = append(errs, fmt.Errorf("asn download: %w", err))
		}
	}
	if !didWork && len(errs) == 0 {
		logf("geolite: databases up to date")
	}

	return errors.Join(errs...)
}
