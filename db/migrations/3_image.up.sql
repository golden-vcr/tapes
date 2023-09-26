begin;

create table tapes.image (
    tape_id integer not null,
    index   integer not null,

    color   text not null,
    width   integer not null,
    height  integer not null,
    rotated boolean not null
);

alter table tapes.image
    add constraint image_tape_id_fk
    foreign key (tape_id) references tapes.tape (id);

alter table tapes.image
    add constraint image_tape_id_index_unique
    unique (tape_id, index);

alter table tapes.image
    add constraint image_color_must_be_hex
    check (color ~* '^#[a-f0-9]{3}[a-f0-9]{3}$');

comment on table tapes.image is
    'Metadata for a single image scanned from a specific tape.';
comment on column tapes.image.index is
    'Zero-indexed value indicating this image''s position in a series of scans made '
    'for the same tape.';
comment on column tapes.image.color is
    'The dominant color in the image, hex-formatted (with hash prefix).';
comment on column tapes.image.width is
    'Width of the image file, in pixels.';
comment on column tapes.image.height is
    'Height of the image file, in pixels.';
comment on column tapes.image.rotated is
    'Whether the image was rotated 90 degrees CCW in order to have a vertical aspect '
    'ratio, in which case it may be displayed with a 90-degree CW rotation applied in '
    'order for any text in the image to be legible.';

commit;
