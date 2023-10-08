package storage

import (
	"fmt"
	"strconv"
)

// toImageMetadata parses the key/value pairs returned as S3 metadata into a valid
// ImageMetadata struct, or returns an error if any required values are missing or
// invalid
func (md Metadata) toImageMetadata() (*ImageMetadata, error) {
	// 'Width' and 'Height' must be specified as positive ints
	width, err := md.parsePositiveInt("Width")
	if err != nil {
		return nil, err
	}
	height, err := md.parsePositiveInt("Height")
	if err != nil {
		return nil, err
	}

	// 'Color' must be specified as a hex string
	colorStr, ok := md["Color"]
	if !ok {
		return nil, fmt.Errorf("metadata value 'Color' is required")
	}
	color, err := parseHexColor(colorStr)
	if err != nil {
		return nil, fmt.Errorf("metadata value 'Color' must be a hex color (got '%s')", colorStr)
	}

	// 'Rotated' must be specified as a boolean
	rotatedStr, ok := md["Rotated"]
	if !ok {
		return nil, fmt.Errorf("metadata value 'Rotated' is required")
	}
	if rotatedStr != "true" && rotatedStr != "false" {
		return nil, fmt.Errorf("metadata value 'Rotated' must be a bool (got '%s')", rotatedStr)
	}
	rotated := rotatedStr == "true"

	return &ImageMetadata{
		Width:   width,
		Height:  height,
		Color:   color,
		Rotated: rotated,
	}, nil
}

func (md Metadata) parsePositiveInt(name string) (int, error) {
	strValue, ok := md[name]
	if !ok {
		return -1, fmt.Errorf("metadata value '%s' is required", name)
	}
	value, err := strconv.Atoi(strValue)
	if err != nil {
		return -1, fmt.Errorf("metadata value '%s' must be an integer (got '%s')", name, strValue)
	}
	if value <= 0 {
		return -1, fmt.Errorf("metadata value '%s' must be positive (got %d)", name, value)
	}
	return value, nil
}
