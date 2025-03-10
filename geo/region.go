package geo

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Region struct {
	Host     string
	City     string
	Location GeoLocation
}

// parseRegion converts an environment variable value into the region name and Region struct.
// The expected format is: region,host,city,lat,lon
func parseRegion(envValue string) (string, Region, error) {
	parts := strings.Split(envValue, ",")
	if len(parts) != 5 {
		return "", Region{}, fmt.Errorf("invalid format, expected 5 parts but got %d", len(parts))
	}
	var r Region
	r.Host = parts[1] // host value
	r.City = parts[2] // city value

	// Parse the latitude.
	if _, err := fmt.Sscanf(parts[3], "%f", &r.Location.Latitude); err != nil {
		return "", Region{}, fmt.Errorf("error parsing latitude: %v", err)
	}

	// Parse the longitude.
	if _, err := fmt.Sscanf(parts[4], "%f", &r.Location.Longitude); err != nil {
		return "", Region{}, fmt.Errorf("error parsing longitude: %v", err)
	}

	regionName := parts[0] // e.g. "asia", "europe", "america"
	return regionName, r, nil
}

// GetRegions reads environment variables using the SERVER_<i> pattern and returns a map of regions.
func GetRegions() (map[string]Region, error) {
	regions := make(map[string]Region)
	for i := 1; ; i++ {
		// Build the environment variable key dynamically.
		key := "SERVER_" + strconv.Itoa(i)
		value := os.Getenv(key)
		if value == "" {
			// Break the loop when an environment variable is not found.
			break
		}
		regionName, region, err := parseRegion(value)
		if err != nil {
			return nil, fmt.Errorf("error parsing %s: %v", key, err)
		}
		regions[regionName] = region
	}
	return regions, nil
}
