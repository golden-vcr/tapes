package queries_test

import (
	"context"
	"testing"

	"github.com/golden-vcr/server-common/querytest"
	"github.com/golden-vcr/tapes/gen/queries"
	"github.com/stretchr/testify/assert"
)

func Test_RegisterFavoriteTape(t *testing.T) {
	tx := querytest.PrepareTx(t)
	q := queries.New(tx)

	_, err := tx.Exec("INSERT INTO tapes.tape (id, created_at, title) VALUES (42, now(), 'Test tape')")
	assert.NoError(t, err)

	querytest.AssertCount(t, tx, 0, "SELECT COUNT(*) FROM tapes.favorite")

	// We should be able to register tape 42 (which we just created) as a favorite
	err = q.RegisterFavoriteTape(context.Background(), queries.RegisterFavoriteTapeParams{
		TwitchUserID: "1234",
		TapeID:       42,
	})
	assert.NoError(t, err)

	// Registering a favorite should be idempotent
	err = q.RegisterFavoriteTape(context.Background(), queries.RegisterFavoriteTapeParams{
		TwitchUserID: "1234",
		TapeID:       42,
	})
	assert.NoError(t, err)

	// We should now have tape 42 registered as a favorite for user 1234
	querytest.AssertCount(t, tx, 1, `
		SELECT COUNT(*) FROM tapes.favorite
			WHERE twitch_user_id = '1234'
			AND tape_id = 42
	`)

	// If tape_id does not reference a valid tape, we should get an error
	err = q.RegisterFavoriteTape(context.Background(), queries.RegisterFavoriteTapeParams{
		TwitchUserID: "1234",
		TapeID:       100,
	})
	assert.Error(t, err)
}

func Test_UnregisterFavoriteTape(t *testing.T) {
	tx := querytest.PrepareTx(t)
	q := queries.New(tx)

	_, err := tx.Exec("INSERT INTO tapes.tape (id, created_at, title) VALUES (42, now(), 'Test tape')")
	assert.NoError(t, err)
	_, err = tx.Exec("INSERT INTO tapes.favorite (twitch_user_id, tape_id) VALUES ('1234', 42)")
	assert.NoError(t, err)

	querytest.AssertCount(t, tx, 1, `
		SELECT COUNT(*) FROM tapes.favorite
			WHERE twitch_user_id = '1234'
			AND tape_id = 42
	`)

	// We should be able to unregister tape 42 (which we just created and registered as
	// a favorite)
	err = q.UnregisterFavoriteTape(context.Background(), queries.UnregisterFavoriteTapeParams{
		TwitchUserID: "1234",
		TapeID:       42,
	})
	assert.NoError(t, err)

	// Unregistering a favorite should be idempotent
	err = q.UnregisterFavoriteTape(context.Background(), queries.UnregisterFavoriteTapeParams{
		TwitchUserID: "1234",
		TapeID:       42,
	})
	assert.NoError(t, err)

	querytest.AssertCount(t, tx, 0, `
		SELECT COUNT(*) FROM tapes.favorite
			WHERE twitch_user_id = '1234'
			AND tape_id = 42
	`)

	// Unregistering a favorite that doesn't exist should be a no-op
	err = q.UnregisterFavoriteTape(context.Background(), queries.UnregisterFavoriteTapeParams{
		TwitchUserID: "5678",
		TapeID:       42,
	})
	assert.NoError(t, err)

	err = q.UnregisterFavoriteTape(context.Background(), queries.UnregisterFavoriteTapeParams{
		TwitchUserID: "1234",
		TapeID:       99,
	})
	assert.NoError(t, err)
}

func Test_GetFavoriteTapes(t *testing.T) {
	tx := querytest.PrepareTx(t)
	q := queries.New(tx)

	_, err := tx.Exec(`
		INSERT INTO tapes.tape (id, created_at, title) VALUES
			(1, now(), 'Tape 1'),
			(2, now(), 'Tape 2'),
			(3, now(), 'Tape 3')
	`)
	assert.NoError(t, err)

	_, err = tx.Exec("INSERT INTO tapes.favorite (twitch_user_id, tape_id) VALUES ('1234', 1), ('1234', 3)")
	assert.NoError(t, err)

	tapeIds, err := q.GetFavoriteTapes(context.Background(), "1234")
	assert.NoError(t, err)
	assert.Equal(t, []int32{1, 3}, tapeIds)

	tapeIds, err = q.GetFavoriteTapes(context.Background(), "5678")
	assert.NoError(t, err)
	assert.Len(t, tapeIds, 0)
}
