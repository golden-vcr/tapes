package catalog

type Listing struct {
	ImageHostUrl string `json:"imageHost"`
	Items        []Item `json:"items"`
}

type Item struct {
	Id                     int            `json:"id"`
	Title                  string         `json:"title"`
	Year                   int            `json:"year"`
	RuntimeInMinutes       int            `json:"runtime"`
	ThumbnailImageFilename string         `json:"thumbnail"`
	Images                 []GalleryImage `json:"images"`
}

type GalleryImage struct {
	Filename string `json:"filename"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Color    string `json:"color"`
	Rotated  bool   `json:"rotated"`
}