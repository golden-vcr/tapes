begin;

alter table tapes.tape
    add column series_name text not null default '';

commit;
