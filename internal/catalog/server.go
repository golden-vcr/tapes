package catalog

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/golden-vcr/tapes/gen/queries"
	"github.com/golden-vcr/tapes/internal/db"
	"github.com/golden-vcr/tapes/internal/storage"
	"github.com/gorilla/mux"
)

type Queries interface {
	GetTapes(ctx context.Context) ([]queries.GetTapesRow, error)
}

type Server struct {
	q            Queries
	imageHostUrl string
}

func NewServer(q Queries, imageHostUrl string) *Server {
	return &Server{
		q:            q,
		imageHostUrl: imageHostUrl,
	}
}

func (s *Server) RegisterRoutes(r *mux.Router) {
	for _, root := range []string{"", "/"} {
		r.Path(root).Methods("GET").HandlerFunc(s.handleGetListing)
	}
}

func (s *Server) handleGetListing(res http.ResponseWriter, req *http.Request) {
	rows, err := s.q.GetTapes(req.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	items := make([]Item, 0, len(rows))
	for _, row := range rows {
		images, err := db.ParseTapeImageArray(row.Images)
		if err != nil {
			http.Error(res, err.Error(), http.StatusInternalServerError)
			return
		}

		galleryImages := make([]GalleryImage, 0, len(images))
		for _, image := range images {
			galleryImages = append(galleryImages, GalleryImage{
				Filename: storage.GetImageFilename(int(row.ID), storage.ImageTypeGallery, int(image.Index)),
				Width:    int(image.Width),
				Height:   int(image.Height),
				Color:    image.Color,
				Rotated:  image.Rotated,
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
		items = append(items, Item{
			Id:                     int(row.ID),
			Title:                  row.Title,
			Year:                   year,
			RuntimeInMinutes:       runtime,
			ThumbnailImageFilename: storage.GetImageFilename(int(row.ID), storage.ImageTypeThumbnail, -1),
			Images:                 galleryImages,
			Tags:                   row.Tags,
		})
	}

	result := Listing{
		ImageHostUrl: s.imageHostUrl,
		Items:        items,
	}
	if err := json.NewEncoder(res).Encode(result); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
