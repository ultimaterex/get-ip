package geolookup

import (
	"fmt"
	"net"
	"net/netip"
	"os"
	"path/filepath"
	"strings"

	"github.com/oschwald/geoip2-golang/v2"
)

// Geo holds City-database fields for JSON and text output.
type Geo struct {
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

// ASN holds ASN-database fields for JSON and text output.
type ASN struct {
	ASN          uint   `json:"asn,omitempty"`
	Organization string `json:"organization,omitempty"`
	Network      string `json:"network,omitempty"`
}

// CityMMDBPath returns GEOLITE_CITY_PATH or the default under ./data.
func CityMMDBPath() string {
	if p := strings.TrimSpace(os.Getenv("GEOLITE_CITY_PATH")); p != "" {
		return p
	}
	return filepath.Join("data", "GeoLite2-City.mmdb")
}

// ASNMMDBPath returns GEOLITE_ASN_PATH or the default under ./data.
func ASNMMDBPath() string {
	if p := strings.TrimSpace(os.Getenv("GEOLITE_ASN_PATH")); p != "" {
		return p
	}
	return filepath.Join("data", "GeoLite2-ASN.mmdb")
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

// LookupCity performs a GeoLite2-City lookup; returns nil if db is nil or no data.
func LookupCity(db *geoip2.Reader, ip net.IP) *Geo {
	a, ok := ipToAddr(ip)
	if !ok {
		return nil
	}
	if db == nil {
		return nil
	}

	rec, err := db.City(a)
	if err != nil || rec == nil || !rec.HasData() {
		return nil
	}

	g := &Geo{}
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

// LookupASN performs a GeoLite2-ASN lookup; returns nil if db is nil or no data.
func LookupASN(db *geoip2.Reader, ip net.IP) *ASN {
	a, ok := ipToAddr(ip)
	if !ok {
		return nil
	}
	if db == nil {
		return nil
	}

	rec, err := db.ASN(a)
	if err != nil || rec == nil || !rec.HasData() {
		return nil
	}

	out := &ASN{
		ASN:          rec.AutonomousSystemNumber,
		Organization: rec.AutonomousSystemOrganization,
	}
	if rec.Network.IsValid() {
		out.Network = rec.Network.String()
	}
	if out.ASN == 0 && out.Organization == "" && out.Network == "" {
		return nil
	}
	return out
}
