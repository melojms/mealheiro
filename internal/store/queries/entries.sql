-- name: GetEntry :one
SELECT * FROM entries WHERE id = ?;

-- name: CreateEntry :one
INSERT INTO entries (type, date, amount_cents, category_id, payer_id, personal, note, status, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, 'confirmed', sqlc.arg(now), sqlc.arg(now))
RETURNING id;

-- name: UpdateEntry :exec
UPDATE entries
SET type         = ?,
    date         = ?,
    amount_cents = ?,
    category_id  = ?,
    payer_id     = ?,
    personal     = ?,
    note         = ?,
    updated_at   = ?
WHERE id = ?;

-- name: ConfirmEntry :exec
UPDATE entries SET status = 'confirmed', amount_cents = ?, updated_at = ? WHERE id = ?;

-- name: DeleteEntry :execrows
DELETE FROM entries WHERE id = ?;
