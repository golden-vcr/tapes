package bucket

type ImageData struct {
	// Filename is the filename of the image, i.e. its object key in the images bucket
	Filename string `json:"filename"`
	// Width is the width of the image in pixels
	Width int `json:"width"`
	// Height is the height of the image in pixels
	Height int `json:"height"`
	// Color is a hex-formatted color (e.g. "#ffcc99") indicating the dominant color in
	// the image
	Color string `json:"color"`
	// Rotated is true if the image has been rotated 90 degrees CCW in order to have a
	// vertical aspect ratio: if so, it may be rotated 90 degrees CW to be displayed to
	// the user with the text in a readable orientation
	Rotated bool `json:"rotated"`
}
