-- name: GetTapes :many
select
    tape.id,
    tape.title,
    tape.year,
    tape.runtime,
    jsonb_agg(jsonb_build_object(
        'index', image.index,
        'color', image.color,
        'width', image.width,
        'height', image.height,
        'rotated', image.rotated
    ) order by image.index) as images,
    array(
        select tag_name
        from tapes.tape_to_tag
        where tape_to_tag.tape_id = tape.id
        order by tag_name
    )::text[] as tags
from tapes.tape
join tapes.image on image.tape_id = tape.id
group by tape.id
order by tape.id;
