package catalog

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/golden-vcr/tapes/gen/queries"
	"github.com/golden-vcr/tapes/internal/db"
	"github.com/golden-vcr/tapes/internal/storage"
	"github.com/golden-vcr/tapes/internal/users"
	"github.com/gorilla/mux"
)

type Queries interface {
	GetTapes(ctx context.Context) ([]queries.GetTapesRow, error)
	GetTape(ctx context.Context, tapeID int32) (queries.GetTapeRow, error)
	GetTapeContributorIds(ctx context.Context) ([]string, error)
}

type Server struct {
	q            Queries
	lookup       users.Lookup
	imageHostUrl string
}

func NewServer(q Queries, lookup users.Lookup, imageHostUrl string) *Server {
	return &Server{
		q:            q,
		lookup:       lookup,
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
	userIds, err := s.q.GetTapeContributorIds(req.Context())
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.lookup.Resolve(req.Context(), userIds); err != nil {
		fmt.Printf("Error resolving contributor usernames: %v\n", err)
	}

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
		contributorName := ""
		if row.ContributorID.Valid {
			contributorName = s.lookup.GetDisplayName(row.ContributorID.String)
		}
		items = append(items, Item{
			Id:                     int(row.ID),
			Title:                  row.Title,
			Year:                   year,
			RuntimeInMinutes:       runtime,
			ThumbnailImageFilename: storage.GetImageFilename(int(row.ID), storage.ImageTypeThumbnail, -1),
			SeriesName:             row.SeriesName,
			ContributorName:        contributorName,
			NumFavorites:           int(row.NumFavorites),
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

	contributorName := ""
	if row.ContributorID.Valid {
		if err := s.lookup.Resolve(req.Context(), []string{row.ContributorID.String}); err != nil {
			fmt.Printf("Error resolving contributor username: %v\n", err)
		}
		contributorName = s.lookup.GetDisplayName(row.ContributorID.String)
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
		SeriesName:             row.SeriesName,
		ContributorName:        contributorName,
		NumFavorites:           int(row.NumFavorites),
		Images:                 galleryImages,
		Tags:                   row.Tags,
	}
	if err := json.NewEncoder(res).Encode(item); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}
