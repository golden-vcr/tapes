package favorites

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golden-vcr/auth"
	"github.com/golden-vcr/tapes/gen/queries"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
)

type Server struct {
	q Queries
}

func NewServer(q *queries.Queries) *Server {
	return &Server{
		q: q,
	}
}

func (s *Server) RegisterRoutes(c auth.Client, r *mux.Router) {
	// Require viewer-level access for routes that keep track of users' favorite tapes
	r.Use(func(next http.Handler) http.Handler {
		return auth.RequireAccess(c, auth.RoleViewer, next)
	})

	// GET /favorites returns the list of IDs that the auth'd user has previously marked
	// as favorites; PATCH /favorite allows a single tape to be flagged or unflagged as
	// a favorite for the auth'd user
	for _, root := range []string{"", "/"} {
		r.Path(root).Methods("GET").HandlerFunc(s.handleGetFavorites)
		r.Path(root).Methods("PATCH").HandlerFunc(s.handlePatchFavorites)
	}
}

func (s *Server) handleGetFavorites(res http.ResponseWriter, req *http.Request) {
	// Identify the user from their authorization token
	claims, err := auth.GetClaims(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get a sorted list of favorite tape IDs from the db
	tapeIds, err := s.q.GetFavoriteTapes(req.Context(), claims.User.Id)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return that list as a JSON-encoded FavoriteTapeSet struct
	result := FavoriteTapeSet{
		TapeIds: make([]int, 0, len(tapeIds)),
	}
	for _, tapeId := range tapeIds {
		result.TapeIds = append(result.TapeIds, int(tapeId))
	}
	if err := json.NewEncoder(res).Encode(result); err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handlePatchFavorites(res http.ResponseWriter, req *http.Request) {
	// Identify the user from their authorization token
	claims, err := auth.GetClaims(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// The request's Content-Type must indicate JSON if set
	contentType := req.Header.Get("content-type")
	if contentType != "" && !strings.HasPrefix(contentType, "application/json") {
		http.Error(res, "content-type not supported", http.StatusBadRequest)
		return
	}

	// Parse the favorite change from the body
	var payload FavoriteTapeChange
	if err := json.NewDecoder(req.Body).Decode(&payload); err != nil {
		http.Error(res, fmt.Sprintf("invalid request payload: %v", err), http.StatusBadRequest)
		return
	}

	// Update the database, and handle foreign-key constraint violations (libpq error
	// code 23503) as a 400; anything else as a 500
	var dbErr error
	if payload.IsFavorite {
		dbErr = s.q.RegisterFavoriteTape(req.Context(), queries.RegisterFavoriteTapeParams{
			TwitchUserID: claims.User.Id,
			TapeID:       int32(payload.TapeId),
		})
	} else {
		dbErr = s.q.UnregisterFavoriteTape(req.Context(), queries.UnregisterFavoriteTapeParams{
			TwitchUserID: claims.User.Id,
			TapeID:       int32(payload.TapeId),
		})
	}
	if dbErr != nil {
		if pqErr, ok := dbErr.(*pq.Error); ok {
			if pqErr.Code.Name() == "foreign_key_violation" {
				http.Error(res, "no such tape", http.StatusBadRequest)
				return
			}
		}
		http.Error(res, dbErr.Error(), http.StatusInternalServerError)
		return
	}

	res.WriteHeader(http.StatusNoContent)
}
