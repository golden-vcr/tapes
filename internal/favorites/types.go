package favorites

import (
	"context"

	"github.com/golden-vcr/tapes/gen/queries"
)

type Queries interface {
	RegisterFavoriteTape(ctx context.Context, arg queries.RegisterFavoriteTapeParams) error
	UnregisterFavoriteTape(ctx context.Context, arg queries.UnregisterFavoriteTapeParams) error
	GetFavoriteTapes(ctx context.Context, twitchUserID string) ([]int32, error)
}

type FavoriteTapeSet struct {
	TapeIds []int `json:"tapeIds"`
}

type FavoriteTapeChange struct {
	TapeId     int  `json:"tapeId"`
	IsFavorite bool `json:"isFavorite"`
}
