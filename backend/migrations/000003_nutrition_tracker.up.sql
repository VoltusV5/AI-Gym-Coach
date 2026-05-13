ALTER TABLE sportapp.nutrition_entries
    ADD COLUMN IF NOT EXISTS meal_type VARCHAR(20) NOT NULL DEFAULT 'snack',
    ADD COLUMN IF NOT EXISTS calories DOUBLE PRECISION NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS dish_id INTEGER;

CREATE TABLE IF NOT EXISTS sportapp.nutrition_dishes (
    id SERIAL PRIMARY KEY,
    title VARCHAR(250) NOT NULL UNIQUE,
    protein_g DOUBLE PRECISION NOT NULL DEFAULT 0,
    fat_g DOUBLE PRECISION NOT NULL DEFAULT 0,
    carbs_g DOUBLE PRECISION NOT NULL DEFAULT 0,
    calories DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_by_user_id INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_nutrition_dishes_title
    ON sportapp.nutrition_dishes (title);

CREATE TABLE IF NOT EXISTS sportapp.nutrition_water_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    logged_on DATE NOT NULL,
    amount_ml INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    UNIQUE(user_id, logged_on),
    FOREIGN KEY (user_id) REFERENCES sportapp.users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sportapp.nutrition_weight_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    logged_on DATE NOT NULL,
    weight_kg DOUBLE PRECISION NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    UNIQUE(user_id, logged_on),
    FOREIGN KEY (user_id) REFERENCES sportapp.users(id) ON DELETE CASCADE
);

UPDATE sportapp.nutrition_entries
SET calories = protein_g * 4 + fat_g * 9 + carbs_g * 4
WHERE calories = 0;