package storage

import (
	"fmt"
	"regexp"
	"strconv"
)

var imageFilenameRegex = regexp.MustCompile(`^(\d{4})_(thumb|[a-z])\.jpg$`)

// imageId describes the details of an image file as encoded in the filename
type imageId struct {
	// tapeId is the integer ID of the tape this image was scanned from
	tapeId int
	// imageType indicates whether this image is a low-res thumbnail or a gallery image
	imageType ImageType
	// galleryIndex indicates an image's ordering in the gallery, if type is "gallery"
	galleryIndex int
}

// parseImageFilename parses information about an image from its filename, returning an
// error if the given string is not a valid image filename
func parseImageFilename(s string) (*imageId, error) {
	match := imageFilenameRegex.FindStringSubmatch(s)
	if match != nil {
		tapeId, _ := strconv.Atoi(match[1])
		if match[2] == "thumb" {
			return &imageId{
				tapeId:    tapeId,
				imageType: ImageTypeThumbnail,
			}, nil
		}
		ord := match[2][0]
		galleryIndex := ord - 'a'
		return &imageId{
			tapeId:       tapeId,
			imageType:    ImageTypeGallery,
			galleryIndex: int(galleryIndex),
		}, nil
	}
	return nil, fmt.Errorf("not a valid image filename matching %s", imageFilenameRegex.String())
}

// GetImageFilename reconstructs the filename associated with an image
func GetImageFilename(tapeId int, imageType ImageType, galleryIndex int) string {
	if imageType == ImageTypeThumbnail {
		return fmt.Sprintf("%04d_thumb.jpg", tapeId)
	}
	ord := 'a' + min(galleryIndex, 25)
	return fmt.Sprintf("%04d_%c.jpg", tapeId, ord)
}
