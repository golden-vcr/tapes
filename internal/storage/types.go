package storage

// ImageType dictates how an image will be used and displayed in the frontend
type ImageType string

const (
	// ImageTypeThumbnail identifies a single, low-res image that can be displayed to
	// represent an entire tape
	ImageTypeThumbnail ImageType = "thumbnail"
	// ImageTypeGallery identifies one in a series of full-res images scanned from a
	// tape, with a particular ordering: typically front of case, then back of case,
	// then front of tape, then any additional scans
	ImageTypeGallery ImageType = "gallery"
)

// Image is a single image file, stored in an S3-compatible bucket, that was scanned
// from a particular tape
type Image struct {
	// Filename is the image filename, i.e. the S3 object key
	Filename string
	// TapeId is the integer ID of the tape this image was scanned from
	TapeId int
	// Type indicates whether this image is a low-res thumbnail or a gallery image
	Type ImageType
	// GalleryImageData includes additional metadata for images of type gallery
	GalleryData *GalleryImageData
}

// GalleryImageData specifies extra data for an image of type gallery
type GalleryImageData struct {
	// Index indicates an image's ordering in the gallery
	Index int
	// Metadata describes the image's width, height, dominant color, and other
	// information required to render the image in the gallery
	Metadata *ImageMetadata
}

// ImageMetadata provides additional data required to render the image in the webapp, as
// encoded in the file metadata (i.e. S3 x-amz-meta-* headers)
type ImageMetadata struct {
	// Width is the width of the image in pixels
	Width int
	// Height is the height of the image in pixels
	Height int
	// Color is a hex-formatted string representing the dominant color in the image
	Color HexColor
	// Rotated is true if the image has been rotated 90 degrees CCW in order to have a
	// vertical aspect ratio: if so, it may be rotated 90 degrees CW to be displayed to
	// the user with the text in a readable orientation
	Rotated bool
}
