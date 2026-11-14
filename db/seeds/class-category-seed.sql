BEGIN;

INSERT INTO class_categories (category, description)
VALUES
    ('private', 'Private class'),
    ('group', 'Group class');

INSERT INTO class_catalogues (name, category_id, description, additional_info, price, currency, created_at, updated_at)
VALUES
    ('Private 1 on 1', 1, 'Private 1 on 1', '{}'::JSONB, 150000, 'IDR', NOW(), NOW()),
    ('Group Elementary JLPT N5', 2, 'Group jlpt n5', '{}'::JSONB, 200000, 'IDR', NOW(), NOW()),
    ('Group Elementary JLPT N4', 2, 'Group jlpt n4', '{}'::JSONB, 250000, 'IDR', NOW(), NOW()),
    ('Group Elementary JLPT N3', 2, 'Group jlpt n3', '{}'::JSONB, 200000, 'IDR', NOW(), NOW()),
    ('Group Elementary JLPT N2', 2, 'Group jlpt n2', '{}'::JSONB, 250000, 'IDR', NOW(), NOW()),
    ('Group Elementary JLPT N1', 2, 'Group jlpt n1', '{}'::JSONB, 250000, 'IDR', NOW(), NOW());

COMMIT;