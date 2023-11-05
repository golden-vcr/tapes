-- name: RegisterFavoriteTape :exec
insert into tapes.favorite (
    twitch_user_id,
    tape_id
) values (
    @twitch_user_id,
    @tape_id
)
on conflict (twitch_user_id, tape_id) do nothing;

-- name: UnregisterFavoriteTape :exec
delete from tapes.favorite
    where favorite.twitch_user_id = @twitch_user_id
    and favorite.tape_id = @tape_id;

-- name: GetFavoriteTapes :many
select
    favorite.tape_id
from tapes.favorite
where favorite.twitch_user_id = @twitch_user_id
order by favorite.tape_id;
