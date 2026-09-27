-- Expenses are shared between the people by default; `personal` opts one out of
-- the shared-expenses split. Templates carry the flag onto generated entries.

-- +goose Up
ALTER TABLE entries ADD COLUMN personal BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE templates ADD COLUMN personal BOOLEAN NOT NULL DEFAULT FALSE;

DROP VIEW entry_view;
CREATE VIEW entry_view AS
SELECT e.id,
       e.type,
       e.date,
       e.amount_cents,
       e.category_id,
       c.name                           AS category_name,
       c.parent_id                      AS parent_category_id,
       coalesce(pc.name, '')            AS parent_category_name, -- '' when top-level
       coalesce(c.parent_id, c.id)      AS top_category_id,
       e.payer_id,
       pe.name                          AS payer_name,
       pe.kind                          AS payer_kind,
       e.personal,
       e.note,
       e.status,
       e.template_id,
       e.template_month,
       cast(coalesce((SELECT group_concat(name, ',')
                      FROM (SELECT t.name
                            FROM entry_tags et
                                     JOIN tags t ON t.id = et.tag_id
                            WHERE et.entry_id = e.id
                            ORDER BY t.name)), '') AS TEXT) AS tags_csv,
       e.created_at,
       e.updated_at
FROM entries e
         JOIN categories c ON c.id = e.category_id
         LEFT JOIN categories pc ON pc.id = c.parent_id
         JOIN people pe ON pe.id = e.payer_id;

-- +goose Down
DROP VIEW entry_view;
CREATE VIEW entry_view AS
SELECT e.id,
       e.type,
       e.date,
       e.amount_cents,
       e.category_id,
       c.name                           AS category_name,
       c.parent_id                      AS parent_category_id,
       coalesce(pc.name, '')            AS parent_category_name, -- '' when top-level
       coalesce(c.parent_id, c.id)      AS top_category_id,
       e.payer_id,
       pe.name                          AS payer_name,
       e.note,
       e.status,
       e.template_id,
       e.template_month,
       cast(coalesce((SELECT group_concat(name, ',')
                      FROM (SELECT t.name
                            FROM entry_tags et
                                     JOIN tags t ON t.id = et.tag_id
                            WHERE et.entry_id = e.id
                            ORDER BY t.name)), '') AS TEXT) AS tags_csv,
       e.created_at,
       e.updated_at
FROM entries e
         JOIN categories c ON c.id = e.category_id
         LEFT JOIN categories pc ON pc.id = c.parent_id
         JOIN people pe ON pe.id = e.payer_id;

ALTER TABLE templates DROP COLUMN personal;
ALTER TABLE entries DROP COLUMN personal;
