// Command resolve prints GeoLite2 city + ASN JSON for any IP using local MMDB files
// (same GEOLITE_*_PATH env vars as the server). Not exposed over HTTP.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/oschwald/geoip2-golang/v2"
	"github.com/ultimaterex/get-ip/internal/geolookup"
)

func main() {
	log.SetFlags(0)
	args := os.Args[1:]
	if len(args) == 0 {
		usage()
		os.Exit(2)
	}

	if args[0] == "fetch" {
		if len(args) != 1 {
			usage()
			os.Exit(2)
		}
		if err := geolookup.Fetch(context.Background(), log.Printf); err != nil {
			log.Fatal(err)
		}
		return
	}

	if args[0] == "--fetch" || args[0] == "-fetch" {
		args = args[1:]
		if err := geolookup.Fetch(context.Background(), log.Printf); err != nil {
			log.Fatal(err)
		}
		if len(args) == 0 {
			return
		}
		if len(args) != 1 {
			usage()
			os.Exit(2)
		}
		runLookup(args[0])
		return
	}

	if len(args) != 1 {
		usage()
		os.Exit(2)
	}
	runLookup(args[0])
}

func usage() {
	me := os.Args[0]
	fmt.Fprintf(os.Stderr, "usage: %s <ip>\n", me)
	fmt.Fprintf(os.Stderr, "       %s fetch\n", me)
	fmt.Fprintf(os.Stderr, "       %s --fetch [<ip>]\n", me)
	fmt.Fprintf(os.Stderr, "Reads GeoLite MMDBs from GEOLITE_CITY_PATH and GEOLITE_ASN_PATH (same as get-ip server).\n")
	fmt.Fprintf(os.Stderr, "`fetch` / `--fetch` download or refresh MMDBs using MAXMIND_ACCOUNT_ID and MAXMIND_LICENSE_KEY.\n")
}

func runLookup(ipStr string) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		log.Fatalf("invalid IP: %s", ipStr)
	}

	cityPath := geolookup.CityMMDBPath()
	asnPath := geolookup.ASNMMDBPath()

	var cityDB *geoip2.Reader
	var asnDB *geoip2.Reader

	if _, err := os.Stat(cityPath); err == nil {
		db, err := geoip2.Open(cityPath)
		if err != nil {
			log.Fatalf("open city db %s: %v", cityPath, err)
		}
		defer db.Close()
		cityDB = db
	} else if !os.IsNotExist(err) {
		log.Fatalf("stat city db: %v", err)
	}

	if _, err := os.Stat(asnPath); err == nil {
		db, err := geoip2.Open(asnPath)
		if err != nil {
			log.Fatalf("open asn db %s: %v", asnPath, err)
		}
		defer db.Close()
		asnDB = db
	} else if !os.IsNotExist(err) {
		log.Fatalf("stat asn db: %v", err)
	}

	if cityDB == nil && asnDB == nil {
		log.Fatalf("no MMDB files found (expected %s and/or %s)", cityPath, asnPath)
	}

	out := struct {
		IP  string         `json:"ip"`
		Geo *geolookup.Geo `json:"geo,omitempty"`
		ASN *geolookup.ASN `json:"asn,omitempty"`
	}{
		IP:  ip.String(),
		Geo: geolookup.LookupCity(cityDB, ip),
		ASN: geolookup.LookupASN(asnDB, ip),
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(out); err != nil {
		log.Fatal(err)
	}
}
