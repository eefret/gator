-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (feed_id, user_id)
    VALUES ($1, $2)
    RETURNING *
)
SELECT
    inserted_feed_follow.*,
    feeds.name AS feed_name,
    users.name AS user_name
FROM inserted_feed_follow
INNER JOIN feeds ON feeds.id = inserted_feed_follow.feed_id
INNER JOIN users ON users.id = inserted_feed_follow.user_id;

-- name: GetFeedFollowsForUser :many
SELECT ff.id, ff.user_id, ff.feed_id, ff.created_at, ff.updated_at,
	       f.name AS feed_name,
	       u.name AS user_name
	FROM feed_follows ff
	INNER JOIN feeds f ON f.id = ff.feed_id
	INNER JOIN users u ON u.id = ff.user_id
	WHERE ff.user_id = $1;

-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows ff
USING feeds f
WHERE ff.user_id = $1
  AND f.url = $2
  AND ff.feed_id = f.id;