package placekey

import (
	_ "embed"
	"testing"
)

func TestFixHeaderInt(t *testing.T) {
	var expected uint64
	switch resolution {
	case 10:
		expected = 621496748577128448
	default:
		t.Fatal("unexpected resolution")
	}
	if fixHeaderInt != expected {
		t.Errorf(`fixHeaderInt = "%d"; wanted %d`, fixHeaderInt, expected)
	}
}

func TestToH3String(t *testing.T) {
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
			got, err := ToH3String(tt.placeKey)
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
			got, err := FromH3String(tt.h3Index)
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

func TestFormatIsValid(t *testing.T) {
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
			want:     true, // H3.IsValid(ToH3Index(placeKey)) = false
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatIsValid(tt.placeKey)
			if got != tt.want {
				t.Errorf("FormatIsValid() got = %v; expected %v", got, tt.want)
			}
		})
	}
}
