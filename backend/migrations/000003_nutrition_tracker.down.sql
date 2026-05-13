DROP TABLE IF EXISTS sportapp.nutrition_weight_logs;
DROP TABLE IF EXISTS sportapp.nutrition_water_logs;
DROP INDEX IF EXISTS sportapp.idx_nutrition_dishes_title;
DROP TABLE IF EXISTS sportapp.nutrition_dishes;

ALTER TABLE sportapp.nutrition_entries
    DROP COLUMN IF EXISTS dish_id,
    DROP COLUMN IF EXISTS calories,
    DROP COLUMN IF EXISTS meal_type;