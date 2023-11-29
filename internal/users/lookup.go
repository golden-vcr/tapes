package users

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nicklaw5/helix/v2"
)

type Lookup interface {
	Resolve(ctx context.Context, ids []string) error
	GetDisplayName(id string) string
}

func NewLookup(twitchClientId string, twitchClientSecret string) (Lookup, error) {
	c, err := helix.NewClient(&helix.Options{
		ClientID:     twitchClientId,
		ClientSecret: twitchClientSecret,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Twitch API client: %w", err)
	}

	res, err := c.RequestAppAccessToken(nil)
	if err == nil && res.StatusCode != http.StatusOK {
		err = fmt.Errorf("got status %d: %s", res.StatusCode, res.ErrorMessage)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get app access token from Twitch API: %w", err)
	}

	c.SetAppAccessToken(res.Data.AccessToken)
	return &twitchUserLookup{
		c:                c,
		displayNamesById: make(map[string]string),
	}, nil
}

// twitchUserLookup is an implementation of users.Lookup that uses the Twitch API in
// order to resolve user's human-facing display names given their numeric Twitch User ID
// values.
type twitchUserLookup struct {
	c                *helix.Client
	displayNamesById map[string]string
}

func (l *twitchUserLookup) Resolve(ctx context.Context, ids []string) error {
	// Find the subset of requested user IDs that we haven't already cached names for
	idsToResolve := make([]string, 0, len(ids))
	for _, id := range ids {
		if _, ok := l.displayNamesById[id]; !ok {
			idsToResolve = append(idsToResolve, id)
			// Twitch API limits us to 100 user IDs per GetUsers call, so just cap our
			// input at 100 and we'll resolve the remainder on subsequent calls
			if len(idsToResolve) == 100 {
				break
			}
		}
	}

	// If we have nothing to look up, early-out
	if len(idsToResolve) == 0 {
		return nil
	}

	// Make a call to the Twitch API to resolve display names for all of our users
	res, err := l.c.GetUsers(&helix.UsersParams{
		IDs: idsToResolve,
	})
	if err != nil {
		return err
	}

	for _, user := range res.Data.Users {
		l.displayNamesById[user.ID] = user.DisplayName
	}
	return nil
}

func (l *twitchUserLookup) GetDisplayName(id string) string {
	displayName, ok := l.displayNamesById[id]
	if !ok {
		return fmt.Sprintf("User %s", id)
	}
	return displayName
}
