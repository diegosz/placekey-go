//nolint:gomnd
package placekey

import (
	"errors"
	"math"

	"github.com/diegosz/placekey-go/internal/h3"
)

const (
	earthRadius float64 = 6371.0 // km
)

var ErrInvalidLatLngRange = errors.New("invalid lat/lng range")

type H3 struct {
	h3 *h3.H3
}

func NewH3() *H3 {
	return &H3{h3: h3.NewH3()}
}

func (c *H3) Close() {
	c.h3.Close()
}

// IsValid returns whether or not the H3 index is a valid cell (hexagon or
// pentagon).
func (c *H3) IsValid(placeKey string) bool {
	x, err := ToH3Index(placeKey)
	if err != nil {
		return false
	}
	return c.h3.IsValid(x)
}

// FromGeo converts a (latitude, longitude) into a PlaceKey.
func (c *H3) FromGeo(lat, lng float64) (string, error) {
	if lat < -90 || lat > 90 || lng < -180 || lng > 180 {
		return "", ErrInvalidLatLngRange
	}
	return encodeH3Int(uint64(c.h3.FromGeo(h3.GeoCoord{Latitude: lat, Longitude: lng}, resolution))), nil
}

// ToGeo converts a PlaceKey into a (latitude, longitude).
func (c *H3) ToGeo(placeKey string) (lat, lng float64, err error) {
	x, err := ToH3Int(placeKey)
	if err != nil {
		return 0.0, 0.0, err
	}
	geo := c.h3.ToGeo(h3.Index(x))
	return geo.Latitude, geo.Longitude, nil
}

// ToGeoBoundary returns the hexagonal polygon boundary of a PlaceKey as a slice
// of (latitude, longitude) coordinates.
func (c *H3) ToGeoBoundary(placeKey string) ([][]float64, error) {
	x, err := ToH3Index(placeKey)
	if err != nil {
		return nil, err
	}
	h := [][]float64{}
	for _, c := range c.h3.ToGeoBoundary(x) {
		h = append(h, []float64{c.Latitude, c.Longitude})
	}
	return h, nil
}

// Distance returns the distance in meters between the centers of two PlaceKeys.
func (c *H3) Distance(placeKey1, placeKey2 string) (float64, error) {
	lat1, lng1, err := c.ToGeo(placeKey1)
	if err != nil {
		return 0, err
	}
	lat2, lng2, err := c.ToGeo(placeKey2)
	if err != nil {
		return 0, err
	}
	return geoDistance(lat1, lng1, lat2, lng2), nil
}

// geoDistance returns the distance in meters between two (latitude, longitude)
// coordinates.
func geoDistance(lat1, lng1, lat2, lng2 float64) float64 {
	rLat1 := radians(lat1)
	rLng1 := radians(lng1)
	rLat2 := radians(lat2)
	rLng2 := radians(lng2)
	havLat := 0.5 * (1 - math.Cos(rLat1-rLat2))
	havLng := 0.5 * (1 - math.Cos(rLng1-rLng2))
	radical := math.Sqrt(havLat + math.Cos(rLat1)*math.Cos(rLat2)*havLng)
	return 2.0 * earthRadius * math.Asin(radical) * 1000
}

// radians converts degrees to radians
func radians(degrees float64) float64 {
	return degrees * math.Pi / 180
}

// degrees converts radians to degrees
func degrees(radians float64) float64 { //nolint:deadcode,unused
	return radians / math.Pi * 180
}
