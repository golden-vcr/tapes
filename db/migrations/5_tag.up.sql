begin;

create table tapes.tape_to_tag (
    tape_id  integer not null,
    tag_name text not null
);

alter table tapes.tape_to_tag
    add constraint tape_to_tag_tape_id_fk
    foreign key (tape_id) references tapes.tape (id);

alter table tapes.tape_to_tag
    add constraint tape_to_tag_unique
    unique (tape_id, tag_name);

comment on table tapes.tape_to_tag is
    'Association of a specific tag name with a given tape.';
comment on column tapes.tape_to_tag.tape_id is
    'Foreign-key reference to the tape which has this tag.';
comment on column tapes.tape_to_tag.tag_name is
    'Canonical name of the tag. Tags are identified solely by a lowercase string, e.g. '
    '"instructional", "arts+crafts", "christmas".';

commit;
