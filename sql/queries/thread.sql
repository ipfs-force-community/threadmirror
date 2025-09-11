-- Thread queries

-- name: GetThreadByID :one
SELECT * FROM thread WHERE id = @thread_id;

-- name: CreateThread :one
INSERT INTO thread (
    id, summary, cid, num_tweets, status, retry_count, version,
    author_id, author_name, author_screen_name, author_profile_image_url
) VALUES (
    @id, @summary, @cid, @num_tweets, @status, @retry_count, @version,
    @author_id, @author_name, @author_screen_name, @author_profile_image_url
) RETURNING *;

-- name: UpdateThreadComplete :exec
UPDATE thread SET
    summary = @summary,
    cid = @cid,
    num_tweets = @num_tweets,
    status = @status,
    retry_count = @retry_count,
    version = version + 1,
    author_id = @author_id,
    author_name = @author_name,
    author_screen_name = @author_screen_name,
    author_profile_image_url = @author_profile_image_url,
    updated_at = NOW()
WHERE id = @id AND version = @expected_version;

-- name: UpdateThreadStatus :exec
UPDATE thread SET
    status = @status,
    version = version + 1,
    updated_at = NOW()
WHERE id = @thread_id AND version = @current_version;

-- name: GetThreadsByIDs :many
SELECT * FROM thread WHERE id = ANY(@thread_ids::uuid[]);

-- name: GetStuckScrapingThreads :many
SELECT * FROM thread 
WHERE status = 'scraping' 
  AND updated_at < @cutoff_time 
  AND retry_count < @max_retries
FOR UPDATE;

-- name: IncrementThreadRetryCount :exec
UPDATE thread SET 
    retry_count = retry_count + 1,
    updated_at = NOW()
WHERE id = ANY(@thread_ids::uuid[]);

-- name: GetOldPendingThreads :many
SELECT * FROM thread 
WHERE status = 'pending' 
  AND created_at < @cutoff_time 
  AND retry_count < @max_retries
FOR UPDATE;

-- name: GetFailedThreadsForRetry :many
SELECT * FROM thread 
WHERE status = 'failed' 
  AND updated_at < @cutoff_time 
  AND retry_count < @max_retries
FOR UPDATE;
