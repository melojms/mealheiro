-- name: UpsertTag :one
INSERT INTO tags (name) VALUES (?)
ON CONFLICT (name) DO UPDATE SET name = excluded.name
RETURNING id;

-- name: ClearEntryTags :exec
DELETE FROM entry_tags WHERE entry_id = ?;

-- name: AddEntryTag :exec
INSERT OR IGNORE INTO entry_tags (entry_id, tag_id) VALUES (?, ?);

-- name: ListUsedTags :many
-- pattern is a LIKE prefix pattern escaped with '\'.
SELECT t.name, count(*) AS count
FROM tags t
         JOIN entry_tags et ON et.tag_id = t.id
WHERE t.name LIKE sqlc.arg(pattern) ESCAPE '\'
GROUP BY t.id, t.name
ORDER BY count DESC, t.name
LIMIT 20;
