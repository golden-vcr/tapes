package catalog

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golden-vcr/tapes/gen/queries"
	"github.com/golden-vcr/tapes/internal/db"
	"github.com/stretchr/testify/assert"
)

func Test_Server_handleGetListing(t *testing.T) {
	imageHostUrl := "https://my-images.biz"
	tests := []struct {
		name       string
		q          *mockQueries
		wantStatus int
		wantBody   string
	}{
		{
			"normal usage",
			&mockQueries{
				rows: []queries.GetTapesRow{
					{
						ID:      1,
						Title:   "Tape one",
						Year:    sql.NullInt32{Valid: true, Int32: 1991},
						Runtime: sql.NullInt32{Valid: true, Int32: 120},
						Images: encodeTapeImages(t, []db.TapeImage{
							{
								Index:   0,
								Color:   "#ffccee",
								Width:   440,
								Height:  1301,
								Rotated: false,
							},
							{
								Index:   1,
								Color:   "#eebbee",
								Width:   441,
								Height:  1300,
								Rotated: true,
							},
						}),
					},
				},
			},
			http.StatusOK,
			`{"imageHost":"https://my-images.biz","items":[{"id":1,"title":"Tape one","year":1991,"runtime":120,"thumbnail":"0001_thumb.jpg","images":[{"filename":"0001_a.jpg","width":440,"height":1301,"color":"#ffccee","rotated":false},{"filename":"0001_b.jpg","width":441,"height":1300,"color":"#eebbee","rotated":true}],"tags":["fitness","instructional"]}]}`,
		},
		{
			"null year and runtime are represented as 0",
			&mockQueries{
				rows: []queries.GetTapesRow{
					{
						ID:      1,
						Title:   "Tape one",
						Year:    sql.NullInt32{},
						Runtime: sql.NullInt32{},
						Images: encodeTapeImages(t, []db.TapeImage{
							{
								Index:   0,
								Color:   "#ffccee",
								Width:   440,
								Height:  1301,
								Rotated: false,
							},
						}),
					},
				},
			},
			http.StatusOK,
			`{"imageHost":"https://my-images.biz","items":[{"id":1,"title":"Tape one","year":0,"runtime":0,"thumbnail":"0001_thumb.jpg","images":[{"filename":"0001_a.jpg","width":440,"height":1301,"color":"#ffccee","rotated":false}],"tags":["fitness","instructional"]}]}`,
		},
		{
			"unexpected JSON format for image data is a 500 error",
			&mockQueries{
				rows: []queries.GetTapesRow{
					{
						ID:      1,
						Title:   "Tape one",
						Year:    sql.NullInt32{},
						Runtime: sql.NullInt32{},
						Images:  []byte(`[{"index":"not-a-valid-int","color":"#ffccee","width": 440,"height": 1301,"rotated":false}]`),
					},
				},
			},
			http.StatusInternalServerError,
			"failed to parse TapeImage array from JSON data: json: cannot unmarshal string into Go struct field TapeImage.index of type int32",
		},
		{
			"database error is a 500 error",
			&mockQueries{
				err: fmt.Errorf("mock error"),
			},
			http.StatusInternalServerError,
			"mock error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				q:            tt.q,
				imageHostUrl: imageHostUrl,
			}
			req := httptest.NewRequest(http.MethodPost, "/", nil)
			res := httptest.NewRecorder()
			s.handleGetListing(res, req)

			b, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			body := strings.TrimSuffix(string(b), "\n")
			assert.Equal(t, tt.wantStatus, res.Code)
			assert.Equal(t, tt.wantBody, body)
		})
	}
}

type mockQueries struct {
	err  error
	rows []queries.GetTapesRow
}

func (m *mockQueries) GetTapes(ctx context.Context) ([]queries.GetTapesRow, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.rows, nil
}

var _ Queries = (*mockQueries)(nil)

func encodeTapeImages(t *testing.T, images []db.TapeImage) json.RawMessage {
	data, err := json.Marshal(images)
	assert.NoError(t, err)
	return data
}
