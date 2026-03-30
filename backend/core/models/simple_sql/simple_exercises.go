package simplesql

import (
	"context"
	"fmt"
	simpleconnection "sport_app/core/models/simple_connection"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InsertExercises(ctx context.Context, conn *pgxpool.Pool) error {
	exercises := []struct {
		Name     string
		Group    string
		Subgroup *string
		Weights  *int
		Injuries bool
	}{
		{"Жим ногами лёжа", "Ноги", String("квадрицепс"), Integer(35), false},
		{"Приседания со штангой", "Ноги", String("квадрицепс"), Integer(20), true},
		{"Приседания в гаке", "Ноги", String("квадрицепс"), Integer(25), false},
		{"Разгибания ног в тренажёре сидя", "Ноги", String("квадрицепс"), Integer(15), false},
		{"Болгарские выпады с гантелями", "Ноги", String("квадрицепс"), Integer(6), false},
		{"Сгибания ног в тренажёре сидя", "Ноги", String("бицепс бедра"), Integer(15), false},
		{"Сгибания ног в тренажёре лёжа", "Ноги", String("бицепс бедра"), Integer(15), false},
		{"Румынская тяга со штангой", "Ноги", String("бицепс бедра"), Integer(15), false},
		{"Румынская тяга с гантелями", "Ноги", String("бицепс бедра"), Integer(5), false},
		{"Гиперэкстензия", "Ноги", String("ягодицы"), Integer(10), false},
		{"Ягодичный мост в тренажёре", "Ноги", String("ягодицы"), Integer(15), false},
		{"Ягодичный мост со штангой", "Ноги", String("ягодицы"), Integer(20), false},
		{"Разведения ног в стороны в тренажёре сидя", "Ноги", String("ягодицы"), Integer(10), false},
		{"Сведения ног друг к другу сидя в тренажёре", "Ноги", String("приводящие"), Integer(10), false},
		{"Отведение ног назад в кроссовере", "Ноги", String("ягодицы"), Integer(3), false},
		{"Отведения ног назад в тренажёре", "Ноги", String("ягодицы"), Integer(10), false},
		{"Подъёмы на носки стоя в тренажёре", "Ноги", String("икры"), Integer(25), false},
		{"Подъёмы на носки сидя в тренажёре", "Ноги", String("икры"), Integer(15), false},
		{"Подъёмы на носки стоя со штангой", "Ноги", String("икры"), Integer(15), false},
		{"Подъёмы на носки стоя с гирей или гантелей на одной ноге", "Ноги", String("икры"), Integer(3), false},
		{"Жим носками в тренажёре для жима ног", "Ноги", String("икры"), Integer(40), false},
		{"Скручивания в тренажёре", "Пресс", nil, Integer(10), false},
		{"Молитва", "Пресс", nil, Integer(10), false},
		{"Махи с гантелями в стороны", "Плечи", String("средняя дельта"), Integer(2), true},
		{"Отведение руки в сторону в кроссовере", "Плечи", String("средняя дельта"), Integer(3), false},
		{"Жим гантелей сидя над головой", "Плечи", String("передняя дельта"), Integer(4), true},
		{"Жим сидя на плечи в тренажёре", "Плечи", String("передняя дельта"), Integer(10), false},
		{"Махи с гантелями в стороны в наклоне сидя", "Плечи", String("задняя дельта"), Integer(2), false},
		{"Махи с гантелями в стороны в наклоне стоя", "Плечи", String("задняя дельта"), Integer(2), true},
		{"Бабочка на плечи", "Плечи", String("задняя дельта"), Integer(5), false},
		{"Разгибания рук в кроссовере", "Руки", String("трицепс"), Integer(5), false},
		{"Фрунзуский жим с гантелями", "Руки", String("трицепс"), Integer(3), true},
		{"Французский жим со штангой", "Руки", String("трицепс"), Integer(5), true},
		{"Разгибание рук с гантелью над головой", "Руки", String("трицепс"), Integer(3), true},
		{"Разгибания рук в кроссовере из-за головы", "Руки", String("трицепс"), Integer(5), false},
		{"Сгибания рук в кроссовере", "Руки", String("бицепс"), Integer(5), false},
		{"Сгибания рук со штангой стоя", "Руки", String("бицепс"), Integer(7), false},
		{"Сгибания рук с гантелями стоя", "Руки", String("бицепс"), Integer(3), false},
		{"Сгибания рук со штангой на скамье Скотта", "Руки", String("бицепс"), Integer(5), true},
		{"Сгибания рук с гантелями на скамье Скотта", "Руки", String("бицепс"), Integer(3), true},
		{"Молотки", "Руки", String("бицепс"), Integer(3), false},
		{"Сгибания рук с гантелями сидя", "Руки", String("бицепс"), Integer(3), false},
		{"Жим в хамере на верх груди", "Грудь", String("грудные верх"), Integer(10), false},
		{"Жим в хамере на середину груди", "Грудь", String("грудные середина"), Integer(15), false},
		{"Жим в хамере на низ груди", "Грудь", String("грудные низ"), Integer(20), false},
		{"Жим штанги на горизонтальной скамье", "Грудь", String("грудные середина"), Integer(20), true},
		{"Жим штанги на наклонной скамье", "Грудь", String("грудные верх"), Integer(15), true},
		{"Жим штанги на скамье с обратным наклоном", "Грудь", String("грудные низ"), Integer(25), true},
		{"Жим гантелей на наклонной скамье", "Грудь", String("грудные верх"), Integer(4), true},
		{"Жим гантелей на горизонтальной скамье", "Грудь", String("грудные середина"), Integer(6), true},
		{"Жим гантелей на скамье с обратным наклоном", "Грудь", String("грудные низ"), Integer(8), true},
		{"Отжимания на брусьях", "Грудь", String("грудные низ"), nil, true},
		{"Отжимания на брусьях с отягощением", "Грудь", String("грудные низ"), Integer(3), true},
		{"Сведение рук в кроссовере на верх груди", "Грудь", String("грудные верх"), Integer(2), false},
		{"Сведение рук в кроссовере на середину груди", "Грудь", String("грудные середина"), Integer(3), false},
		{"Сведение рук в кроссовере на низ груди", "Грудь", String("грудные низ"), Integer(2), false},
		{"Бабочка на грудь", "Грудь", String("грудные середина"), Integer(10), false},
		{"Подтягивания", "Спина", String("тяга сверху широким хватом"), nil, false},
		{"Подтягивания с отягощением", "Спина", String("тяга сверху широким хватом"), Integer(3), true},
		{"Горизонтальная тяга с узкой ручкой", "Спина", String("тяга перед собой узким хватом"), Integer(15), false},
		{"Горизонтальная тяга с широкой ручкой", "Спина", String("тяга перед собой широким хватом"), Integer(12), false},
		{"Вертикальная тяга с узкой ручкой", "Спина", String("тяга сверху узким хватом"), Integer(15), false},
		{"Вертикальная тяга с широкой ручкой", "Спина", String("тяга сверху широким хватом"), Integer(12), false},
		{"Т-гриф", "Спина", String("тяга перед собой широким хватом"), Integer(15), false},
		{"Тяга штанги к поясу", "Спина", String("тяга перед собой широким хватом"), Integer(15), true},
		{"Тяга гантели к поясу", "Спина", String("тяга перед собой узким хватом"), Integer(5), true},
		{"Пуловер", "Спина", String("тяга сверху широким хватом"), Integer(5), false},
	}

	sqlQuery := `INSERT INTO sportapp.exercises (exercises_name, muscular_group, muscular_subgroup, working_weights, safe_for_injuries) VALUES `
	args := []any{}
	for i, u := range exercises {
		if i > 0 {
			sqlQuery += ","
		}
		sqlQuery += fmt.Sprintf("($%d, $%d, $%d, $%d, $%d)",
			i*5+1, i*5+2, i*5+3, i*5+4, i*5+5)
		args = append(args, u.Name, u.Group, u.Subgroup, u.Weights, u.Injuries)
	}

	_, err := conn.Exec(ctx, sqlQuery, args...)
	return err
}

func EnsureExercisesSeeded(ctx context.Context, conn *simpleconnection.ConnectionPool) error {
	var n int64
	if err := conn.QueryRow(ctx, `SELECT COUNT(*) FROM sportapp.exercises`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return InsertExercises(ctx, conn.Pool)
}

func String(s string) *string {
	return &s
}

func Integer(n int) *int {
	return &n
}
