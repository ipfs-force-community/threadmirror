-- BotCookie queries

-- name: GetBotCookieByID :one
SELECT * FROM bot_cookie WHERE id = @id AND deleted_at IS NULL;

-- name: GetBotCookieByEmailAndUsername :one
SELECT * FROM bot_cookie 
WHERE email = @email AND username = @username AND deleted_at IS NULL;

-- name: CreateBotCookie :one
INSERT INTO bot_cookie (
    email, username, cookies_data
) VALUES (
    @email, @username, @cookies_data
) RETURNING *;

-- name: UpdateBotCookie :exec
UPDATE bot_cookie SET
    email = @email,
    username = @username,
    cookies_data = @cookies_data,
    updated_at = NOW()
WHERE id = @id AND deleted_at IS NULL;

-- name: SoftDeleteBotCookie :exec
UPDATE bot_cookie SET
    deleted_at = NOW()
WHERE id = @id AND deleted_at IS NULL;

-- name: ListBotCookies :many
SELECT * FROM bot_cookie 
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT @limit_ OFFSET @offset_;

-- name: CountBotCookies :one
SELECT COUNT(*) FROM bot_cookie WHERE deleted_at IS NULL;
