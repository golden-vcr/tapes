begin;

create table tapes.sync (
    uuid        uuid primary key,
    started_at  timestamptz not null,
    finished_at timestamptz,
    error       text,
    num_tapes   integer,
    warnings    text
);

comment on table tapes.sync is
    'Record of an attempt to sync tape and image data to the GVCR database.';
comment on column tapes.sync.uuid is
    'Unique identifier for this sync.';
comment on column tapes.sync.started_at is
    'Time at which the sync started.';
comment on column tapes.sync.finished_at is
    'Time at which the sync finished, unless still in progress or abandoned.';
comment on column tapes.sync.error is
    'Error message that ended the sync prematurely.';
comment on column tapes.sync.num_tapes is
    'Number of tapes successfully synced.';
comment on column tapes.sync.warnings is
    'Newline-delimited string containing all warning lines emitted during the sync.';

commit;
