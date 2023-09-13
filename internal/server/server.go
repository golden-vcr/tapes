package server

import (
	"encoding/json"
	"net/http"

	"github.com/golden-vcr/tapes/internal/bucket"
	"github.com/golden-vcr/tapes/internal/sheets"
)

type Server struct {
	http.Handler

	sheets *sheets.Client
	bucket *bucket.Client
}

func New(sheetsClient *sheets.Client, bucketClient *bucket.Client) *Server {
	s := &Server{
		sheets: sheetsClient,
		bucket: bucketClient,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleGetTapeListing)
	s.Handler = mux
	return s
}

func (s *Server) handleGetTapeListing(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	rows, err := s.sheets.ListTapes(req.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	imageDataByTapeId := s.bucket.GetImageData(req.Context())

	items := make([]TapeListingItem, 0, len(rows))
	for _, row := range rows {
		imageData, ok := imageDataByTapeId[row.ID]
		if !ok || len(imageData) == 0 {
			continue
		}
		images := make([]TapeImageData, 0, len(imageData))
		for i := range imageData {
			images = append(images, TapeImageData{
				Filename: imageData[i].Filename,
				Width:    imageData[i].Width,
				Height:   imageData[i].Height,
				Color:    imageData[i].Color,
				Rotated:  imageData[i].Rotated,
			})
		}
		items = append(items, TapeListingItem{
			Id:                     row.ID,
			Title:                  row.Title,
			Year:                   row.Year,
			RuntimeMinutes:         row.RuntimeMin,
			ThumbnailImageFilename: bucket.GetThumbnailKey(row.ID),
			Images:                 images,
		})
	}

	result := TapeListing{
		Tapes:        items,
		ImageHostUrl: s.bucket.GetImageHostURL(),
	}
	if err := json.NewEncoder(res).Encode(result); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
