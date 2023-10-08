package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Metadata_toImageMetadata(t *testing.T) {
	tests := []struct {
		name    string
		md      Metadata
		wantErr string
		want    *ImageMetadata
	}{
		{
			"width is required",
			Metadata{
				"Height":  "1500",
				"Color":   "#fe99cc",
				"Rotated": "true",
			},
			"metadata value 'Width' is required",
			nil,
		},
		{
			"height is required",
			Metadata{
				"Width":   "700",
				"Color":   "#fe99cc",
				"Rotated": "true",
			},
			"metadata value 'Height' is required",
			nil,
		},
		{
			"color is required",
			Metadata{
				"Width":   "700",
				"Height":  "1500",
				"Rotated": "true",
			},
			"metadata value 'Color' is required",
			nil,
		},
		{
			"rotated is required",
			Metadata{
				"Width":  "700",
				"Height": "1500",
				"Color":  "#fe99cc",
			},
			"metadata value 'Rotated' is required",
			nil,
		},
		{
			"normal usage",
			Metadata{
				"Width":   "700",
				"Height":  "1500",
				"Color":   "#fe99cc",
				"Rotated": "true",
			},
			"",
			&ImageMetadata{
				Width:   700,
				Height:  1500,
				Color:   "#fe99cc",
				Rotated: true,
			},
		},
		{
			"width must be positive",
			Metadata{
				"Width":   "-50",
				"Height":  "1500",
				"Color":   "#fe99cc",
				"Rotated": "true",
			},
			"metadata value 'Width' must be positive (got -50)",
			nil,
		},
		{
			"width must be an integer",
			Metadata{
				"Width":   "foo",
				"Height":  "1500",
				"Color":   "#fe99cc",
				"Rotated": "true",
			},
			"metadata value 'Width' must be an integer (got 'foo')",
			nil,
		},
		{
			"height must be positive",
			Metadata{
				"Width":   "700",
				"Height":  "0",
				"Color":   "#fe99cc",
				"Rotated": "true",
			},
			"metadata value 'Height' must be positive (got 0)",
			nil,
		},
		{
			"height must be an integer",
			Metadata{
				"Width":   "700",
				"Height":  "bar",
				"Color":   "#fe99cc",
				"Rotated": "true",
			},
			"metadata value 'Height' must be an integer (got 'bar')",
			nil,
		},
		{
			"color must be a hex color",
			Metadata{
				"Width":   "700",
				"Height":  "1500",
				"Color":   "blue",
				"Rotated": "true",
			},
			"metadata value 'Color' must be a hex color (got 'blue')",
			nil,
		},
		{
			"rotated must be a bool",
			Metadata{
				"Width":   "700",
				"Height":  "1500",
				"Color":   "#fe99cc",
				"Rotated": "maybe",
			},
			"metadata value 'Rotated' must be a bool (got 'maybe')",
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.md.toImageMetadata()
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
