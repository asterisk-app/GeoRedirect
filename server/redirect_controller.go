package server

import (
	"errors"
	"log"
	"net"
	"net/http"
	"strings"

	"io.github.asterisk-app.geo-redirect/geo"
)

type RedirectController struct {
	geoService *geo.GeoLocationService
}

func NewRedirectController(geoService *geo.GeoLocationService) *RedirectController {
	return &RedirectController{
		geoService: geoService,
	}
}

func (rc *RedirectController) Redirect(w http.ResponseWriter, r *http.Request) {
	// Extract the IP address from the request.
	clientIP, err := rc.getClientIP(r)
	if err != nil {
		log.Printf("error getting IP address: [%s]\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Get client location data
	clientLocation, err := rc.geoService.GetLocation(clientIP)
	if err != nil {
		log.Printf("error getting location for IP \"%s\": [%v]\n", clientIP, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	} else {
		log.Printf("Client IP: %s, Location: %v\n", clientIP, clientLocation)
	}
	// Find nearest server
	var region = rc.geoService.GetNearestServer(clientLocation)
	log.Printf("Nearest server is located in %s and host is %s", region.City, region.Host)
	// Log the redirection.
	log.Printf("Redirecting IP %s to %s", clientIP, region.Host)
	// Set Cache-Control header to cache the redirection for 1 hour.
	w.Header().Set("Cache-Control", "max-age=3600")
	http.Redirect(w, r, region.Host, http.StatusFound) // 302 Found
}

// getClientIP extracts the client IP address from the request
func (rc *RedirectController) getClientIP(r *http.Request) (string, error) {
	// Check for X-Forwarded-For header
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// The first IP in the list is the original client IP
		ips := strings.Split(forwarded, ",")
		return strings.TrimSpace(ips[0]), nil
	}
	// Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return "", errors.New("unable to determine IP address")
	}
	return ip, nil
}
