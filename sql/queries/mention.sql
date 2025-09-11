-- Mention queries

-- name: GetMentionByID :one
SELECT m.*, t.* FROM mention m
JOIN thread t ON m.thread_id = t.id
WHERE m.id = @mention_id;

-- name: GetMentionByUserIDAndThreadID :one
SELECT m.*, t.* FROM mention m
JOIN thread t ON m.thread_id = t.id
WHERE m.user_id = @user_id AND m.thread_id = @thread_id;

-- name: CreateMention :one
INSERT INTO mention (
    id, user_id, thread_id, mention_create_at
) VALUES (
    @id, @user_id, @thread_id, @mention_create_at
) RETURNING *;

-- name: UpdateMention :exec
UPDATE mention SET
    user_id = @user_id,
    thread_id = @thread_id,
    mention_create_at = @mention_create_at,
    updated_at = NOW()
WHERE id = @id;

-- name: GetMentions :many
SELECT m.*, t.* FROM mention m
JOIN thread t ON m.thread_id = t.id
WHERE (@user_id::text IS NULL OR m.user_id = @user_id)
ORDER BY m.created_at DESC
LIMIT @limit_ OFFSET @offset_;

-- name: CountMentions :one
SELECT COUNT(*) FROM mention m
WHERE (@user_id::text IS NULL OR m.user_id = @user_id);

-- name: GetMentionsByUser :many
SELECT m.*, t.* FROM mention m
JOIN thread t ON m.thread_id = t.id
WHERE m.user_id = @user_id
ORDER BY m.created_at DESC
LIMIT @limit_ OFFSET @offset_;

-- name: CountMentionsByUser :one
SELECT COUNT(*) FROM mention m
WHERE m.user_id = @user_id;
