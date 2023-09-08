package server

type TapeListing struct {
	Tapes        []TapeListingItem `json:"tapes"`
	ImageHostUrl string            `json:"imageHostUrl"`
}

type TapeListingItem struct {
	Id                     int      `json:"id"`
	Title                  string   `json:"title"`
	Year                   int      `json:"year"`
	RuntimeMinutes         int      `json:"runtimeMinutes"`
	Color                  string   `json:"color"`
	ThumbnailImageFilename string   `json:"thumbnailImageFilename"`
	ImageFilenames         []string `json:"imageFilenames"`
}
