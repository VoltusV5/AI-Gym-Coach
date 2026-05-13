package users_postgres_repository

import (
	"context"
	"fmt"
)

func (r *UsersRepository) EnsureExercisesSeeded(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, r.pool.OpTimeout())
	defer cancel()

	var n int64
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.exercises`).Scan(&n); err != nil {
		return fmt.Errorf("count exercises: %w", err)
	}
	if n > 0 {
		return nil
	}

	return r.insertExercises(ctx)
}

func (r *UsersRepository) insertExercises(ctx context.Context) error {
	exercises := []struct {
		Name     string
		Group    string
		Subgroup *string
		Weights  *int
		Injuries bool
	}{
		{"Жим ногами лёжа", "Ноги", strPtr("квадрицепс"), intPtr(35), false},
		{"Приседания со штангой", "Ноги", strPtr("квадрицепс"), intPtr(20), true},
		{"Приседания в гаке", "Ноги", strPtr("квадрицепс"), intPtr(25), false},
		{"Разгибания ног в тренажёре сидя", "Ноги", strPtr("квадрицепс"), intPtr(15), false},
		{"Болгарские выпады с гантелями", "Ноги", strPtr("квадрицепс"), intPtr(6), false},
		{"Сгибания ног в тренажёре сидя", "Ноги", strPtr("бицепс бедра"), intPtr(15), false},
		{"Сгибания ног в тренажёре лёжа", "Ноги", strPtr("бицепс бедра"), intPtr(15), false},
		{"Румынская тяга со штангой", "Ноги", strPtr("бицепс бедра"), intPtr(15), false},
		{"Румынская тяга с гантелями", "Ноги", strPtr("бицепс бедра"), intPtr(5), false},
		{"Гиперэкстензия", "Ноги", strPtr("ягодицы"), intPtr(10), false},
		{"Ягодичный мост в тренажёре", "Ноги", strPtr("ягодицы"), intPtr(15), false},
		{"Ягодичный мост со штангой", "Ноги", strPtr("ягодицы"), intPtr(20), false},
		{"Разведения ног в стороны в тренажёре сидя", "Ноги", strPtr("ягодицы"), intPtr(10), false},
		{"Сведения ног друг к другу сидя в тренажёре", "Ноги", strPtr("приводящие"), intPtr(10), false},
		{"Отведение ног назад в кроссовере", "Ноги", strPtr("ягодицы"), intPtr(3), false},
		{"Отведения ног назад в тренажёре", "Ноги", strPtr("ягодицы"), intPtr(10), false},
		{"Подъёмы на носки стоя в тренажёре", "Ноги", strPtr("икры"), intPtr(25), false},
		{"Подъёмы на носки сидя в тренажёре", "Ноги", strPtr("икры"), intPtr(15), false},
		{"Подъёмы на носки стоя со штангой", "Ноги", strPtr("икры"), intPtr(15), false},
		{"Подъёмы на носки стоя с гирей или гантелей на одной ноге", "Ноги", strPtr("икры"), intPtr(3), false},
		{"Жим носками в тренажёре для жима ног", "Ноги", strPtr("икры"), intPtr(40), false},
		{"Скручивания в тренажёре", "Пресс", nil, intPtr(10), false},
		{"Молитва", "Пресс", nil, intPtr(10), false},
		{"Махи с гантелями в стороны", "Плечи", strPtr("средняя дельта"), intPtr(2), true},
		{"Отведение руки в сторону в кроссовере", "Плечи", strPtr("средняя дельта"), intPtr(3), false},
		{"Жим гантелей сидя над головой", "Плечи", strPtr("передняя дельта"), intPtr(4), true},
		{"Жим сидя на плечи в тренажёре", "Плечи", strPtr("передняя дельта"), intPtr(10), false},
		{"Махи с гантелями в стороны в наклоне сидя", "Плечи", strPtr("задняя дельта"), intPtr(2), false},
		{"Махи с гантелями в стороны в наклоне стоя", "Плечи", strPtr("задняя дельта"), intPtr(2), true},
		{"Бабочка на плечи", "Плечи", strPtr("задняя дельта"), intPtr(5), false},
		{"Разгибания рук в кроссовере", "Руки", strPtr("трицепс"), intPtr(5), false},
		{"Фрунзуский жим с гантелями", "Руки", strPtr("трицепс"), intPtr(3), true},
		{"Французский жим со штангой", "Руки", strPtr("трицепс"), intPtr(5), true},
		{"Разгибание рук с гантелью над головой", "Руки", strPtr("трицепс"), intPtr(3), true},
		{"Разгибания рук в кроссовере из-за головы", "Руки", strPtr("трицепс"), intPtr(5), false},
		{"Сгибания рук в кроссовере", "Руки", strPtr("бицепс"), intPtr(5), false},
		{"Сгибания рук со штангой стоя", "Руки", strPtr("бицепс"), intPtr(7), false},
		{"Сгибания рук с гантелями стоя", "Руки", strPtr("бицепс"), intPtr(3), false},
		{"Сгибания рук со штангой на скамье Скотта", "Руки", strPtr("бицепс"), intPtr(5), true},
		{"Сгибания рук с гантелями на скамье Скотта", "Руки", strPtr("бицепс"), intPtr(3), true},
		{"Молотки", "Руки", strPtr("бицепс"), intPtr(3), false},
		{"Сгибания рук с гантелями сидя", "Руки", strPtr("бицепс"), intPtr(3), false},
		{"Жим в хамере на верх груди", "Грудь", strPtr("грудные верх"), intPtr(10), false},
		{"Жим в хамере на середину груди", "Грудь", strPtr("грудные середина"), intPtr(15), false},
		{"Жим в хамере на низ груди", "Грудь", strPtr("грудные низ"), intPtr(20), false},
		{"Жим штанги на горизонтальной скамье", "Грудь", strPtr("грудные середина"), intPtr(20), true},
		{"Жим штанги на наклонной скамье", "Грудь", strPtr("грудные верх"), intPtr(15), true},
		{"Жим штанги на скамье с обратным наклоном", "Грудь", strPtr("грудные низ"), intPtr(25), true},
		{"Жим гантелей на наклонной скамье", "Грудь", strPtr("грудные верх"), intPtr(4), true},
		{"Жим гантелей на горизонтальной скамье", "Грудь", strPtr("грудные середина"), intPtr(6), true},
		{"Жим гантелей на скамье с обратным наклоном", "Грудь", strPtr("грудные низ"), intPtr(8), true},
		{"Отжимания на брусьях", "Грудь", strPtr("грудные низ"), nil, true},
		{"Отжимания на брусьях с отягощением", "Грудь", strPtr("грудные низ"), intPtr(3), true},
		{"Сведение рук в кроссовере на верх груди", "Грудь", strPtr("грудные верх"), intPtr(2), false},
		{"Сведение рук в кроссовере на середину груди", "Грудь", strPtr("грудные середина"), intPtr(3), false},
		{"Сведение рук в кроссовере на низ груди", "Грудь", strPtr("грудные низ"), intPtr(2), false},
		{"Бабочка на грудь", "Грудь", strPtr("грудные середина"), intPtr(10), false},
		{"Подтягивания", "Спина", strPtr("тяга сверху широким хватом"), nil, false},
		{"Подтягивания с отягощением", "Спина", strPtr("тяга сверху широким хватом"), intPtr(3), true},
		{"Горизонтальная тяга с узкой ручкой", "Спина", strPtr("тяга перед собой узким хватом"), intPtr(15), false},
		{"Горизонтальная тяга с широкой ручкой", "Спина", strPtr("тяга перед собой широким хватом"), intPtr(12), false},
		{"Вертикальная тяга с узкой ручкой", "Спина", strPtr("тяга сверху узким хватом"), intPtr(15), false},
		{"Вертикальная тяга с широкой ручкой", "Спина", strPtr("тяга сверху широким хватом"), intPtr(12), false},
		{"Т-гриф", "Спина", strPtr("тяга перед собой широким хватом"), intPtr(15), false},
		{"Тяга штанги к поясу", "Спина", strPtr("тяга перед собой широким хватом"), intPtr(15), true},
		{"Тяга гантели к поясу", "Спина", strPtr("тяга перед собой узким хватом"), intPtr(5), true},
		{"Пуловер", "Спина", strPtr("тяга сверху широким хватом"), intPtr(5), false},
	}

	query := `INSERT INTO sportapp.exercises (exercises_name, muscular_group, muscular_subgroup, working_weights, safe_for_injuries) VALUES `
	args := make([]any, 0, len(exercises)*5)
	for i, ex := range exercises {
		if i > 0 {
			query += ","
		}
		query += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)",
			i*5+1, i*5+2, i*5+3, i*5+4, i*5+5)
		args = append(args, ex.Name, ex.Group, ex.Subgroup, ex.Weights, ex.Injuries)
	}

	if _, err := r.pool.Exec(ctx, query, args...); err != nil {
		return fmt.Errorf("seed exercises: %w", err)
	}

	return nil
}

func strPtr(s string) *string { return &s }

func intPtr(n int) *int { return &n }