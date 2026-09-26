-- name: ListTemplateViews :many
SELECT t.id,
       t.type,
       t.category_id,
       c.name                                                                              AS category_name,
       coalesce(pc.name, '')                                                                AS parent_category_name, -- '' when top-level
       t.payer_id,
       p.name                                                                              AS payer_name,
       t.amount_cents,
       t.variable,
       t.note,
       t.start_month,
       t.end_month,
       t.active,
       cast(coalesce((SELECT max(r.month) FROM template_runs r WHERE r.template_id = t.id), '') AS TEXT) AS last_generated_month
FROM templates t
         JOIN categories c ON c.id = t.category_id
         LEFT JOIN categories pc ON pc.id = c.parent_id
         JOIN people p ON p.id = t.payer_id
ORDER BY t.active DESC, t.type, c.name COLLATE NOCASE, t.id;

-- name: GetTemplate :one
SELECT * FROM templates WHERE id = ?;

-- name: CreateTemplate :one
INSERT INTO templates (type, category_id, payer_id, amount_cents, variable, note, start_month, end_month, active)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
RETURNING id;

-- name: UpdateTemplate :execrows
UPDATE templates
SET type         = ?,
    category_id  = ?,
    payer_id     = ?,
    amount_cents = ?,
    variable     = ?,
    note         = ?,
    start_month  = ?,
    end_month    = ?,
    active       = ?,
    updated_at   = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
WHERE id = ?;

-- name: DeleteTemplate :execrows
DELETE FROM templates WHERE id = ?;

-- name: GetTemplateCategory :one
SELECT * FROM categories WHERE id = ?;

-- name: ListActiveTemplates :many
SELECT * FROM templates WHERE active ORDER BY id;

-- name: ListTemplateRunMonths :many
SELECT month FROM template_runs WHERE template_id = ?;

-- name: LatestTemplateEntryAmount :one
SELECT amount_cents FROM entries WHERE template_id = ? ORDER BY date DESC, id DESC LIMIT 1;

-- name: InsertGeneratedEntry :exec
INSERT INTO entries (type, date, amount_cents, category_id, payer_id, note, status, template_id, template_month)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?);

-- name: InsertTemplateRun :exec
INSERT INTO template_runs (template_id, month) VALUES (?, ?);
