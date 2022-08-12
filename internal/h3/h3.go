package h3

import (
	"math"
	"strconv"
	"unsafe"

	"github.com/akhenakh/goh3/ch3"
	"modernc.org/libc"
)

var (
	deg2rad = math.Pi / 180.0
	rad2deg = 180.0 / math.Pi
)

type H3Index ch3.TH3Index

type GeoCoord struct {
	Latitude, Longitude float64
}

type GeoPolygon struct {
	Geofence []GeoCoord

	Holes [][]GeoCoord
}

type H3 struct {
	*libc.TLS
}

func NewH3() *H3 {
	return &H3{TLS: libc.NewTLS()}
}

func (c *H3) Close() {
	c.TLS.Close()
}

func (c *H3) FromGeo(geo GeoCoord, res int) H3Index {
	cgeo := ch3.TGeoCoord{
		Flat: deg2rad * geo.Latitude,
		Flon: deg2rad * geo.Longitude,
	}
	return H3Index(ch3.XgeoToH3(c.TLS, uintptr(unsafe.Pointer(&cgeo)), int32(res)))
}

func (c *H3) ToGeo(h H3Index) GeoCoord {
	cg := ch3.TGeoCoord{}
	ch3.Xh3ToGeo(c.TLS, ch3.TH3Index(h), uintptr(unsafe.Pointer(&cg)))
	g := GeoCoord{}
	g.Latitude = rad2deg * cg.Flat
	g.Longitude = rad2deg * cg.Flon
	return g
}

func (c *H3) ToString(h H3Index) string {
	return strconv.FormatUint(uint64(h), 16)
}

func (c *H3) ToGeoBoundary(h H3Index) []GeoCoord {
	gb := ch3.TGeoBoundary{}
	ch3.Xh3ToGeoBoundary(c.TLS, ch3.TH3Index(h), uintptr(unsafe.Pointer(&gb)))
	gs := make([]GeoCoord, 0, gb.FnumVerts)
	for i := 0; i < int(gb.FnumVerts); i++ {
		g := GeoCoord{}
		g.Latitude = rad2deg * gb.Fverts[i].Flat
		g.Longitude = rad2deg * gb.Fverts[i].Flon
		gs = append(gs, g)
	}
	return gs
}

func (c *H3) IsValid(h H3Index) bool {
	return ch3.Xh3IsValid(c.TLS, ch3.TH3Index(h)) == 1
}
