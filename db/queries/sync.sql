-- name: SyncTape :exec
insert into tapes.tape (
    id,
    created_at,
    title,
    year,
    runtime
) values (
    @id,
    now(),
    @title,
    sqlc.narg('year'),
    sqlc.narg('runtime')
)
on conflict (id) do update set
    title = excluded.title,
    year = excluded.year,
    runtime = excluded.runtime;

-- name: SyncImage :exec
insert into tapes.image (
    tape_id,
    index,
    color,
    width,
    height,
    rotated
) values (
    @tape_id,
    @index,
    @color,
    @width,
    @height,
    @rotated
)
on conflict (tape_id, index) do update set
    color = excluded.color,
    width = excluded.width,
    height = excluded.height,
    rotated = excluded.rotated;
