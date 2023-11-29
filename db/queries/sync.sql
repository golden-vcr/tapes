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
    runtime,
    contributor_id
) values (
    @id,
    now(),
    @title,
    sqlc.narg('year'),
    sqlc.narg('runtime'),
    sqlc.narg('contributor_id')
)
on conflict (id) do update set
    title = excluded.title,
    year = excluded.year,
    runtime = excluded.runtime,
    contributor_id = excluded.contributor_id;

-- name: SyncTapeTags :exec
with deleted as (
    delete from tapes.tape_to_tag
        where tape_id = @tape_id
        and not (tag_name = any(@tag_names::text[]))
)
insert into tapes.tape_to_tag (tape_id, tag_name)
select
    @tape_id as tape_id,
    tag_name from unnest(@tag_names::text[]) as tag_name
on conflict do nothing;

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
