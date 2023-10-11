package db

import (
	"encoding/json"
	"fmt"
)

// TapeImage is the JSON format used by the GetTapes query when returning data about the
// images for a tape
type TapeImage struct {
	Index   int32  `json:"index"`
	Color   string `json:"color"`
	Width   int32  `json:"width"`
	Height  int32  `json:"height"`
	Rotated bool   `json:"rotated"`
}

// ParseTapeImageArray accepts a JSON-formatted array of objects represented tape
// images, as returned by the GetTapes query
func ParseTapeImageArray(data json.RawMessage) ([]TapeImage, error) {
	var images []TapeImage
	if err := json.Unmarshal(data, &images); err != nil {
		return nil, fmt.Errorf("failed to parse TapeImage array from JSON data: %v", err)
	}
	return images, nil
}

// ParseTapeTagsArray accepts a JSON-formatted array of strings, representing the list
// of tag names returned by the GetTapes query
func ParseTapeTagsArray(data json.RawMessage) ([]string, error) {
	var tags []string
	if err := json.Unmarshal(data, &tags); err != nil {
		return nil, fmt.Errorf("failed to parse tag name array from JSON data: %v", err)
	}
	return tags, nil
}
