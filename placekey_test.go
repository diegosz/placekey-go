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

func TestHeaderInt(t *testing.T) {
	pk := NewPlaceKey()
	defer pk.Close()
	var expected uint64
	switch resolution {
	case 10:
		expected = 621496748577128448
	}
	if pk.headerInt != expected {
		t.Errorf(`headerInt = "%d"; wanted %d`, pk.headerInt, expected)
	}
}

func TestToGeo(t *testing.T) {
	pk := NewPlaceKey()
	defer pk.Close()
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
			gotLat, gotLng, err := pk.ToGeo(tt.placeKey)
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

func TestToH3String(t *testing.T) {
	pk := NewPlaceKey()
	defer pk.Close()
	tests := []struct {
		name     string
		placeKey string
		want     string
		wantErr  bool
	}{
		{
			name:     "0,0",
			placeKey: "@dvt-smp-tvz",
			want:     "8a754e64992ffff",
			wantErr:  false,
		},
		{
			name:     "SF City Hall",
			placeKey: "@5vg-7gq-tvz",
			want:     "8a2830828767fff",
			wantErr:  false,
		},
		{
			name:     "Ferry Building in San Francisco",
			placeKey: "zzw-22y@5vg-7gt-qzz",
			want:     "8a283082a677fff",
			wantErr:  false,
		},
		{
			name:     "EXO",
			placeKey: "@nxd-g5g-xyv",
			want:     "8ac2e31064effff",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pk.ToH3String(tt.placeKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToH3String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ToH3String() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromH3String(t *testing.T) {
	pk := NewPlaceKey()
	defer pk.Close()
	tests := []struct {
		name    string
		h3Index string
		want    string
		wantErr bool
	}{
		{
			name:    "0,0",
			h3Index: "8a754e64992ffff",
			want:    "@dvt-smp-tvz",
			wantErr: false,
		},
		{
			name:    "SF City Hall",
			h3Index: "8a2830828767fff",
			want:    "@5vg-7gq-tvz",
			wantErr: false,
		},
		{
			name:    "Ferry Building in San Francisco",
			h3Index: "8a283082a677fff",
			want:    "@5vg-7gt-qzz",
			wantErr: false,
		},
		{
			name:    "EXO",
			h3Index: "8ac2e31064effff",
			want:    "@nxd-g5g-xyv",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pk.FromH3String(tt.h3Index)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromH3String() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FromH3String() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToHexagonalBoundary(t *testing.T) {
	pk := NewPlaceKey()
	defer pk.Close()
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
			got, err := pk.ToHexagonalBoundary(tt.h3Index)
			if (err != nil) != tt.wantErr {
				t.Errorf("ToHexagonalBoundary() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if reflect.DeepEqual(got, tt.want) {
				t.Errorf("ToHexagonalBoundary() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDistance(t *testing.T) {
	pk := NewPlaceKey()
	defer pk.Close()
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
		distance, err := pk.Distance(placeKey1, placeKey2)
		if err != nil {
			t.Fatal(err)
		}
		got := math.Abs(distance/1000 - expectDist)
		if got > expectDistErr {
			t.Errorf("Distance() got = %f; exceeds %f expected error", got, expectDistErr)
		}
	}
}

func TestFromGeoToGeo(t *testing.T) {
	pk := NewPlaceKey()
	defer pk.Close()
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
		got, err := pk.FromGeo(lat, lng)
		if err != nil {
			t.Fatal(err)
		}
		if got != expected {
			t.Errorf(`FromGeo() line %d got = "%s"; expected %s`, index, got, expected)
		}
		gotLat, gotLng, err := pk.ToGeo(got)
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

func TestFormatIsValid(t *testing.T) {
	pk := NewPlaceKey()
	defer pk.Close()
	tests := []struct {
		name     string
		placeKey string
		want     bool
	}{
		{
			name:     "222-227@dvt-smp-tvz",
			placeKey: "222-227@dvt-smp-tvz",
			want:     true,
		},
		{
			name:     "where with no @",
			placeKey: "5vg-7gq-tvz",
			want:     true,
		},
		{
			name:     "where with @",
			placeKey: "@5vg-7gq-tvz",
			want:     true,
		},
		{
			name:     "single tuple what with where",
			placeKey: "zzz@5vg-7gq-tvz",
			want:     true,
		},
		{
			name:     "double tuple what with where",
			placeKey: "222-zzz@5vg-7gq-tvz",
			want:     true,
		},
		{
			name:     "long address encoding with where",
			placeKey: "2222-zzz@5vg-7gq-tvz",
			want:     false,
		},
		{
			name:     "long poi encoding with where",
			placeKey: "222-zzzz@5vg-7gq-tvz",
			want:     false,
		},
		{
			name:     "long address and poi encoding with where",
			placeKey: "22222222-zzzzzzzzz@5vg-7gq-tvz",
			want:     false,
		},
		{
			name:     "@123-456-789",
			placeKey: "@123-456-789",
			want:     false,
		},
		{
			name:     "short where part",
			placeKey: "@abc",
			want:     false,
		},
		{
			name:     "short where part",
			placeKey: "abc-xyz",
			want:     false,
		},
		{
			name:     "no dashes",
			placeKey: "abcxyz234",
			want:     false,
		},
		{
			name:     "padding character in what",
			placeKey: "abc-345@abc-234-xyz",
			want:     false,
		},
		{
			name:     "replacement character in what",
			placeKey: "ebc-345@abc-234-xyz",
			want:     false,
		},
		{
			name:     "missing what part",
			placeKey: "bcd-345@",
			want:     false,
		},
		{
			name:     "short address encoding",
			placeKey: "22-zzz@abc-234-xyz",
			want:     false,
		},
		{
			name:     "short poi encoding",
			placeKey: "222-zz@abc-234-xyz",
			want:     false,
		},
		{
			name:     "invalid where value",
			placeKey: "@abc-234-xyz",
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pk.FormatIsValid(tt.placeKey)
			if got != tt.want {
				t.Errorf("FormatIsValid() got = %v; expected %v", got, tt.want)
			}
		})
	}
}

func BenchmarkGeoToPlacekey(b *testing.B) {
	pk := NewPlaceKey()
	defer pk.Close()
	for i := 0; i < b.N; i++ {
		_, _ = pk.FromGeo(37.779274, -122.419262)
	}
}

func BenchmarkPlacekeyToGeo(b *testing.B) {
	pk := NewPlaceKey()
	defer pk.Close()
	for i := 0; i < b.N; i++ {
		_, _, _ = pk.ToGeo("@5vg-7gq-tvz")
	}
}
