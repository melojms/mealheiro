-- name: ListPeople :many
SELECT * FROM people ORDER BY sort_order, id;

-- name: GetPerson :one
SELECT * FROM people WHERE id = ?;

-- name: RenamePerson :one
UPDATE people SET name = ? WHERE id = ? RETURNING *;
