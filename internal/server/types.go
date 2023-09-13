package server

type TapeListing struct {
	Tapes        []TapeListingItem `json:"tapes"`
	ImageHostUrl string            `json:"imageHostUrl"`
}

type TapeListingItem struct {
	Id                     int             `json:"id"`
	Title                  string          `json:"title"`
	Year                   int             `json:"year"`
	RuntimeMinutes         int             `json:"runtimeMinutes"`
	Color                  string          `json:"color"`
	ThumbnailImageFilename string          `json:"thumbnailImageFilename"`
	Images                 []TapeImageData `json:"images"`
}

type TapeImageData struct {
	Filename string `json:"filename"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Color    string `json:"color"`
	Rotated  bool   `json:"rotated"`
}
