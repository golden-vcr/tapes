package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/golden-vcr/tapes/gen/queries"
	"github.com/golden-vcr/tapes/internal/bucket"
)

type Server struct {
	http.Handler

	q         *queries.Queries
	bucketUrl string
}

func New(q *queries.Queries, bucketUrl string) *Server {
	s := &Server{
		q:         q,
		bucketUrl: bucketUrl,
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

	rows, err := s.q.GetTapes(req.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	items := make([]TapeListingItem, 0, len(rows))
	for _, row := range rows {
		var imageData []imageResult
		if err := json.Unmarshal(row.Images, &imageData); err != nil {
			fmt.Printf("Failed to unmarshal JSON image data from DB result for tape %d: %v\n", row.ID, err)
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		images := make([]TapeImageData, 0, len(imageData))
		for i := range imageData {
			images = append(images, TapeImageData{
				Filename: bucket.GetImageKey(int(row.ID), int(imageData[i].Index)),
				Width:    int(imageData[i].Width),
				Height:   int(imageData[i].Height),
				Color:    imageData[i].Color,
				Rotated:  imageData[i].Rotated,
			})
		}
		year := 0
		if row.Year.Valid {
			year = int(row.Year.Int32)
		}
		runtime := 0
		if row.Runtime.Valid {
			runtime = int(row.Runtime.Int32)
		}
		items = append(items, TapeListingItem{
			Id:                     int(row.ID),
			Title:                  row.Title,
			Year:                   year,
			RuntimeMinutes:         runtime,
			ThumbnailImageFilename: bucket.GetThumbnailKey(int(row.ID)),
			Images:                 images,
		})
	}

	result := TapeListing{
		Tapes:        items,
		ImageHostUrl: s.bucketUrl,
	}
	if err := json.NewEncoder(res).Encode(result); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

type imageResult struct {
	Index   int32  `json:"index"`
	Color   string `json:"color"`
	Width   int32  `json:"width"`
	Height  int32  `json:"height"`
	Rotated bool   `json:"rotated"`
}
