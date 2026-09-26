-- Aggregates used by /api/reports/* and /api/insights. Every query takes an
-- optional payer_id filter (NULL = everyone).

-- name: ReportAggregate :many
-- Entry totals grouped by month, type, leaf category, status and recurring flag.
SELECT CAST(substr(v.date, 1, 7) AS TEXT)                 AS month,
       v.type,
       v.category_id,
       v.top_category_id,
       CAST(v.status = 'pending' AS BOOLEAN)               AS pending,
       CAST(v.template_month IS NOT NULL AS BOOLEAN)       AS recurring,
       CAST(sum(v.amount_cents) AS INTEGER)                AS amount_cents,
       CAST(count(*) AS INTEGER)                           AS entry_count
FROM entry_view v
WHERE v.date >= sqlc.arg('from_date')
  AND v.date <= sqlc.arg('to_date')
  AND (sqlc.narg('payer_id') IS NULL OR v.payer_id = sqlc.narg('payer_id'))
GROUP BY 1, 2, 3, 4, 5, 6
ORDER BY 1, 2, 3, 5, 6;

-- name: ReportCategories :many
-- All categories, archived included, so history keeps its names.
SELECT * FROM categories ORDER BY id;

-- name: ReportTopExpenses :many
SELECT * FROM entry_view v
WHERE v.type = 'expense'
  AND v.date >= sqlc.arg('from_date')
  AND v.date <= sqlc.arg('to_date')
  AND (sqlc.narg('payer_id') IS NULL OR v.payer_id = sqlc.narg('payer_id'))
ORDER BY v.amount_cents DESC, v.date, v.id
LIMIT sqlc.arg('max_rows');

-- name: ReportNonRecurringExpenseSum :one
SELECT CAST(coalesce(sum(e.amount_cents), 0) AS INTEGER) AS amount_cents
FROM entries e
WHERE e.type = 'expense'
  AND e.template_month IS NULL
  AND e.date >= sqlc.arg('from_date')
  AND e.date <= sqlc.arg('to_date')
  AND (sqlc.narg('payer_id') IS NULL OR e.payer_id = sqlc.narg('payer_id'));

-- name: ReportTemplatesStarting :many
SELECT t.id,
       t.type,
       t.category_id,
       t.amount_cents,
       t.variable,
       t.note,
       c.name AS category_name
FROM templates t
         JOIN categories c ON c.id = t.category_id
WHERE t.start_month = sqlc.arg('month')
  AND (sqlc.narg('payer_id') IS NULL OR t.payer_id = sqlc.narg('payer_id'))
ORDER BY t.id;

-- name: ReportEntryYears :many
SELECT DISTINCT CAST(CAST(substr(e.date, 1, 4) AS INTEGER) AS INTEGER) AS year
FROM entries e
WHERE (sqlc.narg('payer_id') IS NULL OR e.payer_id = sqlc.narg('payer_id'))
ORDER BY 1 DESC;
