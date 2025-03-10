package geo

import (
	"math"
)

// Haversine function to calculate the distance between two points
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Radius of Earth in kilometers

	// Convert latitude and longitude from degrees to radians
	lat1Rad := degreesToRadians(lat1)
	lon1Rad := degreesToRadians(lon1)
	lat2Rad := degreesToRadians(lat2)
	lon2Rad := degreesToRadians(lon2)

	// Haversine formula
	deltaLat := lat2Rad - lat1Rad
	deltaLon := lon2Rad - lon1Rad

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Distance in kilometers
	return R * c
}

// Helper function to convert degrees to radians
func degreesToRadians(deg float64) float64 {
	return deg * (math.Pi / 180)
}

func CalculateGeoDistance(l1 GeoLocation, l2 GeoLocation) float64 {
	distance := haversine(float64(l1.Latitude), float64(l1.Longitude), float64(l2.Latitude), float64(l2.Longitude))
	return distance
}
