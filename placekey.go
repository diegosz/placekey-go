//nolint:gomnd
package placekey

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/diegosz/placekey-go/internal/h3"
)

const (
	resolution             int    = 10
	maxResolution          int    = 15
	baseResolution         int    = 12
	baseCellIncrement      uint64 = 1 << (3 * 15)
	unusedResolutionFiller uint64 = 1<<(3*(15-12)) - 1 // 15-baseResolution
	alphabet               string = "23456789bcdfghjkmnpqrstvwxyz"
	alphabetLength         int64  = 28
	codeLength             int    = 9
	tupleLength            int    = 3
	paddingChar            string = "a"
	replacementChars       string = "eu"
)

var ErrInvalidResolution = errors.New("invalid resolution")
var ErrInvalidParts = errors.New("invalid parts")

var (
	fixHeaderInt   uint64
	replacementMap = map[string]string{
		"prn":   "pre",
		"f4nny": "f4nne",
		"tw4t":  "tw4e",
		"ngr":   "ngu", // 'u' avoids introducing 'gey'
		"dck":   "dce",
		"vjn":   "vju", // 'u' avoids introducing 'jew'
		"fck":   "fce",
		"pns":   "pne",
		"sht":   "she",
		"kkk":   "kke",
		"fgt":   "fgu", // 'u' avoids introducing 'gey'
		"dyk":   "dye",
		"bch":   "bce",
	}
	firstTupleRegex = "[" + alphabet + replacementChars + paddingChar + "]{3}"
	tupleRegex      = "[" + alphabet + replacementChars + "]{3}"
	whereRegex      = regexp.MustCompile("^" + strings.Join([]string{firstTupleRegex, tupleRegex, tupleRegex}, "-") + "$")
	whatRegex       = regexp.MustCompile("^[" + alphabet + "]{3}(-[" + alphabet + "]{3})?$")
)

func init() {
	ch3 := h3.NewH3()
	headerBits := fmt.Sprintf("%064s",
		strconv.FormatUint(uint64(ch3.FromGeo(h3.GeoCoord{Latitude: 0.0, Longitude: 0.0}, resolution)), 2))[:12]
	x, err := strconv.ParseUint(headerBits, 2, 64)
	if err != nil {
		panic(err)
	}
	fixHeaderInt = x * 1 << 52
	ch3.Close()
}

// FromH3Index converts an H3 index into a PlaceKey string.
func FromH3Index(index h3.Index) (string, error) {
	return FromH3Int(uint64(index))
}

// FromH3String converts an H3 hexadecimal string into a PlaceKey string.
func FromH3String(h3String string) (string, error) {
	x, err := strconv.ParseUint(h3String, 16, 64)
	if err != nil {
		return "", err
	}
	return FromH3Int(x)
}

// FromH3Int converts an H3 integer into a PlaceKey.
func FromH3Int(h3Int uint64) (string, error) {
	if inferResolution(h3Int) != resolution {
		return "", ErrInvalidResolution
	}
	return encodeH3Int(h3Int), nil
}

func fromH3IntUnvalidatedResolution(h3Int uint64) (string, error) {
	return encodeH3Int(h3Int), nil
}

// ToH3Index converts a PlaceKey string into an H3 index.
func ToH3Index(placeKey string) (h3.Index, error) {
	x, err := ToH3Int(placeKey)
	if err != nil {
		return 0, err
	}
	return h3.Index(x), nil
}

// ToH3String converts a PlaceKey string into an H3 string.
func ToH3String(placeKey string) (string, error) {
	x, err := ToH3Int(placeKey)
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(x, 16), nil
}

// ToH3Int converts a PlaceKey to an H3 integer.
func ToH3Int(placeKey string) (uint64, error) {
	_, where, err := parsePlacekey(placeKey)
	if err != nil {
		return 0, err
	}
	return decodeToH3Int(where), nil
}

// GetPrefixDistanceMap returns a map of the length of a shared PlaceKey prefix to the
// maximal distance in meters between two PlaceKeys sharing a prefix of that length.
func GetPrefixDistanceMap() map[int]float64 {
	return map[int]float64{
		1: 1.274e7,
		2: 2.777e6,
		3: 1.065e6,
		4: 1.524e5,
		5: 2.177e4,
		6: 8227.0,
		7: 1176.0,
		8: 444.3,
		9: 63.47,
	}
}

// FormatIsValid returns a boolean for whether or not the format of a PlaceKey
// is valid.
//
// It does not check the H3 validity, to do so use
// H3.IsValid(ToH3Index(placeKey)).
func FormatIsValid(placeKey string) bool {
	what, where, err := parsePlacekey(placeKey)
	if err != nil {
		return false
	}
	if what != "" {
		return wherePartIsValid(where) && whatPartIsValid(what)
	}
	return wherePartIsValid(where)
}

func whatPartIsValid(what string) bool {
	return whatRegex.MatchString(what)
}

func wherePartIsValid(where string) bool {
	return whereRegex.MatchString(where)
}

// split a PlaceKey in to what and where parts.
func parsePlacekey(placeKey string) (what, where string, err error) {
	if strings.Contains(placeKey, "@") {
		ww := strings.Split(placeKey, "@")
		if len(ww) != 2 {
			return "", "", ErrInvalidParts
		}
		return ww[0], ww[1], nil
	}
	return "", placeKey, nil
}

// encodeH3Int shortens an H3 integer to only include location data up to the
// base resolution.
func encodeH3Int(h3Int uint64) string {
	shortH3Int := shortenH3Int(h3Int)
	encodedShortH3 := encodeShortInt(shortH3Int)
	cleanEncodedShortH3 := cleanString(encodedShortH3)
	if len(cleanEncodedShortH3) <= codeLength {
		cleanEncodedShortH3 = strings.Repeat(paddingChar, codeLength-len(cleanEncodedShortH3)) + cleanEncodedShortH3
	}
	tuples := []string{}
	for i := 0; i < len(cleanEncodedShortH3); i += tupleLength {
		tuples = append(tuples, cleanEncodedShortH3[i:i+tupleLength])
	}
	return "@" + strings.Join(tuples, "-")
}

func encodeShortInt(x int64) string {
	if x == 0 {
		return string(alphabet[0])
	}
	res := ""
	for x > 0 {
		remainder := x % alphabetLength
		res = string(alphabet[remainder]) + res
		x /= alphabetLength
	}
	return res
}

func decodeToH3Int(wherePart string) uint64 {
	code := stripEncoding(wherePart)
	dirtyEncoding := dirtyString(code)
	shortH3Int := decodeString(dirtyEncoding)
	return unshortenH3Int(shortH3Int)
}

func decodeString(s string) int64 {
	var val int64
	for i := len(s) - 1; i >= 0; i-- {
		val += power64(alphabetLength, len(s)-1-i) * int64(strings.Index(alphabet, string(s[i])))
	}
	return val
}

// shortenH3Int shorten an H3 integer to only include location data up to the
// base resolution.
func shortenH3Int(h3Int uint64) int64 {
	// cuts off the 12 left-most bits that don't code location
	out := (h3Int + baseCellIncrement) % (1 << 52)
	// cuts off the rightmost bits corresponding to resolutions greater than the base resolution
	out >>= (3 * (15 - baseResolution))
	return int64(out)
}

func unshortenH3Int(shortH3Int int64) uint64 {
	unShiftedInt := shortH3Int << (3 * (15 - baseResolution))
	rebuiltInt := fixHeaderInt + unusedResolutionFiller - baseCellIncrement + uint64(unShiftedInt)
	return rebuiltInt
}

func stripEncoding(s string) string {
	s = strings.ReplaceAll(s, "@", "")
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, paddingChar, "")
	return s
}

func cleanString(s string) string {
	for k, v := range replacementMap {
		if strings.Contains(s, k) {
			s = strings.ReplaceAll(s, k, v)
		}
	}
	return s
}

func dirtyString(s string) string {
	for k, v := range replacementMap {
		// replacement should be in reversed order
		if strings.Contains(s, v) {
			s = strings.ReplaceAll(s, v, k)
		}
	}
	return s
}

func inferResolution(h3Int uint64) int {
	// resolution can be inferred from the number of digits in the cells bit
	// layout which are not 0b111, as the digit value can only be 0b111 when
	// that digit is greater than the resolution of the index
	cells := h3Int
	res := maxResolution
	for i := maxResolution; i >= 0; i-- {
		d := cells & 7 // 0b111
		if d != 7 {
			break
		}
		cells >>= 3
		res--
	}
	return res
}

func power64(base int64, exponent int) int64 {
	if exponent == 0 {
		return 1
	}
	return (base * power64(base, exponent-1))
}
