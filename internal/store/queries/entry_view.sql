-- name: GetEntryView :one
SELECT * FROM entry_view WHERE id = ?;

-- name: ListPendingEntryViews :many
SELECT * FROM entry_view WHERE status = 'pending' ORDER BY date, id;
