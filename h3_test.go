package placekey

import (
	_ "embed"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

const tolerance float64 = 0.001

//go:embed test/example_geos.csv
var exampleGeosCSV []byte

//go:embed test/example_distances.tsv
var exampleDistanceTSV []byte

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) <= tolerance
}

func TestH3IsValid(t *testing.T) {
	c := NewH3()
	defer c.Close()
	tests := []struct {
		name     string
		placeKey string
		want     bool
		wantErr  bool
	}{
		{
			name:     "invalid where value",
			placeKey: "@abc-234-xyz",
			want:     false,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, err := ToH3Index(tt.placeKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToH3Index() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got := c.h3.IsValid(x)
			if got != tt.want {
				t.Errorf("FormatIsValid() got = %v; expected %v", got, tt.want)
			}
		})
	}
}

func TestH3ToGeo(t *testing.T) {
	c := NewH3()
	defer c.Close()
	tests := []struct {
		name     string
		placeKey string
		wantLat  float64
		wantLng  float64
		wantErr  bool
	}{
		{
			name:     "0,0",
			placeKey: "@dvt-smp-tvz",
			wantLat:  0,
			wantLng:  0,
			wantErr:  false,
		},
		{
			name:     "SF City Hall",
			placeKey: "@5vg-7gq-tvz",
			wantLat:  37.779274,
			wantLng:  -122.419262,
			wantErr:  false,
		},
		{
			name:     "EXO",
			placeKey: "@nxd-g5g-xyv",
			wantLat:  -34.63582919120901,
			wantLng:  -58.41313384603939,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotLat, gotLng, err := c.ToGeo(tt.placeKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToGeo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !almostEqual(gotLat, tt.wantLat) {
				t.Errorf("ToGeo() gotLat = %v, want %v", gotLat, tt.wantLat)
			}
			if !almostEqual(gotLng, tt.wantLng) {
				t.Errorf("ToGeo() gotLng = %v, want %v", gotLng, tt.wantLng)
			}
		})
	}
}

func TestH3ToGeoBoundary(t *testing.T) {
	c := NewH3()
	defer c.Close()
	tests := []struct {
		name    string
		h3Index string
		want    []struct{ Lat, Lng float64 }
		wantErr bool
	}{
		{
			name:    "0,0",
			h3Index: "8a2a1072b59ffff",
			want: []struct{ Lat, Lng float64 }{
				{40.690058601, -74.044151762},
				{40.689907695, -74.045061792},
				{40.689270936, -74.045341418},
				{40.688785091, -74.044711031},
				{40.688935993, -74.043801021},
				{40.689572744, -74.043521377},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.ToGeoBoundary(tt.h3Index)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToGeoBoundary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToGeoBoundary() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestH3Distance(t *testing.T) {
	c := NewH3()
	defer c.Close()
	lines := strings.Split(string(exampleDistanceTSV), "\n")
	for index, line := range lines {
		if index == 0 {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) != 8 {
			continue
		}
		// place_key1 geo_1 place_key2 geo_2 distance(km) error
		placeKey1 := parts[0]
		placeKey2 := parts[2]
		expectDistStr := parts[4]
		expectDistErrStr := parts[5]
		expectDist, err := strconv.ParseFloat(expectDistStr, 64)
		if err != nil {
			t.Fatal(err)
		}
		expectDistErr, err := strconv.ParseFloat(expectDistErrStr, 64)
		if err != nil {
			t.Fatal(err)
		}
		distance, err := c.Distance(placeKey1, placeKey2)
		if err != nil {
			t.Fatal(err)
		}
		got := math.Abs(distance/1000 - expectDist)
		if got > expectDistErr {
			t.Errorf("Distance() got = %f; exceeds %f expected error", got, expectDistErr)
		}
	}
}

func TestH3FromGeoToGeo(t *testing.T) {
	c := NewH3()
	defer c.Close()
	lines := strings.Split(string(exampleGeosCSV), "\n")
	for index, line := range lines {
		if index == 0 {
			continue
		}
		parts := strings.Split(line, ",")
		if len(parts) != 8 {
			continue
		}
		// lat,lng,h3_r10,h3_int_r10,placekey,h3_lat,h3_lng,info
		latStr := parts[0]
		longStr := parts[1]
		expected := parts[4]
		lat, err := strconv.ParseFloat(latStr, 64)
		if err != nil {
			t.Fatal(err)
		}
		lng, err := strconv.ParseFloat(longStr, 64)
		if err != nil {
			t.Fatal(err)
		}
		got, err := c.FromGeo(lat, lng)
		if err != nil {
			t.Fatal(err)
		}
		if got != expected {
			t.Errorf(`FromGeo() line %d got = "%s"; expected %s`, index, got, expected)
		}
		gotLat, gotLng, err := c.ToGeo(got)
		if err != nil {
			t.Fatal(err)
		}
		if math.Abs(lat-gotLat) > 0.1 {
			t.Errorf("ToGeo() gotLat = %v, expected %v", gotLat, lat)
		}
		if math.Abs(lng-gotLng) > 0.1 {
			t.Errorf("ToGeo() gotLng = %v, expected %v", gotLng, lng)
		}
	}
}

func BenchmarkH3GeoToPlacekey(b *testing.B) {
	c := NewH3()
	defer c.Close()
	for i := 0; i < b.N; i++ {
		_, _ = c.FromGeo(37.779274, -122.419262)
	}
}

func BenchmarkH3PlacekeyToGeo(b *testing.B) {
	c := NewH3()
	defer c.Close()
	for i := 0; i < b.N; i++ {
		_, _, _ = c.ToGeo("@5vg-7gq-tvz")
	}
}

func TestH3ToGeoIssues(t *testing.T) {
	c := NewH3()
	defer c.Close()
	tests := []struct {
		name    string
		h3Index string
		wantLat float64
		wantLng float64
		wantErr bool
	}{
		{
			// https://github.com/uber/h3-go/issues/7
			name:    "ToGeo function return values inconsistent #7",
			h3Index: "8c194ad30d067ff",
			wantLat: 51.523416454245556,
			wantLng: -0.08106823052469281,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, err := FromH3String(tt.h3Index)
			if err != nil {
				t.Fatal(err)
			}
			gotLat, gotLng, err := c.ToGeo(x)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToGeo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !almostEqual(gotLat, tt.wantLat) {
				t.Errorf("ToGeo() gotLat = %v, want %v", gotLat, tt.wantLat)
			}
			if !almostEqual(gotLng, tt.wantLng) {
				t.Errorf("ToGeo() gotLng = %v, want %v", gotLng, tt.wantLng)
			}
		})
	}
}
