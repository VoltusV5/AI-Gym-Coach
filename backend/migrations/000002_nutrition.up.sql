CREATE TABLE sportapp.nutrition_entries (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    title VARCHAR(250) NOT NULL,
    protein_g DOUBLE PRECISION NOT NULL DEFAULT 0,
    fat_g DOUBLE PRECISION NOT NULL DEFAULT 0,
    carbs_g DOUBLE PRECISION NOT NULL DEFAULT 0,
    consumed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    FOREIGN KEY (user_id) REFERENCES sportapp.users(id) ON DELETE CASCADE
);

CREATE INDEX idx_nutrition_entries_user_consumed_at
    ON sportapp.nutrition_entries(user_id, consumed_at DESC);

CREATE TABLE sportapp.nutrition_favorites (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    title VARCHAR(250) NOT NULL,
    protein_g DOUBLE PRECISION NOT NULL DEFAULT 0,
    fat_g DOUBLE PRECISION NOT NULL DEFAULT 0,
    carbs_g DOUBLE PRECISION NOT NULL DEFAULT 0,
    unit_type VARCHAR(20) NOT NULL DEFAULT 'gram',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ,
    FOREIGN KEY (user_id) REFERENCES sportapp.users(id) ON DELETE CASCADE
);

CREATE TABLE sportapp.nutrition_goals (
    user_id INTEGER PRIMARY KEY,
    protein_g DOUBLE PRECISION NOT NULL DEFAULT 0,
    fat_g DOUBLE PRECISION NOT NULL DEFAULT 0,
    carbs_g DOUBLE PRECISION NOT NULL DEFAULT 0,
    calories DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES sportapp.users(id) ON DELETE CASCADE
);

CREATE TABLE sportapp.nutrition_catalog (
    id SERIAL PRIMARY KEY,
    title VARCHAR(250) NOT NULL UNIQUE,
    protein_g DOUBLE PRECISION NOT NULL DEFAULT 0,
    fat_g DOUBLE PRECISION NOT NULL DEFAULT 0,
    carbs_g DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ
);

INSERT INTO sportapp.nutrition_catalog (title, protein_g, fat_g, carbs_g) VALUES
('Куриная грудка', 22.5, 1.9, 0),
('Филе индейки', 21.6, 3.9, 0),
('Говядина постная', 21.5, 5.7, 0.8),
('Лосось', 20.3, 13.1, 0),
('Тунец в собственном соку', 19, 0.6, 0.08),
('Яйцо куриное', 12.4, 8.65, 0.96),
('Яичный белок', 10.7, 0, 2.36),
('Творог 5%', 17, 5, 3),
('Йогурт греческий', 10.3, 0.17, 3.64),
('Сыр моцарелла', 23.7, 17.8, 4.44),
('Овсянка', 13.5, 5.9, 68.7),
('Гречка', 11.1, 3.0, 71.1),
('Рис белый', 7.0, 1.0, 80.3),
('Рис бурый', 7.25, 3.31, 76.7),
('Булгур', 11.8, 2.42, 75.9),
('Чечевица сухая', 23.6, 1.92, 62.2),
('Нут сухой', 21.3, 6.27, 60.4),
('Фасоль красная сухая', 21.3, 1.16, 60),
('Банан', 0.74, 0.29, 23),
('Яблоко', 0.2, 0.15, 14.8),
('Киви', 1.06, 0.44, 14),
('Апельсин', 0.91, 0.15, 11.8),
('Клубника', 0.64, 0.22, 7.96),
('Черника', 0.7, 0.31, 14.6),
('Картофель', 2.27, 0.36, 17.8),
('Батат', 1.58, 0.38, 17.3),
('Брокколи', 2.57, 0.07, 6.27),
('Цветная капуста', 1.64, 0.24, 4.72),
('Огурец', 0.63, 0.18, 2.95),
('Томаты', 0.84, 0.5, 3.32),
('Авокадо', 1.81, 20.3, 8.32),
('Миндаль', 21.5, 51.1, 20),
('Грецкий орех', 14.6, 69.7, 10.9),
('Арахисовая паста', 24.0, 49.4, 22.7),
('Оливковое масло', 0, 93.7, 0),
('Подсолнечное масло', 0, 93.2, 0),
('Хлеб цельнозерновой', 12.3, 2.98, 43.1),
('Хлеб белый', 9.43, 3.45, 49.2),
('Семена чиа', 17.0, 32.9, 38.3),
('Семена льна', 18.0, 37.3, 34.4)
ON CONFLICT (title) DO NOTHING;
