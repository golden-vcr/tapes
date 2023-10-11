package queries_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/golden-vcr/server-common/querytest"
	"github.com/golden-vcr/tapes/gen/queries"
	"github.com/golden-vcr/tapes/internal/db"
	"github.com/stretchr/testify/assert"
)

func Test_GetTapes(t *testing.T) {
	tx := querytest.PrepareTx(t)
	q := queries.New(tx)

	querytest.AssertCount(t, tx, 0, "SELECT COUNT(*) FROM tapes.tape")
	querytest.AssertCount(t, tx, 0, "SELECT COUNT(*) FROM tapes.tape_to_tag")
	querytest.AssertCount(t, tx, 0, "SELECT COUNT(*) FROM tapes.image")

	err := q.SyncTape(context.Background(), queries.SyncTapeParams{
		ID:    1,
		Title: "Tape one",
		Year:  sql.NullInt32{Valid: true, Int32: 1994},
	})
	assert.NoError(t, err)

	err = q.SyncTapeTags(context.Background(), queries.SyncTapeTagsParams{
		TapeID:   1,
		TagNames: []string{"fitness", "instructional"},
	})
	assert.NoError(t, err)

	err = q.SyncImage(context.Background(), queries.SyncImageParams{
		TapeID:  1,
		Index:   0,
		Width:   500,
		Height:  1000,
		Color:   "#ff0000",
		Rotated: false,
	})
	assert.NoError(t, err)
	err = q.SyncImage(context.Background(), queries.SyncImageParams{
		TapeID:  1,
		Index:   1,
		Width:   501,
		Height:  1001,
		Color:   "#00ff00",
		Rotated: true,
	})
	assert.NoError(t, err)

	querytest.AssertCount(t, tx, 1, "SELECT COUNT(*) FROM tapes.tape")
	querytest.AssertCount(t, tx, 2, "SELECT COUNT(*) FROM tapes.tape_to_tag")
	querytest.AssertCount(t, tx, 2, "SELECT COUNT(*) FROM tapes.image")

	rows, err := q.GetTapes(context.Background())
	assert.NoError(t, err)
	assert.Len(t, rows, 1)

	row := rows[0]
	assert.Equal(t, int32(1), row.ID)
	assert.Equal(t, "Tape one", row.Title)
	assert.Equal(t, sql.NullInt32{Valid: true, Int32: 1994}, row.Year)
	assert.Equal(t, sql.NullInt32{}, row.Runtime)
	images, err := db.ParseTapeImageArray(row.Images)
	assert.NoError(t, err)
	assert.Equal(t, []db.TapeImage{
		{
			Index:   0,
			Color:   "#ff0000",
			Width:   500,
			Height:  1000,
			Rotated: false,
		},
		{
			Index:   1,
			Color:   "#00ff00",
			Width:   501,
			Height:  1001,
			Rotated: true,
		},
	}, images)
	assert.Equal(t, []string{"fitness", "instructional"}, row.Tags)
}
