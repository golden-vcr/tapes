package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ListImages(t *testing.T) {
	tests := []struct {
		name         string
		c            *mockClient
		wantErr      string
		wantWarnings []Warning
		wantImages   []Image
	}{
		{
			"normal usage",
			&mockClient{metadataByFilename: map[string]Metadata{
				"0115_thumb.jpg": {},
				"0115_a.jpg":     {"Width": "817", "Height": "1499", "Color": "#fee", "Rotated": "false"},
				"0042_thumb.jpg": {},
				"0042_a.jpg":     {"Width": "700", "Height": "1500", "Color": "#febe99", "Rotated": "false"},
				"0042_b.jpg":     {"Width": "703", "Height": "1550", "Color": "#beb001", "Rotated": "true"},
			}},
			"",
			[]Warning{},
			[]Image{
				{
					Filename: "0042_thumb.jpg",
					TapeId:   42,
					Type:     ImageTypeThumbnail,
				},
				{
					Filename: "0042_a.jpg",
					TapeId:   42,
					Type:     ImageTypeGallery,
					GalleryData: &GalleryImageData{
						Index: 0,
						Metadata: &ImageMetadata{
							Width:   700,
							Height:  1500,
							Color:   "#febe99",
							Rotated: false,
						},
					},
				},
				{
					Filename: "0042_b.jpg",
					TapeId:   42,
					Type:     ImageTypeGallery,
					GalleryData: &GalleryImageData{
						Index: 1,
						Metadata: &ImageMetadata{
							Width:   703,
							Height:  1550,
							Color:   "#beb001",
							Rotated: true,
						},
					},
				},
				{
					Filename: "0115_thumb.jpg",
					TapeId:   115,
					Type:     ImageTypeThumbnail,
				},
				{
					Filename: "0115_a.jpg",
					TapeId:   115,
					Type:     ImageTypeGallery,
					GalleryData: &GalleryImageData{
						Index: 0,
						Metadata: &ImageMetadata{
							Width:   817,
							Height:  1499,
							Color:   "#fee",
							Rotated: false,
						},
					},
				},
			},
		},
		{
			"files with invalid/non-image filenames are logged as warnings and ignored",
			&mockClient{metadataByFilename: map[string]Metadata{
				"0041_thumb.jpg":         {},
				"0041_a.jpg":             {"Width": "701", "Height": "1501", "Color": "#aabbcc", "Rotated": "true"},
				"0042_thumb.jpg":         {},
				"0042_a.jpg":             {"Width": "700", "Height": "1500", "Color": "#febe99", "Rotated": "false"},
				"0043_somethingelse.jpg": {},
				"whatever.txt":           {},
			}},
			"",
			[]Warning{
				{
					Filename: "0043_somethingelse.jpg",
					Message:  "not a valid image filename matching ^(\\d{4})_(thumb|[a-z])\\.jpg$",
				},
				{
					Filename: "whatever.txt",
					Message:  "not a valid image filename matching ^(\\d{4})_(thumb|[a-z])\\.jpg$",
				},
			},
			[]Image{
				{
					Filename: "0041_thumb.jpg",
					TapeId:   41,
					Type:     ImageTypeThumbnail,
				},
				{
					Filename: "0041_a.jpg",
					TapeId:   41,
					Type:     ImageTypeGallery,
					GalleryData: &GalleryImageData{
						Index: 0,
						Metadata: &ImageMetadata{
							Width:   701,
							Height:  1501,
							Color:   "#aabbcc",
							Rotated: true,
						},
					},
				},
				{
					Filename: "0042_thumb.jpg",
					TapeId:   42,
					Type:     ImageTypeThumbnail,
				},
				{
					Filename: "0042_a.jpg",
					TapeId:   42,
					Type:     ImageTypeGallery,
					GalleryData: &GalleryImageData{
						Index: 0,
						Metadata: &ImageMetadata{
							Width:   700,
							Height:  1500,
							Color:   "#febe99",
							Rotated: false,
						},
					},
				},
			},
		},
		{
			"invalid metadata is logged as a warning and excludes affected tape id",
			&mockClient{metadataByFilename: map[string]Metadata{
				"0041_thumb.jpg": {},
				"0041_a.jpg":     {"Width": "701", "Color": "#aabbcc", "Rotated": "true"},
				"0042_thumb.jpg": {},
				"0042_a.jpg":     {"Width": "700", "Height": "1500", "Color": "#febe99", "Rotated": "false"},
			}},
			"",
			[]Warning{
				{
					Filename: "0041_a.jpg",
					Message:  "metadata value 'Height' is required",
				},
			},
			[]Image{
				{
					Filename: "0042_thumb.jpg",
					TapeId:   42,
					Type:     ImageTypeThumbnail,
				},
				{
					Filename: "0042_a.jpg",
					TapeId:   42,
					Type:     ImageTypeGallery,
					GalleryData: &GalleryImageData{
						Index: 0,
						Metadata: &ImageMetadata{
							Width:   700,
							Height:  1500,
							Color:   "#febe99",
							Rotated: false,
						},
					},
				},
			},
		},
		{
			"tape with gallery image but no thumbnail image is ignored",
			&mockClient{metadataByFilename: map[string]Metadata{
				"0041_a.jpg":     {"Width": "701", "Height": "1501", "Color": "#aabbcc", "Rotated": "true"},
				"0042_thumb.jpg": {},
				"0042_a.jpg":     {"Width": "700", "Height": "1500", "Color": "#febe99", "Rotated": "false"},
			}},
			"",
			[]Warning{
				{
					Filename: "0041_a.jpg",
					Message:  "tape 41 has gallery image(s) but no accompanying thumbnail image",
				},
			},
			[]Image{
				{
					Filename: "0042_thumb.jpg",
					TapeId:   42,
					Type:     ImageTypeThumbnail,
				},
				{
					Filename: "0042_a.jpg",
					TapeId:   42,
					Type:     ImageTypeGallery,
					GalleryData: &GalleryImageData{
						Index: 0,
						Metadata: &ImageMetadata{
							Width:   700,
							Height:  1500,
							Color:   "#febe99",
							Rotated: false,
						},
					},
				},
			},
		},
		{
			"tape with thumbnail image but no gallery images is ignored",
			&mockClient{metadataByFilename: map[string]Metadata{
				"0041_thumb.jpg": {},
				"0042_thumb.jpg": {},
				"0042_a.jpg":     {"Width": "700", "Height": "1500", "Color": "#febe99", "Rotated": "false"},
			}},
			"",
			[]Warning{
				{
					Filename: "0041_thumb.jpg",
					Message:  "tape 41 has thumbnail image but no accompanying gallery image(s)",
				},
			},
			[]Image{
				{
					Filename: "0042_thumb.jpg",
					TapeId:   42,
					Type:     ImageTypeThumbnail,
				},
				{
					Filename: "0042_a.jpg",
					TapeId:   42,
					Type:     ImageTypeGallery,
					GalleryData: &GalleryImageData{
						Index: 0,
						Metadata: &ImageMetadata{
							Width:   700,
							Height:  1500,
							Color:   "#febe99",
							Rotated: false,
						},
					},
				},
			},
		},
		{
			"empty bucket lists 0 images without error",
			&mockClient{},
			"",
			[]Warning{},
			[]Image{},
		},
		{
			"failure to list filenames is an error",
			&mockClient{
				listFilenamesErr: fmt.Errorf("mock error"),
			},
			"failed to list filenames from storage bucket: mock error",
			[]Warning{},
			[]Image{},
		},
		{
			"failure to get file metadata is an error",
			&mockClient{
				getFileMetadataErr: fmt.Errorf("mock error"),
				metadataByFilename: map[string]Metadata{
					"0042_thumb.jpg": {},
					"0042_a.jpg":     {"Width": "700", "Height": "1500", "Color": "#febe99", "Rotated": "false"},
				},
			},
			"failed to get metadata for image file 0042_a.jpg: mock error",
			[]Warning{},
			[]Image{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			images, warnings, err := ListImages(context.Background(), tt.c)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.wantErr)
				assert.Nil(t, warnings)
				assert.Nil(t, images)
			} else {
				assert.NoError(t, err)
				assert.ElementsMatch(t, tt.wantWarnings, warnings)
				assert.Equal(t, tt.wantImages, images)
			}
		})
	}
}

type mockClient struct {
	listFilenamesErr   error
	getFileMetadataErr error
	metadataByFilename map[string]Metadata
}

func (m *mockClient) ListFilenames(ctx context.Context) ([]string, error) {
	if m.listFilenamesErr != nil {
		return nil, m.listFilenamesErr
	}
	filenames := make([]string, 0, len(m.metadataByFilename))
	for filename := range m.metadataByFilename {
		filenames = append(filenames, filename)
	}
	return filenames, nil
}

func (m *mockClient) GetFileMetadata(ctx context.Context, filename string) (Metadata, error) {
	if m.getFileMetadataErr != nil {
		return nil, m.getFileMetadataErr
	}
	metadata, ok := m.metadataByFilename[filename]
	if !ok {
		return nil, fmt.Errorf("no such file")
	}
	return metadata, nil
}

var _ Client = (*mockClient)(nil)
