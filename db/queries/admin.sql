-- name: ApplySeries :execresult
update tapes.tape set series_name = @series_name
where tape.id in (sqlc.arg('tape_ids')::integer[]);
