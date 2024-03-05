begin;

create table tapes.tape (
    id         integer primary key,
    created_at timestamptz not null,

    title   text not null,
    year    integer,
    runtime integer
);

alter table tapes.tape
    add constraint id_must_be_positive
    check (id > 0);

comment on table tapes.tape is
    'Details of a single VHS tape in the Golden VCR library.';
comment on column tapes.tape.id is
    'Numeric ID with which the tape is identified in the inventory spreadsheet.';
comment on column tapes.tape.created_at is
    'Timestamp at which this tape was first synced to the database.';
comment on column tapes.tape.title is
    'Title entered for this tape in the spreadsheet.';
comment on column tapes.tape.year is
    'Release year noted for this tape in the spreadsheet; or NULL if unknown.';
comment on column tapes.tape.runtime is
    'Runtime (in minutes) noted for this tape in the spreadsheet; or NULL if unknown.';

commit;
