package admin

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/golden-vcr/auth"
	"github.com/golden-vcr/tapes/gen/queries"
	"github.com/gorilla/mux"
)

type Queries interface {
	ApplySeries(ctx context.Context, arg queries.ApplySeriesParams) (sql.Result, error)
}

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
		return auth.RequireAccess(c, auth.RoleBroadcaster, next)
	})

	r.Path("/apply-series").Methods("POST").HandlerFunc(s.handleApplySeries)
}

func (s *Server) handleApplySeries(res http.ResponseWriter, req *http.Request) {
	// Identify the user from their authorization token
	_, err := auth.GetClaims(req)
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get args from request body
	if err := req.ParseForm(); err != nil {
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	seriesName := req.PostForm.Get("seriesName")
	if seriesName == "" {
		http.Error(res, "seriesName is required", http.StatusBadRequest)
		return
	}
	tapeIdsStr := req.PostForm.Get("tapeIds")
	if tapeIdsStr == "" {
		http.Error(res, "tapeIds is required", http.StatusBadRequest)
		return
	}

	// Parse list of target tape IDs
	tapeIds := make([]int32, 0)
	for _, token := range strings.Split(tapeIdsStr, ",") {
		tapeId, err := strconv.Atoi(token)
		if err != nil {
			http.Error(res, "tapeIds must be supplied as a comma-delimited list of integers", http.StatusBadRequest)
			return
		}
		tapeIds = append(tapeIds, int32(tapeId))
	}
	if len(tapeIds) == 0 {
		http.Error(res, "no target tapes", http.StatusBadRequest)
		return
	}

	// Update the seriesName value for all target tapes
	_, err = s.q.ApplySeries(req.Context(), queries.ApplySeriesParams{
		SeriesName: seriesName,
		TapeIds:    tapeIds,
	})
	if err != nil {
		http.Error(res, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return 204
	res.WriteHeader(http.StatusNoContent)
}
