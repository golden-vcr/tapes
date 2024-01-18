package favorites

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/golden-vcr/auth"
	authmock "github.com/golden-vcr/auth/mock"
	"github.com/golden-vcr/tapes/gen/queries"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func Test_Server_handleGetFavorites(t *testing.T) {
	tests := []struct {
		name       string
		q          *mockQueries
		wantStatus int
		wantBody   string
	}{
		{
			"with no favorites recorded, result is empty set",
			&mockQueries{},
			http.StatusOK,
			`{"tapeIds":[]}`,
		},
		{
			"favorite tape IDs are returned from DB",
			&mockQueries{
				favorites: []queries.RegisterFavoriteTapeParams{
					{
						TwitchUserID: "54321",
						TapeID:       1,
					},
					{
						TwitchUserID: "54321",
						TapeID:       3,
					},
					{
						TwitchUserID: "10002",
						TapeID:       4,
					},
				},
			},
			http.StatusOK,
			`{"tapeIds":[1,3]}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				q: tt.q,
			}
			handler := auth.RequireAccess(
				authmock.NewClient().AllowTwitchUserAccessToken("mock-token", auth.RoleViewer, auth.UserDetails{
					Id:          "54321",
					Login:       "jerry",
					DisplayName: "Jerry",
				}), auth.RoleViewer,
				http.HandlerFunc(s.handleGetFavorites),
			)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("authorization", "mock-token")
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)

			// Verify expected body and status code
			b, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			body := strings.TrimSuffix(string(b), "\n")
			assert.Equal(t, tt.wantStatus, res.Code)
			assert.Equal(t, tt.wantBody, body)
		})
	}
}

func Test_Server_handlePatchFavorites(t *testing.T) {
	tests := []struct {
		name                string
		q                   *mockQueries
		requestBody         string
		wantStatus          int
		wantBody            string
		wantFavoriteTapeIds []int32
	}{
		{
			"new tape can be registered as favorite",
			&mockQueries{
				validTapeIds: []int32{1, 2, 3, 4},
				favorites: []queries.RegisterFavoriteTapeParams{
					{
						TwitchUserID: "54321",
						TapeID:       1,
					},
				},
			},
			`{"tapeId":2,"isFavorite":true}`,
			http.StatusNoContent,
			"",
			[]int32{1, 2},
		},
		{
			"existing tape can be unregistered as favorite",
			&mockQueries{
				validTapeIds: []int32{1, 2, 3, 4},
				favorites: []queries.RegisterFavoriteTapeParams{
					{
						TwitchUserID: "54321",
						TapeID:       1,
					},
					{
						TwitchUserID: "54321",
						TapeID:       2,
					},
				},
			},
			`{"tapeId":2,"isFavorite":false}`,
			http.StatusNoContent,
			"",
			[]int32{1},
		},
		{
			"favoriting an already-favorited tape is a no-op",
			&mockQueries{
				validTapeIds: []int32{1, 2, 3, 4},
				favorites: []queries.RegisterFavoriteTapeParams{
					{
						TwitchUserID: "54321",
						TapeID:       1,
					},
				},
			},
			`{"tapeId":1,"isFavorite":true}`,
			http.StatusNoContent,
			"",
			[]int32{1},
		},
		{
			"unfavoriting an not-yet-favorited tape is a no-op",
			&mockQueries{
				validTapeIds: []int32{1, 2, 3, 4},
				favorites: []queries.RegisterFavoriteTapeParams{
					{
						TwitchUserID: "54321",
						TapeID:       1,
					},
				},
			},
			`{"tapeId":2,"isFavorite":false}`,
			http.StatusNoContent,
			"",
			[]int32{1},
		},
		{
			"attempting to register a nonexistent tape as a favorite is a 400 error",
			&mockQueries{
				validTapeIds: []int32{1, 2, 3, 4},
			},
			`{"tapeId":500,"isFavorite":true}`,
			http.StatusBadRequest,
			"no such tape",
			nil,
		},
		{
			"attempting to unregister a nonexistent tape as a favorite is a 400 error",
			&mockQueries{
				validTapeIds: []int32{1, 2, 3, 4},
			},
			`{"tapeId":500,"isFavorite":false}`,
			http.StatusBadRequest,
			"no such tape",
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				q: tt.q,
			}
			handler := auth.RequireAccess(
				authmock.NewClient().AllowTwitchUserAccessToken("mock-token", auth.RoleViewer, auth.UserDetails{
					Id:          "54321",
					Login:       "jerry",
					DisplayName: "Jerry",
				}), auth.RoleViewer,
				http.HandlerFunc(s.handlePatchFavorites),
			)
			req := httptest.NewRequest(http.MethodPatch, "/", strings.NewReader(tt.requestBody))
			req.Header.Set("authorization", "mock-token")
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, req)

			// Verify expected body and status code
			b, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			body := strings.TrimSuffix(string(b), "\n")
			assert.Equal(t, tt.wantStatus, res.Code)
			assert.Equal(t, tt.wantBody, body)

			// Verify expected db state changes
			favoriteTapeIds := make([]int32, 0)
			for _, favorite := range tt.q.favorites {
				if favorite.TwitchUserID == "54321" {
					favoriteTapeIds = append(favoriteTapeIds, favorite.TapeID)
				}
			}
			assert.ElementsMatch(t, favoriteTapeIds, tt.wantFavoriteTapeIds)
		})
	}
}

type mockQueries struct {
	validTapeIds []int32
	favorites    []queries.RegisterFavoriteTapeParams
}

func (m *mockQueries) RegisterFavoriteTape(ctx context.Context, arg queries.RegisterFavoriteTapeParams) error {
	if !m.isValidTapeId(arg.TapeID) {
		return &pq.Error{
			Code:    pq.ErrorCode("23503"),
			Message: "oh no, it's a foreign key violation",
		}
	}
	for _, favorite := range m.favorites {
		if favorite.TwitchUserID == arg.TwitchUserID && favorite.TapeID == arg.TapeID {
			return nil
		}
	}
	m.favorites = append(m.favorites, arg)
	return nil
}

func (m *mockQueries) UnregisterFavoriteTape(ctx context.Context, arg queries.UnregisterFavoriteTapeParams) error {
	if !m.isValidTapeId(arg.TapeID) {
		return &pq.Error{
			Code:    pq.ErrorCode("23503"),
			Message: "oh no, it's a foreign key violation",
		}
	}
	removeAtIndex := -1
	for i := range m.favorites {
		if m.favorites[i].TwitchUserID == arg.TwitchUserID && m.favorites[i].TapeID == arg.TapeID {
			removeAtIndex = i
			break
		}
	}
	if removeAtIndex >= 0 {
		m.favorites = append(m.favorites[:removeAtIndex], m.favorites[removeAtIndex+1:]...)
	}
	return nil
}

func (m *mockQueries) GetFavoriteTapes(ctx context.Context, twitchUserID string) ([]int32, error) {
	tapeIds := make([]int32, 0)
	for _, favorite := range m.favorites {
		if favorite.TwitchUserID == twitchUserID {
			tapeIds = append(tapeIds, favorite.TapeID)
		}
	}
	sort.Slice(tapeIds, func(i, j int) bool { return tapeIds[i] < tapeIds[j] })
	return tapeIds, nil
}

func (m *mockQueries) isValidTapeId(tapeId int32) bool {
	for _, validTapeId := range m.validTapeIds {
		if validTapeId == tapeId {
			return true
		}
	}
	return false
}
