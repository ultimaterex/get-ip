package main

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/oschwald/geoip2-golang/v2"
)

// MaxMind GeoLite2 direct download (HTTPS + Basic auth). See:
// https://dev.maxmind.com/geoip/updating-databases/
const geoliteCityDownloadURL = "https://download.maxmind.com/geoip/databases/GeoLite2-City/download?suffix=tar.gz"

var (
	geoMu sync.RWMutex
	geoDB *geoip2.Reader
)

// geoRecord is included in /all and /json when GeoLite2-City data is available.
type geoRecord struct {
	City          string `json:"city,omitempty"`
	Region        string `json:"region,omitempty"`
	RegionISO     string `json:"region_iso,omitempty"`
	Country       string `json:"country,omitempty"`
	CountryName   string `json:"country_name,omitempty"`
	Continent     string `json:"continent,omitempty"`
	ContinentCode string `json:"continent_code,omitempty"`
	Postal        string `json:"postal,omitempty"`
	Loc           string `json:"loc,omitempty"`
	Timezone      string `json:"timezone,omitempty"`
}

func geoliteDBPath() string {
	if p := strings.TrimSpace(os.Getenv("GEOLITE_CITY_PATH")); p != "" {
		return p
	}
	return filepath.Join("data", "GeoLite2-City.mmdb")
}

func maxmindCreds() (accountID, licenseKey string) {
	accountID = strings.TrimSpace(os.Getenv("MAXMIND_ACCOUNT_ID"))
	licenseKey = strings.TrimSpace(os.Getenv("MAXMIND_LICENSE_KEY"))
	return accountID, licenseKey
}

func maxAgeForRefresh() time.Duration {
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

// initGeoLite downloads (when configured) and opens the GeoLite2-City database.
// It is non-fatal: the HTTP server still runs if GeoIP is unavailable.
func initGeoLite(ctx context.Context) {
	path := geoliteDBPath()
	acc, key := maxmindCreds()
	if acc != "" && key != "" {
		need, reason := shouldDownloadGeolite(path, maxAgeForRefresh())
		if need {
			log.Printf("geolite: updating database (%s)", reason)
			if err := downloadGeoLiteCity(ctx, path, acc, key); err != nil {
				log.Printf("geolite: download failed: %v (using existing file if any)", err)
			}
		}
	} else {
		log.Println("geolite: MAXMIND_ACCOUNT_ID / MAXMIND_LICENSE_KEY not set; only using local MMDB if present")
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			log.Println("geolite: no database file; /all and /json will omit geo")
			return
		}
		log.Printf("geolite: stat %s: %v", path, err)
		return
	}

	if err := openGeoDB(path); err != nil {
		log.Printf("geolite: open: %v", err)
	}
}

func shouldDownloadGeolite(path string, maxAge time.Duration) (bool, string) {
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

func downloadGeoLiteCity(ctx context.Context, destPath, accountID, licenseKey string) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, geoliteCityDownloadURL, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(accountID, licenseKey)

	client := &http.Client{
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

	resp, err := client.Do(req)
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

	tmp, err := os.CreateTemp(filepath.Dir(destPath), "GeoLite2-City-*.mmdb")
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
		if !strings.HasSuffix(h.Name, "GeoLite2-City.mmdb") {
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
		return fmt.Errorf("tar.gz did not contain GeoLite2-City.mmdb")
	}

	if err := os.Rename(tmpPath, destPath); err != nil {
		return err
	}
	ok = true
	log.Printf("geolite: wrote %s", destPath)
	return nil
}

func openGeoDB(path string) error {
	geoMu.Lock()
	defer geoMu.Unlock()
	if geoDB != nil {
		_ = geoDB.Close()
		geoDB = nil
	}
	db, err := geoip2.Open(path)
	if err != nil {
		return err
	}
	geoDB = db
	log.Printf("geolite: loaded %s (%s)", path, db.Metadata().DatabaseType)
	return nil
}

func ipToAddr(ip net.IP) (netip.Addr, bool) {
	if ip == nil {
		return netip.Addr{}, false
	}
	if v4 := ip.To4(); v4 != nil {
		a, ok := netip.AddrFromSlice(v4)
		return a, ok
	}
	return netip.AddrFromSlice(ip.To16())
}

func lookupVisitorGeo(r *http.Request) *geoRecord {
	ip := preferredIPv4(r)
	if ip == nil {
		ip = preferredIPv6(r)
	}
	if ip == nil {
		return nil
	}
	return lookupGeo(ip)
}

func lookupGeo(ip net.IP) *geoRecord {
	a, ok := ipToAddr(ip)
	if !ok {
		return nil
	}

	geoMu.RLock()
	db := geoDB
	geoMu.RUnlock()
	if db == nil {
		return nil
	}

	rec, err := db.City(a)
	if err != nil || rec == nil || !rec.HasData() {
		return nil
	}

	g := &geoRecord{}
	if n := rec.City.Names.English; n != "" {
		g.City = n
	}
	if len(rec.Subdivisions) > 0 {
		g.Region = rec.Subdivisions[0].Names.English
		g.RegionISO = rec.Subdivisions[0].ISOCode
	}
	if c := rec.Country.ISOCode; c != "" {
		g.Country = c
	}
	if n := rec.Country.Names.English; n != "" {
		g.CountryName = n
	}
	if n := rec.Continent.Names.English; n != "" {
		g.Continent = n
	}
	if c := rec.Continent.Code; c != "" {
		g.ContinentCode = c
	}
	if c := rec.Postal.Code; c != "" {
		g.Postal = c
	}
	if rec.Location.HasCoordinates() {
		g.Loc = fmt.Sprintf("%f,%f", *rec.Location.Latitude, *rec.Location.Longitude)
	}
	if tz := rec.Location.TimeZone; tz != "" {
		g.Timezone = tz
	}

	if g.City == "" && g.Country == "" && g.CountryName == "" && g.Loc == "" && g.Timezone == "" {
		return nil
	}
	return g
}

// writeGeoSection appends a GeoLite2 block to /all when lookupVisitorGeo returns data.
func writeGeoSection(b *strings.Builder, r *http.Request) {
	g := lookupVisitorGeo(r)
	if g == nil {
		return
	}
	fmt.Fprintf(b, "GeoLite2 (city-level, approximate)\n")
	if g.City != "" {
		fmt.Fprintf(b, "  City: %s\n", g.City)
	}
	if g.Region != "" {
		if g.RegionISO != "" {
			fmt.Fprintf(b, "  Region: %s (%s)\n", g.Region, g.RegionISO)
		} else {
			fmt.Fprintf(b, "  Region: %s\n", g.Region)
		}
	}
	if g.CountryName != "" && g.Country != "" {
		fmt.Fprintf(b, "  Country: %s (%s)\n", g.CountryName, g.Country)
	} else if g.Country != "" {
		fmt.Fprintf(b, "  Country: %s\n", g.Country)
	} else if g.CountryName != "" {
		fmt.Fprintf(b, "  Country: %s\n", g.CountryName)
	}
	if g.Continent != "" {
		if g.ContinentCode != "" {
			fmt.Fprintf(b, "  Continent: %s (%s)\n", g.Continent, g.ContinentCode)
		} else {
			fmt.Fprintf(b, "  Continent: %s\n", g.Continent)
		}
	}
	if g.Postal != "" {
		fmt.Fprintf(b, "  Postal: %s\n", g.Postal)
	}
	if g.Loc != "" {
		fmt.Fprintf(b, "  Loc: %s\n", g.Loc)
	}
	if g.Timezone != "" {
		fmt.Fprintf(b, "  Timezone: %s\n", g.Timezone)
	}
}
