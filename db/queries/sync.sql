-- name: CreateSync :exec
insert into tapes.sync (
    uuid,
    started_at
) values (
    @uuid,
    now()
);

-- name: RecordFailedSync :exec
update tapes.sync set
    finished_at = now(),
    error = @error::text
where
    sync.uuid = @uuid
    and finished_at is null;

-- name: RecordSuccessfulSync :exec
update tapes.sync set
    finished_at = now(),
    num_tapes = @num_tapes::integer,
    warnings = @warnings::text
where
    sync.uuid = @uuid
    and finished_at is null;

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
