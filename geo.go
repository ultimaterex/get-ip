package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/oschwald/geoip2-golang/v2"

	"github.com/ultimaterex/get-ip/internal/geolookup"
)

var (
	geoMu     sync.RWMutex
	geoCityDB *geoip2.Reader
	geoASNDB  *geoip2.Reader
)

// initGeoLite downloads (when configured) and opens GeoLite2-City and GeoLite2-ASN databases.
// It is non-fatal: the HTTP server still runs if GeoIP is unavailable.
func initGeoLite(ctx context.Context) {
	geolookup.SyncIfConfigured(ctx, log.Printf)

	cityPath := geolookup.CityMMDBPath()
	asnPath := geolookup.ASNMMDBPath()

	if _, err := os.Stat(cityPath); err == nil {
		if err := openCityDB(cityPath); err != nil {
			log.Printf("geolite: open city db: %v", err)
		}
	} else if !os.IsNotExist(err) {
		log.Printf("geolite: stat city db: %v", err)
	}

	if _, err := os.Stat(asnPath); err == nil {
		if err := openASNDB(asnPath); err != nil {
			log.Printf("geolite: open asn db: %v", err)
		}
	} else if !os.IsNotExist(err) {
		log.Printf("geolite: stat asn db: %v", err)
	}

	if geoCityDB == nil && geoASNDB == nil {
		if _, e1 := os.Stat(cityPath); os.IsNotExist(e1) {
			if _, e2 := os.Stat(asnPath); os.IsNotExist(e2) {
				log.Println("geolite: no MMDB files found; geo/asn omitted from responses")
			}
		}
	}
}

func openCityDB(path string) error {
	geoMu.Lock()
	defer geoMu.Unlock()
	if geoCityDB != nil {
		_ = geoCityDB.Close()
		geoCityDB = nil
	}
	db, err := geoip2.Open(path)
	if err != nil {
		return err
	}
	geoCityDB = db
	log.Printf("geolite: loaded %s (%s)", path, db.Metadata().DatabaseType)
	return nil
}

func openASNDB(path string) error {
	geoMu.Lock()
	defer geoMu.Unlock()
	if geoASNDB != nil {
		_ = geoASNDB.Close()
		geoASNDB = nil
	}
	db, err := geoip2.Open(path)
	if err != nil {
		return err
	}
	geoASNDB = db
	log.Printf("geolite: loaded %s (%s)", path, db.Metadata().DatabaseType)
	return nil
}

func lookupVisitorGeo(r *http.Request) *geolookup.Geo {
	ip := preferredIPv4(r)
	if ip == nil {
		ip = preferredIPv6(r)
	}
	if ip == nil {
		return nil
	}
	return lookupGeo(ip)
}

func lookupGeo(ip net.IP) *geolookup.Geo {
	geoMu.RLock()
	db := geoCityDB
	geoMu.RUnlock()
	return geolookup.LookupCity(db, ip)
}

// writeGeoSection appends an estimated-location block to /all when lookupVisitorGeo returns data.
func writeGeoSection(b *strings.Builder, r *http.Request) {
	g := lookupVisitorGeo(r)
	if g == nil {
		return
	}
	fmt.Fprintf(b, "Estimated location\n")
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

func lookupVisitorASN(r *http.Request) *geolookup.ASN {
	ip := preferredIPv4(r)
	if ip == nil {
		ip = preferredIPv6(r)
	}
	if ip == nil {
		return nil
	}
	return lookupASN(ip)
}

func lookupASN(ip net.IP) *geolookup.ASN {
	geoMu.RLock()
	db := geoASNDB
	geoMu.RUnlock()
	return geolookup.LookupASN(db, ip)
}

// writeASNSection appends an ASN / network block to /all when lookupVisitorASN returns data.
func writeASNSection(b *strings.Builder, r *http.Request) {
	a := lookupVisitorASN(r)
	if a == nil {
		return
	}
	fmt.Fprintf(b, "Network\n")
	if a.ASN != 0 {
		fmt.Fprintf(b, "  ASN: %d\n", a.ASN)
	}
	if a.Organization != "" {
		fmt.Fprintf(b, "  Organization: %s\n", a.Organization)
	}
	if a.Network != "" {
		fmt.Fprintf(b, "  Network: %s\n", a.Network)
	}
}
