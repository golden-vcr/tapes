package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseImageFilename(t *testing.T) {
	tests := []struct {
		filename string
		want     *imageId
	}{
		{"", nil},
		{"34_thumb.jpg", nil},
		{"0034.jpg", nil},
		{"0034_thumb.jpg", &imageId{34, ImageTypeThumbnail, 0}},
		{"0034_a.jpg", &imageId{34, ImageTypeGallery, 0}},
		{"0034_b.jpg", &imageId{34, ImageTypeGallery, 1}},
		{"0034_c.jpg", &imageId{34, ImageTypeGallery, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got, err := parseImageFilename(tt.filename)
			if tt.want == nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_GetImageFilename(t *testing.T) {
	tests := []struct {
		name         string
		tapeId       int
		imageType    ImageType
		galleryIndex int
		want         string
	}{
		{
			"thumbnail image",
			123,
			ImageTypeThumbnail,
			0,
			"0123_thumb.jpg",
		},
		{
			"gallery image 0",
			123,
			ImageTypeGallery,
			0,
			"0123_a.jpg",
		},
		{
			"gallery image 1",
			123,
			ImageTypeGallery,
			1,
			"0123_b.jpg",
		},
		{
			"gallery past the 26th are clamped at z",
			123,
			ImageTypeGallery,
			99999,
			"0123_z.jpg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetImageFilename(tt.tapeId, tt.imageType, tt.galleryIndex)
			assert.Equal(t, tt.want, got)
		})
	}
}
