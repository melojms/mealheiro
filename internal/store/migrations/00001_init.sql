-- +goose Up
CREATE TABLE people (
    id         INTEGER PRIMARY KEY,
    name       TEXT    NOT NULL,
    kind       TEXT    NOT NULL CHECK (kind IN ('person', 'joint')),
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

-- Two levels max: parent_id must reference a top-level category (enforced in the app).
CREATE TABLE categories (
    id         INTEGER PRIMARY KEY,
    parent_id  INTEGER REFERENCES categories (id) ON DELETE RESTRICT,
    type       TEXT    NOT NULL CHECK (type IN ('expense', 'income', 'investment')),
    name       TEXT    NOT NULL,
    icon       TEXT    NOT NULL,
    color      TEXT    NOT NULL,
    archived   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE UNIQUE INDEX categories_name_uniq ON categories (type, coalesce(parent_id, 0), name COLLATE NOCASE);
CREATE INDEX categories_parent_idx ON categories (parent_id);

-- Monthly recurring templates. Generated entries are dated the 1st of each month.
CREATE TABLE templates (
    id           INTEGER PRIMARY KEY,
    type         TEXT    NOT NULL CHECK (type IN ('expense', 'income', 'investment')),
    category_id  INTEGER NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
    payer_id     INTEGER NOT NULL REFERENCES people (id) ON DELETE RESTRICT,
    amount_cents INTEGER NOT NULL CHECK (amount_cents > 0),
    variable     BOOLEAN NOT NULL DEFAULT FALSE,
    note         TEXT    NOT NULL DEFAULT '',
    start_month  TEXT    NOT NULL, -- YYYY-MM
    end_month    TEXT,             -- YYYY-MM, inclusive, NULL = open ended
    active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at   TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

-- One row per (template, month) ever generated. Makes generation idempotent even
-- after the user deletes a generated entry.
CREATE TABLE template_runs (
    template_id INTEGER NOT NULL REFERENCES templates (id) ON DELETE CASCADE,
    month       TEXT    NOT NULL, -- YYYY-MM
    created_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    PRIMARY KEY (template_id, month)
);

CREATE TABLE entries (
    id             INTEGER PRIMARY KEY,
    type           TEXT    NOT NULL CHECK (type IN ('expense', 'income', 'investment')),
    date           TEXT    NOT NULL, -- YYYY-MM-DD, local date
    amount_cents   INTEGER NOT NULL CHECK (amount_cents > 0),
    category_id    INTEGER NOT NULL REFERENCES categories (id) ON DELETE RESTRICT,
    payer_id       INTEGER NOT NULL REFERENCES people (id) ON DELETE RESTRICT,
    note           TEXT    NOT NULL DEFAULT '',
    status         TEXT    NOT NULL DEFAULT 'confirmed' CHECK (status IN ('confirmed', 'pending')),
    template_id    INTEGER REFERENCES templates (id) ON DELETE SET NULL,
    template_month TEXT, -- YYYY-MM when generated from a template; NOT NULL means "recurring"
    created_at     TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at     TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX entries_date_idx ON entries (date);
CREATE INDEX entries_category_idx ON entries (category_id);
CREATE INDEX entries_type_date_idx ON entries (type, date);
CREATE INDEX entries_status_idx ON entries (status);
CREATE INDEX entries_template_idx ON entries (template_id);

CREATE TABLE tags (
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE -- lowercase-normalized by the app
);

CREATE TABLE entry_tags (
    entry_id INTEGER NOT NULL REFERENCES entries (id) ON DELETE CASCADE,
    tag_id   INTEGER NOT NULL REFERENCES tags (id) ON DELETE CASCADE,
    PRIMARY KEY (entry_id, tag_id)
);

CREATE INDEX entry_tags_tag_idx ON entry_tags (tag_id);

-- Budgets are versioned by month: the row with the greatest effective_from <= month applies.
-- category_id NULL = overall monthly spend cap. amount_cents NULL = budget removed from that month on.
CREATE TABLE budgets (
    id             INTEGER PRIMARY KEY,
    category_id    INTEGER REFERENCES categories (id) ON DELETE CASCADE,
    effective_from TEXT    NOT NULL, -- YYYY-MM
    amount_cents   INTEGER CHECK (amount_cents IS NULL OR amount_cents > 0),
    created_at     TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE UNIQUE INDEX budgets_uniq ON budgets (coalesce(category_id, 0), effective_from);

-- +goose Down
DROP TABLE budgets;
DROP TABLE entry_tags;
DROP TABLE tags;
DROP TABLE entries;
DROP TABLE template_runs;
DROP TABLE templates;
DROP TABLE categories;
DROP TABLE people;
