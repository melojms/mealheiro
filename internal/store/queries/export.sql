-- name: ListExportEntries :many
SELECT * FROM entry_view
WHERE date >= sqlc.arg(from_date)
  AND date <= sqlc.arg(to_date)
ORDER BY date, id;
