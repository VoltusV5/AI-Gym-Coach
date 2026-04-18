ALTER TABLE sportapp.nutrition_entries
    ADD COLUMN IF NOT EXISTS grams DOUBLE PRECISION NOT NULL DEFAULT 100;

ALTER TABLE sportapp.nutrition_dishes
    ADD COLUMN IF NOT EXISTS base_grams DOUBLE PRECISION NOT NULL DEFAULT 100;

UPDATE sportapp.nutrition_entries
SET grams = 100
WHERE grams <= 0;

UPDATE sportapp.nutrition_dishes
SET base_grams = 100
WHERE base_grams <= 0;
