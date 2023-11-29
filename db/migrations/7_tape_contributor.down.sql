begin;

alter table tapes.tape
    drop column contributor_id;

commit;

