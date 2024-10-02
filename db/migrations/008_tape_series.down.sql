begin;

alter table tapes.tape
    drop column series_name;

commit;
