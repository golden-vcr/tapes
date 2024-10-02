-- name: ApplySeries :execresult
update tapes.tape set series_name = @series_name
where tape.id = any(sqlc.arg('tape_ids')::integer[]);
