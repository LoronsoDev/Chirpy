-- name: AddChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: GetAllChirpsAscOrder :many
SELECT * FROM chirps
ORDER BY created_at ASC;

-- name: GetAllChirpsDescOrder :many
SELECT * FROM chirps
ORDER BY created_at DESC;

-- name: GetChirp :one
SELECT * FROM chirps
WHERE id = $1;

-- name: RemoveChirp :exec
DELETE FROM chirps
WHERE id = $1;

-- name: GetChirpsFromUser :many
SELECT * FROM chirps
WHERE user_id = $1;