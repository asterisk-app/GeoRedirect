package geo

import (
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"net/netip"
	"os"

	"github.com/oschwald/maxminddb-golang/v2"
)

// GeoLocation represents the geographical details of an IP address
type GeoLocation struct {
	City        string
	CountryCode string
	Latitude    float32
	Longitude   float32
	State       string
	Timezone    string
}

// Service handles geolocation operations
type GeoLocationService struct {
	IPv4DBFile string
	IPv6DBFile string
	BaseURL    string
	regions    map[string]Region
}

// NewGeoLocationService creates a new instance of GeoLocation Service
func NewGeoLocationService(regions map[string]Region, ipv4DBFile, ipv6DBFile, baseUrl string) *GeoLocationService {
	return &GeoLocationService{
		IPv4DBFile: ipv4DBFile,
		IPv6DBFile: ipv6DBFile,
		BaseURL:    baseUrl,
		regions:    regions,
	}
}

// DownloadDBs downloads both IPv4 and IPv6 databases
func (g *GeoLocationService) DownloadDBs() bool {
	ipv4URL := fmt.Sprintf("%s%s", g.BaseURL, g.IPv4DBFile)
	ipv6URL := fmt.Sprintf("%s%s", g.BaseURL, g.IPv6DBFile)
	return g.download(ipv4URL, g.IPv4DBFile) && g.download(ipv6URL, g.IPv6DBFile)
}

// download handles downloading a file from a given URL
func (g *GeoLocationService) download(url, outputFile string) bool {
	log.Println("Downloading file:", url)

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Printf("error:%s\n", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Failed to download file. HTTP Status:", resp.Status)
		return false
	}

	file, err := os.Create(outputFile)
	if err != nil {
		log.Println("error creating file:", err)
		return false
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Println("error saving file:", err)
		return false
	}

	log.Println("download complete:", outputFile)
	return true
}

func (g *GeoLocationService) GetNearestServer(clientLocation *GeoLocation) Region {
	log.Printf("searching nearest server from total %d servers\n", len(g.regions))
	var distance float64 = math.MaxFloat64
	var nearestServer string
	for key, region := range g.regions {
		var currentDistance = CalculateGeoDistance(region.Location, *clientLocation)
		log.Printf("Client's (%s) distance from %s is:%.2f\n", clientLocation.City, region.City, currentDistance)
		if currentDistance <= distance {
			distance = currentDistance
			nearestServer = key
		}
	}
	return g.regions[nearestServer]
}

// GetLocation fetches the geolocation for an IP address
func (g *GeoLocationService) GetLocation(ip string) (*GeoLocation, error) {
	var parsedIp = net.ParseIP(ip)
	// Check if the ip address is from localhost (ipv4/ipv6)
	if parsedIp.IsLoopback() {
		log.Printf("%s is localhost ip address", ip)
		return &GeoLocation{}, fmt.Errorf("%s is localhost ip address", ip)
	}
	if parsedIp.To4() != nil {
		log.Printf("%s is IPV4 address\n", ip)
		return g.getLocationIPv4(ip)
	}
	log.Printf("%s is IPV6 address\n", ip)
	return g.getLocationIPv6(ip)
}

// getLocationIPv4 fetches the geolocation for an IPv4 address
func (g *GeoLocationService) getLocationIPv4(ip string) (*GeoLocation, error) {
	return g.getLocationFromDB(ip, g.IPv4DBFile)
}

// getLocationIPv6 fetches the geolocation for an IPv6 address
func (g *GeoLocationService) getLocationIPv6(ip string) (*GeoLocation, error) {
	return g.getLocationFromDB(ip, g.IPv6DBFile)
}

// getLocationFromDB is a helper function to fetch geolocation from the specified database
func (g *GeoLocationService) getLocationFromDB(ip, dbFile string) (*GeoLocation, error) {
	db, err := maxminddb.Open(dbFile)
	if err != nil {
		return &GeoLocation{}, err
	}
	defer db.Close()

	addr := netip.MustParseAddr(ip)
	result := db.Lookup(addr)
	var geoLocationMap map[string]any
	err = result.Decode(&geoLocationMap)
	if err != nil {
		return &GeoLocation{}, err
	}

	// Create GeoLocation directly from map values
	geoLocation := GeoLocation{
		City:        geoLocationMap["city"].(string),
		CountryCode: geoLocationMap["country_code"].(string),
		Latitude:    geoLocationMap["latitude"].(float32),
		Longitude:   geoLocationMap["longitude"].(float32),
		State:       geoLocationMap["state1"].(string),
		Timezone:    geoLocationMap["timezone"].(string),
	}

	log.Printf("GeoLocation coordinates:%v\n", geoLocation)
	return &geoLocation, nil
}
