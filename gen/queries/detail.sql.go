// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.20.0
// source: detail.sql

package queries

import (
	"context"
	"database/sql"
	"encoding/json"
)

const getTapes = `-- name: GetTapes :many
select
    tape.id,
    tape.title,
    tape.year,
    tape.runtime,
    jsonb_agg(jsonb_build_object(
        'index', image.index,
        'color', image.color,
        'width', image.width,
        'height', image.height,
        'rotated', image.rotated
    ) order by image.index) as images,
    jsonb_agg(tape_to_tag.tag_name order by tape_to_tag.tag_name) as tags
from tapes.tape
join tapes.image on image.tape_id = tape.id
join tapes.tape_to_tag on tape_to_tag.tape_id = tape.id
group by tape.id
order by tape.id
`

type GetTapesRow struct {
	ID      int32
	Title   string
	Year    sql.NullInt32
	Runtime sql.NullInt32
	Images  json.RawMessage
	Tags    json.RawMessage
}

func (q *Queries) GetTapes(ctx context.Context) ([]GetTapesRow, error) {
	rows, err := q.db.QueryContext(ctx, getTapes)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []GetTapesRow
	for rows.Next() {
		var i GetTapesRow
		if err := rows.Scan(
			&i.ID,
			&i.Title,
			&i.Year,
			&i.Runtime,
			&i.Images,
			&i.Tags,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}
