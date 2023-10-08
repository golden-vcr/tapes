package storage

import (
	"errors"
	"regexp"
)

// hexColorRegexp matches a 3-digit or 6-digit hex color string, with a hash prepended
var hexColorRegexp = regexp.MustCompile(`^#[a-zA-Z0-9]{3}(?:[a-zA-Z0-9]{3})?$`)

// ErrInvalidHexColor is returned when parsing fails
var ErrInvalidHexColor = errors.New("invalid hex color")

// HexColor is a string value representing a color, encoded as hex
type HexColor string

// parseHexColor ensures that the given string is a valid hex color, returning an error
// if not
func parseHexColor(s string) (HexColor, error) {
	if !hexColorRegexp.MatchString(s) {
		return "#cccccc", ErrInvalidHexColor
	}
	return HexColor(s), nil
}
