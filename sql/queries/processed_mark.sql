-- ProcessedMark queries

-- name: GetProcessedMark :one
SELECT * FROM processed_mark WHERE key = @key AND type = @type;

-- name: CreateProcessedMark :one
INSERT INTO processed_mark (key, type) 
VALUES (@key, @type) 
RETURNING *;

-- name: UpsertProcessedMark :one
INSERT INTO processed_mark (key, type) 
VALUES (@key, @type)
ON CONFLICT (key, type) DO UPDATE SET
    updated_at = NOW()
RETURNING *;

-- name: DeleteProcessedMark :exec
DELETE FROM processed_mark WHERE key = @key AND type = @type;

-- name: DeleteOldProcessedMarks :exec
DELETE FROM processed_mark WHERE created_at < @cutoff_time;
