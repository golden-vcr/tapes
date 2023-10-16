package catalog

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/golden-vcr/tapes/gen/queries"
	"github.com/golden-vcr/tapes/internal/db"
	"github.com/golden-vcr/tapes/internal/storage"
	"github.com/gorilla/mux"
)

type Queries interface {
	GetTapes(ctx context.Context) ([]queries.GetTapesRow, error)
	GetTape(ctx context.Context, tapeID int32) (queries.GetTapeRow, error)
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
	r.Path("/{id}").Methods("GET").HandlerFunc(s.handleGetDetails)
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

func (s *Server) handleGetDetails(res http.ResponseWriter, req *http.Request) {
	tapeIdStr, ok := mux.Vars(req)["id"]
	if !ok || tapeIdStr == "" {
		http.Error(res, "failed to parse 'id' from URL", http.StatusInternalServerError)
		return
	}
	tapeId, err := strconv.Atoi(tapeIdStr)
	if err != nil {
		http.Error(res, "tape ID must be an integer", http.StatusBadRequest)
		return
	}

	row, err := s.q.GetTape(req.Context(), int32(tapeId))
	if errors.Is(err, sql.ErrNoRows) {
		http.Error(res, "no such tape", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

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
	item := Item{
		Id:                     int(row.ID),
		Title:                  row.Title,
		Year:                   year,
		RuntimeInMinutes:       runtime,
		ThumbnailImageFilename: storage.GetImageFilename(int(row.ID), storage.ImageTypeThumbnail, -1),
		Images:                 galleryImages,
		Tags:                   row.Tags,
	}
	if err := json.NewEncoder(res).Encode(item); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
