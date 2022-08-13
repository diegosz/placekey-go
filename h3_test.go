package placekey

import (
	_ "embed"
	"fmt"
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

func TestH3_IsValid(t *testing.T) {
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
			c := NewH3()
			defer c.Close()
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

func TestH3_ToGeo(t *testing.T) {
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
			c := NewH3()
			defer c.Close()
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

func TestH3_Distance(t *testing.T) {
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

func TestH3_FromGeo_ToGeo(t *testing.T) {
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

func BenchmarkH3_GeoToPlacekey(b *testing.B) {
	c := NewH3()
	defer c.Close()
	for i := 0; i < b.N; i++ {
		_, _ = c.FromGeo(37.779274, -122.419262)
	}
}

func BenchmarkH3_PlacekeyToGeo(b *testing.B) {
	c := NewH3()
	defer c.Close()
	for i := 0; i < b.N; i++ {
		_, _, _ = c.ToGeo("@5vg-7gq-tvz")
	}
}

func TestH3_ToGeoIssues(t *testing.T) {
	tests := []struct {
		name    string
		h3Int   uint64
		wantLat float64
		wantLng float64
		wantErr bool
	}{
		{
			// https://github.com/uber/h3-go/issues/7
			name:    "ToGeo function return values inconsistent #7",
			h3Int:   630948894377797631, // "8c194ad30d067ff"
			wantLat: 51.523416454245556,
			wantLng: -0.08106823052469281,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewH3()
			defer c.Close()
			x, err := fromH3IntUnvalidatedResolution(tt.h3Int)
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

// func TestH3_ToGeoBoundary(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		h3Index string
// 		want    [][]float64
// 		wantErr bool
// 	}{
// 		{
// 			name:    "8a2a1072b59ffff",
// 			h3Index: "8a2a1072b59ffff", // "@627-wc5-z2k" // 622236750694711295
// 			want: [][]float64{
// 				{40.6900586009536, -74.04415176176158},
// 				{40.689907694525196, -74.04506179239633},
// 				{40.689270936043556, -74.04534141750702},
// 				{40.688785090724046, -74.04471103053613},
// 				{40.68893599264273, -74.04380102076256},
// 				{40.689572744390546, -74.04352137709905},
// 			},
// 			wantErr: false,
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			c := NewH3()
// 			defer c.Close()
// 			pk, err := FromH3String(tt.h3Index)
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			got, err := c.ToGeoBoundary(pk)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("ToGeoBoundary() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("ToGeoBoundary() got = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }

// FIXME: The above TestH3_ToGeoBoundary test is not working as expected.
// It works fine in debug mode, but fails to run in testing.
// ToGeoBoundary returns empty. Something smells...
// The example test ExampleH3_ToGeoBoundary below is a working fine.
// Also running in an executable binary works fine, at least for the moment...

func ExampleH3_ToGeoBoundary() {
	c := NewH3()
	defer c.Close()
	tests := []struct {
		name    string
		h3Index string
		level   int
		want    [][]float64
		wantErr bool
	}{
		{
			name:    "8a2a1072b59ffff",
			h3Index: "8a2a1072b59ffff", // "@627-wc5-z2k" // 622236750694711295
			level:   0,
			want: [][]float64{
				{40.6900586009536, -74.04415176176158},
				{40.689907694525196, -74.04506179239633},
				{40.689270936043556, -74.04534141750702},
				{40.688785090724046, -74.04471103053613},
				{40.68893599264273, -74.04380102076256},
				{40.689572744390546, -74.04352137709905},
			},
			wantErr: false,
		},
		{
			name:    "pentagon resolution 10",
			h3Index: "8ac200000007fff",
			level:   0,
			want: [][]float64{
				{-39.100455452692714, -57.70029017862053},
				{-39.10035523525216, -57.69953126249851},
				{-39.09976414050208, -57.69941956740002},
				{-39.09949904241486, -57.70010944253918},
				{-39.09992629443885, -57.700647510006675},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		fmt.Println(tt.name)
		pk, err := FromH3String(tt.h3Index)
		if err != nil {
			fmt.Println(err)
			continue
		}
		lat, lng, err := c.ToGeo(pk)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Printf("center: {%s, %s}\n", strconv.FormatFloat(lat, 'f', -1, 64), strconv.FormatFloat(lng, 'f', -1, 64))
		if err != nil {
			fmt.Println(err)
			continue
		}
		got, err := c.ToGeoBoundary(pk)
		fmt.Println("boundary:")
		if (err != nil) != tt.wantErr {
			fmt.Println(err)
			continue
		}
		for _, v := range got {
			fmt.Printf("{%s, %s},\n", strconv.FormatFloat(v[0], 'f', -1, 64), strconv.FormatFloat(v[1], 'f', -1, 64))
		}
		if !reflect.DeepEqual(got, tt.want) {
			fmt.Printf("ToGeoBoundary() got = %v, want %v\n", got, tt.want)
		}
	}
	// Output:
	// 8a2a1072b59ffff
	// center: {40.68942184369931, -74.04443139990863}
	// boundary:
	// {40.6900586009536, -74.04415176176158},
	// {40.689907694525196, -74.04506179239633},
	// {40.689270936043556, -74.04534141750702},
	// {40.688785090724046, -74.04471103053613},
	// {40.68893599264273, -74.04380102076256},
	// {40.689572744390546, -74.04352137709905},
	// pentagon resolution 10
	// center: {-39.1000000339759, -57.69999959221297}
	// boundary:
	// {-39.100455452692714, -57.70029017862053},
	// {-39.10035523525216, -57.69953126249851},
	// {-39.09976414050208, -57.69941956740002},
	// {-39.09949904241486, -57.70010944253918},
	// {-39.09992629443885, -57.700647510006675},
}
