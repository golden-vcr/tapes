package queries_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/golden-vcr/server-common/querytest"
	"github.com/golden-vcr/tapes/gen/queries"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func Test_CreateSync(t *testing.T) {
	tx := querytest.PrepareTx(t)
	q := queries.New(tx)

	querytest.AssertCount(t, tx, 0, "SELECT COUNT(*) FROM tapes.sync")

	syncUuid := uuid.MustParse("d0311839-3666-46f3-82df-f97d28716a34")
	err := q.CreateSync(context.Background(), syncUuid)
	assert.NoError(t, err)

	querytest.AssertCount(t, tx, 1, `
		SELECT COUNT(*) FROM tapes.sync
			WHERE uuid = 'd0311839-3666-46f3-82df-f97d28716a34'
			AND started_at = now()
			AND finished_at IS NULL
			AND error IS NULL
			AND num_tapes IS NULL
			AND warnings IS NULL
	`)
}

func Test_RecordFailedSync(t *testing.T) {
	tx := querytest.PrepareTx(t)
	q := queries.New(tx)

	querytest.AssertCount(t, tx, 0, "SELECT COUNT(*) FROM tapes.sync")

	syncUuid := uuid.MustParse("2944e900-71d5-48a0-8638-9ec263571149")
	err := q.CreateSync(context.Background(), syncUuid)
	assert.NoError(t, err)

	q.RecordFailedSync(context.Background(), queries.RecordFailedSyncParams{
		Uuid:  syncUuid,
		Error: "something horrible happened",
	})

	querytest.AssertCount(t, tx, 1, `
		SELECT COUNT(*) FROM tapes.sync
			WHERE uuid = '2944e900-71d5-48a0-8638-9ec263571149'
			AND started_at = now()
			AND finished_at = now()
			AND error = 'something horrible happened'
			AND num_tapes IS NULL
			AND warnings IS NULL
	`)
}

func Test_RecordSuccessfulSync(t *testing.T) {
	tx := querytest.PrepareTx(t)
	q := queries.New(tx)

	querytest.AssertCount(t, tx, 0, "SELECT COUNT(*) FROM tapes.sync")

	syncUuid := uuid.MustParse("12cc8adb-1bfe-43c9-a757-11e76531e077")
	err := q.CreateSync(context.Background(), syncUuid)
	assert.NoError(t, err)

	q.RecordSuccessfulSync(context.Background(), queries.RecordSuccessfulSyncParams{
		Uuid:     syncUuid,
		NumTapes: 42,
		Warnings: "tape 4 has mold on it\nsomething smells bad",
	})

	querytest.AssertCount(t, tx, 1, `
		SELECT COUNT(*) FROM tapes.sync
			WHERE uuid = '12cc8adb-1bfe-43c9-a757-11e76531e077'
			AND started_at = now()
			AND finished_at = now()
			AND error IS NULL
			AND num_tapes = 42
			AND warnings = 'tape 4 has mold on it
something smells bad'
	`)
}

func Test_SyncTape(t *testing.T) {
	tx := querytest.PrepareTx(t)
	q := queries.New(tx)

	querytest.AssertCount(t, tx, 0, "SELECT COUNT(*) FROM tapes.tape")

	err := q.SyncTape(context.Background(), queries.SyncTapeParams{
		ID:    101,
		Title: "My coool tape",
		Year:  sql.NullInt32{Valid: true, Int32: 1997},
	})
	assert.NoError(t, err)
	querytest.AssertCount(t, tx, 1, `
		SELECT COUNT(*) FROM tapes.tape
			WHERE id = 101
			AND title = 'My coool tape'
			AND year = 1997
			AND runtime IS NULL
	`)

	err = q.SyncTape(context.Background(), queries.SyncTapeParams{
		ID:      101,
		Title:   "My cool tape",
		Runtime: sql.NullInt32{Valid: true, Int32: 120},
	})
	assert.NoError(t, err)
	querytest.AssertCount(t, tx, 1, `
		SELECT COUNT(*) FROM tapes.tape
			WHERE id = 101
			AND title = 'My cool tape'
			AND year IS NULL
			AND runtime = 120
	`)

	querytest.AssertCount(t, tx, 1, "SELECT COUNT(*) FROM tapes.tape")
}

func Test_SyncTapeTags(t *testing.T) {
	tx := querytest.PrepareTx(t)
	q := queries.New(tx)

	querytest.AssertCount(t, tx, 0, "SELECT COUNT(*) FROM tapes.tape_to_tag")

	err := q.SyncTape(context.Background(), queries.SyncTapeParams{
		ID:    15,
		Title: "Test tape",
	})
	assert.NoError(t, err)

	err = q.SyncTapeTags(context.Background(), queries.SyncTapeTagsParams{
		TapeID:   15,
		TagNames: []string{"instructional", "arts+crafts"},
	})
	assert.NoError(t, err)
	querytest.AssertCount(t, tx, 2, `
		SELECT COUNT(*) FROM tapes.tape_to_tag
			WHERE tape_id = 15
			AND tag_name IN ('instructional', 'arts+crafts')
	`)

	err = q.SyncTapeTags(context.Background(), queries.SyncTapeTagsParams{
		TapeID:   15,
		TagNames: []string{"instructional", "fitness"},
	})
	assert.NoError(t, err)
	querytest.AssertCount(t, tx, 2, `
		SELECT COUNT(*) FROM tapes.tape_to_tag
			WHERE tape_id = 15
			AND tag_name IN ('instructional', 'fitness')
	`)

	querytest.AssertCount(t, tx, 2, "SELECT COUNT(*) FROM tapes.tape_to_tag")
}

func Test_SyncImage(t *testing.T) {
	tx := querytest.PrepareTx(t)
	q := queries.New(tx)

	querytest.AssertCount(t, tx, 0, "SELECT COUNT(*) FROM tapes.image")

	err := q.SyncTape(context.Background(), queries.SyncTapeParams{
		ID:    99,
		Title: "Tape 99",
	})
	assert.NoError(t, err)

	err = q.SyncImage(context.Background(), queries.SyncImageParams{
		TapeID:  99,
		Index:   0,
		Color:   "#ffcc00",
		Width:   700,
		Height:  1500,
		Rotated: false,
	})
	assert.NoError(t, err)
	querytest.AssertCount(t, tx, 1, `
		SELECT COUNT(*) FROM tapes.image
			WHERE tape_id = 99
			AND index = 0
			AND color = '#ffcc00'
			AND width = 700
			AND height = 1500
			AND NOT rotated
	`)

	err = q.SyncImage(context.Background(), queries.SyncImageParams{
		TapeID:  99,
		Index:   0,
		Color:   "#ffcc00",
		Width:   780,
		Height:  1500,
		Rotated: false,
	})
	assert.NoError(t, err)
	querytest.AssertCount(t, tx, 1, `
		SELECT COUNT(*) FROM tapes.image
			WHERE tape_id = 99
			AND index = 0
			AND color = '#ffcc00'
			AND width = 780
			AND height = 1500
			AND NOT rotated
	`)

	err = q.SyncImage(context.Background(), queries.SyncImageParams{
		TapeID:  99,
		Index:   1,
		Color:   "#ffee99",
		Width:   701,
		Height:  1499,
		Rotated: true,
	})
	assert.NoError(t, err)
	querytest.AssertCount(t, tx, 1, `
		SELECT COUNT(*) FROM tapes.image
			WHERE tape_id = 99
			AND index = 1
			AND color = '#ffee99'
			AND width = 701
			AND height = 1499
			AND rotated
	`)

	querytest.AssertCount(t, tx, 2, "SELECT COUNT(*) FROM tapes.image")
}
