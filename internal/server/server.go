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
	mux.HandleFunc("/tapes", s.handleGetTapeListing)
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
	imageCounts, err := s.bucket.GetImageCounts(req.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	items := make([]TapeListingItem, 0, len(rows))
	for _, row := range rows {
		numImages, ok := imageCounts[row.ID]
		if !ok || numImages <= 0 {
			continue
		}
		imageFilenames := make([]string, 0, numImages)
		for imageIndex := 0; imageIndex < numImages; imageIndex++ {
			imageFilenames = append(imageFilenames, bucket.GetImageKey(row.ID, imageIndex))
		}
		items = append(items, TapeListingItem{
			Id:                     row.ID,
			Title:                  row.Title,
			Year:                   row.Year,
			RuntimeMinutes:         row.RuntimeMin,
			ThumbnailImageFilename: bucket.GetThumbnailKey(row.ID),
			ImageFilenames:         imageFilenames,
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
