package main

import (
	"errors"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"io.github.asterisk-app.geo-redirect/geo"
	"io.github.asterisk-app.geo-redirect/server"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Printf("some error occurred while reading .env file [error: %s]", err)
	}

	// Get IP address db file names and base url to download them
	var baseURL = os.Getenv("BASE_URL")
	var ipv4DBFile = os.Getenv("IPV4_DB_FILE")
	var ipv6DBFile = os.Getenv("IPV6_DB_FILE")

	if ipv4DBFile == "" || ipv6DBFile == "" || baseURL == "" {
		log.Fatal("failed to find either db file names or base url to download them")
	}

	regions, err := geo.GetRegions()
	if err != nil {
		log.Printf("error retrieving regions: %v\n", err)
		return
	}

	// Initialize geo service
	geoService := geo.NewGeoLocationService(regions, ipv4DBFile, ipv6DBFile, baseURL)
	// Download geo databases if needed
	if !fileExists(geoService.IPv4DBFile) || !fileExists(geoService.IPv6DBFile) {
		success := geoService.DownloadDBs()
		if !success {
			log.Fatal("failed to download geo location databases")
		}
	}

	// Initialize redirect controller
	redirectController := server.NewRedirectController(geoService)
	// Set up HTTP routes
	http.HandleFunc("/", redirectController.Redirect)
	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("starting geo redirect server on port %s\n", port)
	err = http.ListenAndServe(":"+port, nil)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("error starting server:", err)
	}
}

// Helper function to check if file exists
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
