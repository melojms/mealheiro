-- +goose Up
INSERT INTO people (id, name, kind, sort_order) VALUES
    (1, 'Person A', 'person', 1),
    (2, 'Person B', 'person', 2),
    (3, 'Joint', 'joint', 3);

-- Top-level expense categories.
INSERT INTO categories (id, parent_id, type, name, icon, color) VALUES
    (1, NULL, 'expense', 'Water', 'droplet', '#0ea5e9'),
    (2, NULL, 'expense', 'Electricity', 'zap', '#eab308'),
    (3, NULL, 'expense', 'Gas', 'flame', '#f97316'),
    (4, NULL, 'expense', 'Internet & Phone', 'wifi', '#6366f1'),
    (5, NULL, 'expense', 'House', 'house', '#8b5cf6'),
    (6, NULL, 'expense', 'Groceries', 'shopping-cart', '#10b981'),
    (7, NULL, 'expense', 'Subscriptions', 'repeat', '#ec4899'),
    (8, NULL, 'expense', 'Car', 'car', '#ef4444'),
    (9, NULL, 'expense', 'Transport', 'bus', '#14b8a6'),
    (10, NULL, 'expense', 'Clothes', 'shirt', '#d946ef'),
    (11, NULL, 'expense', 'Technology', 'laptop', '#3b82f6'),
    (12, NULL, 'expense', 'Eating out', 'utensils', '#f59e0b'),
    (13, NULL, 'expense', 'Health', 'heart-pulse', '#f43f5e'),
    (14, NULL, 'expense', 'Leisure', 'party-popper', '#a855f7'),
    (15, NULL, 'expense', 'Travel', 'plane', '#06b6d4'),
    (16, NULL, 'expense', 'Gifts', 'gift', '#84cc16'),
    (17, NULL, 'expense', 'Personal care', 'sparkles', '#fb7185'),
    (18, NULL, 'expense', 'Education', 'graduation-cap', '#22c55e'),
    (19, NULL, 'expense', 'Other', 'circle-ellipsis', '#64748b');

-- Expense subcategories (inherit the parent's color).
INSERT INTO categories (id, parent_id, type, name, icon, color) VALUES
    (50, 5, 'expense', 'Rent', 'key-round', '#8b5cf6'),
    (51, 5, 'expense', 'Mortgage', 'landmark', '#8b5cf6'),
    (52, 5, 'expense', 'Condomínio', 'building-2', '#8b5cf6'),
    (53, 5, 'expense', 'Insurance', 'shield-check', '#8b5cf6'),
    (54, 5, 'expense', 'Maintenance', 'wrench', '#8b5cf6'),
    (55, 9, 'expense', 'Public', 'train-front', '#14b8a6'),
    (56, 9, 'expense', 'Taxi/Uber', 'car-taxi-front', '#14b8a6');

-- Income.
INSERT INTO categories (id, parent_id, type, name, icon, color) VALUES
    (100, NULL, 'income', 'Salary', 'briefcase', '#10b981'),
    (101, NULL, 'income', 'Bonus', 'award', '#22c55e'),
    (102, NULL, 'income', 'Refund', 'undo-2', '#14b8a6'),
    (103, NULL, 'income', 'Side income', 'coins', '#84cc16');

-- Investments.
INSERT INTO categories (id, parent_id, type, name, icon, color) VALUES
    (200, NULL, 'investment', 'ETFs/Stocks', 'trending-up', '#6366f1'),
    (201, NULL, 'investment', 'Savings account', 'piggy-bank', '#0ea5e9'),
    (202, NULL, 'investment', 'BTC', 'bitcoin', '#f59e0b');

-- +goose Down
DELETE FROM categories;
DELETE FROM people;
