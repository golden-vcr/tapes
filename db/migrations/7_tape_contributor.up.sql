begin;

alter table tapes.tape
    add column contributor_id text;

comment on column tapes.tape.contributor_id is
    'Twitch User ID of the viewer who contributed this tape to the library, if any.';

commit;
