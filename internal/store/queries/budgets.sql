-- name: ListBudgetsUpTo :many
-- All budget rows effective on or before the month, oldest first; the last row
-- per category is the effective one.
SELECT * FROM budgets WHERE effective_from <= ? ORDER BY effective_from, id;

-- name: DeleteBudgetAt :exec
DELETE FROM budgets WHERE category_id IS sqlc.narg(category_id) AND effective_from = sqlc.arg(effective_from);

-- name: InsertBudget :exec
INSERT INTO budgets (category_id, effective_from, amount_cents) VALUES (?, ?, ?);

-- name: ListTopExpenseCategories :many
SELECT * FROM categories WHERE type = 'expense' AND parent_id IS NULL ORDER BY id;

-- name: MonthExpenseByTopCategory :many
SELECT cast(coalesce(c.parent_id, c.id) AS INTEGER)                                       AS top_category_id,
       cast(sum(e.amount_cents) AS INTEGER)                                               AS spent_cents,
       cast(sum(CASE WHEN e.status = 'pending' THEN e.amount_cents ELSE 0 END) AS INTEGER) AS pending_cents
FROM entries e
         JOIN categories c ON c.id = e.category_id
WHERE e.type = 'expense'
  AND e.date >= sqlc.arg(from_date)
  AND e.date <= sqlc.arg(to_date)
GROUP BY coalesce(c.parent_id, c.id);
