-- name: GetCategory :one
SELECT * FROM categories WHERE id = ?;

-- name: ListCategoriesWithCounts :many
-- usage_count: entries dated in [since, until]; for a top-level category this
-- includes its children. entry_count: all-time entries on this category only.
SELECT c.id, c.parent_id, c.type, c.name, c.icon, c.color, c.archived,
       (SELECT count(*)
        FROM entries e
                 JOIN categories ec ON ec.id = e.category_id
        WHERE (ec.id = c.id OR ec.parent_id = c.id)
          AND e.date >= sqlc.arg(since)
          AND e.date <= sqlc.arg(until))                           AS usage_count,
       (SELECT count(*) FROM entries e WHERE e.category_id = c.id) AS entry_count
FROM categories c
ORDER BY usage_count DESC, c.name COLLATE NOCASE, c.id;

-- name: GetCategoryWithCounts :one
SELECT c.id, c.parent_id, c.type, c.name, c.icon, c.color, c.archived,
       (SELECT count(*)
        FROM entries e
                 JOIN categories ec ON ec.id = e.category_id
        WHERE (ec.id = c.id OR ec.parent_id = c.id)
          AND e.date >= sqlc.arg(since)
          AND e.date <= sqlc.arg(until))                           AS usage_count,
       (SELECT count(*) FROM entries e WHERE e.category_id = c.id) AS entry_count
FROM categories c
WHERE c.id = sqlc.arg(id);

-- name: ListTopLevelColors :many
SELECT color FROM categories WHERE type = ? AND parent_id IS NULL;

-- name: CreateCategory :one
INSERT INTO categories (parent_id, type, name, icon, color, created_at)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: UpdateCategory :exec
UPDATE categories SET name = ?, icon = ?, color = ?, archived = ? WHERE id = ?;

-- name: SetCategoryArchived :exec
UPDATE categories SET archived = ? WHERE id = ?;

-- name: ArchiveChildCategories :exec
UPDATE categories SET archived = TRUE WHERE parent_id = ?;

-- name: CountCategoryReferences :one
SELECT (SELECT count(*) FROM entries e WHERE e.category_id = sqlc.arg(id))    AS entries,
       (SELECT count(*) FROM templates t WHERE t.category_id = sqlc.arg(id))  AS templates,
       (SELECT count(*) FROM budgets b WHERE b.category_id = sqlc.arg(id))    AS budgets,
       (SELECT count(*) FROM categories ch WHERE ch.parent_id = sqlc.arg(id)) AS children;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = ?;
