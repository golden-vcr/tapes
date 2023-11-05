begin;

create table tapes.favorite (
    twitch_user_id text not null,
    tape_id        integer not null
);

comment on table tapes.favorite is
    'Records the fact that a specific user has marked a single tape as one of their '
    'favorite tapes.';
comment on column tapes.favorite.twitch_user_id is
    'ID of the user who marked this tape as a favorite.';
comment on column tapes.favorite.tape_id is
    'ID of the tape that the user has marked as a favorite.';

alter table tapes.favorite
    add constraint favorite_tape_id_fk
    foreign key (tape_id) references tapes.tape (id);

alter table tapes.favorite
    add constraint favorite_user_id_tape_id_unique
    unique (twitch_user_id, tape_id);

commit;
