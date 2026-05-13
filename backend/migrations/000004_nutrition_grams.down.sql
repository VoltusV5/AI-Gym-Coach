ALTER TABLE sportapp.nutrition_dishes
    DROP COLUMN IF EXISTS base_grams;

ALTER TABLE sportapp.nutrition_entries
    DROP COLUMN IF EXISTS grams;