-- Rename the seeded House subcategory, leaving it alone if the user already renamed it.

-- +goose Up
UPDATE categories SET name = 'Management' WHERE id = 52 AND name = 'Condomínio';

-- +goose Down
UPDATE categories SET name = 'Condomínio' WHERE id = 52 AND name = 'Management';
