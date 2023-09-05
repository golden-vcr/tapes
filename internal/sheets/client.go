package sheets

import (
	"context"
	"fmt"
	"sort"
	"time"
)

const SheetName = "Tapes"

type Client struct {
	sheetsApiKey  string
	spreadsheetId string
	ttl           time.Duration

	rows           RowLookup
	expirationTime time.Time
}

func NewClient(ctx context.Context, sheetsApiKey string, spreadsheetId string, ttl time.Duration) (*Client, error) {
	client := &Client{
		sheetsApiKey:  sheetsApiKey,
		spreadsheetId: spreadsheetId,
		ttl:           ttl,

		rows:           nil,
		expirationTime: time.UnixMilli(0),
	}
	if err := client.updateRows(ctx); err != nil {
		return nil, err
	}
	return client, nil
}

func (c *Client) ListTapes(ctx context.Context) ([]Row, error) {
	rows, err := c.getRows(ctx)
	if err != nil {
		return nil, err
	}

	ids := make([]int, 0, len(rows))
	for k := range rows {
		ids = append(ids, k)
	}
	sort.Ints(ids)

	result := make([]Row, 0, len(ids))
	for _, id := range ids {
		result = append(result, rows[id])
	}
	return result, nil
}

func (c *Client) updateRows(ctx context.Context) error {
	result, err := getSheetValues(ctx, c.sheetsApiKey, c.spreadsheetId, SheetName)
	if err != nil {
		return err
	}
	rows, err := buildRowLookup(result)
	if err != nil {
		return err
	}
	c.rows = rows
	c.expirationTime = time.Now().Add(c.ttl)
	return nil
}

func (c *Client) getRows(ctx context.Context) (RowLookup, error) {
	var err error
	if time.Now().After(c.expirationTime) {
		err = c.updateRows(ctx)
		if err != nil && c.rows != nil {
			fmt.Printf("WARNING: Failed to read spreadsheet data; reusing stale data: %v", err)
			err = nil
		}
	}
	return c.rows, err
}
