package simplesql

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InsertExercises(ctx context.Context, conn *pgxpool.Pool) error {
	exercises := []struct {
		Name     string
		Group    string
		Subgroup *string
		Weights  int
		Injuries bool
	}{
		{"Жим ногами лёжа", "Квадрицепсы", String("Ягодицы"), 35, false},
		{"Приседания со штангой", "Квадрицепсы", String("Ягодицы"), 20, true},
		{"Приседания в гаке", "Квадрицепсы", String("Ягодицы"), 25, false},
		{"Разгибания ног в тренажёре сидя", "Квадрицепсы", nil, 15, false},
		{"Болгарские выпады с гантелями", "Квадрицепсы", String("Ягодицы"), 6, false},
		{"Сгибания ног в тренажёре сидя", "Бицепс бедра", String("Икры"), 15, false},
		{"Сгибания ног в тренажёре лёжа", "Бицепс бедра", String("Икры"), 15, false},
		{"Румынская тяга со штангой", "Бицепс бедра", String("Ягодицы"), 15, false},
		{"Румынская тяга с гантелями", "Бицепс бедра", String("Ягодицы"), 5, false},
		{"Гиперэкстензия", "Ягодицы", String("Бицепс бедра"), 10, false},
		{"Ягодичный мост в тренажёре", "Ягодицы", String("Бицепс бедра"), 15, false},
		{"Ягодичный мост со штангой", "Ягодицы", String("Бицепс бедра"), 20, false},
		{"Разведения ног в стороны в тренажёре сидя", "Ягодицы", nil, 10, false},
		{"Сведения ног друг к другу сидя в тренажёре", "Приводящие", nil, 10, false},
		{"Отведение ног назад в кроссовере", "Ягодицы", nil, 3, false},
		{"Отведения ног назад в тренажёре", "Ягодицы", String("Бицепс бедра"), 10, false},
		{"Подъёмы на носки стоя в тренажёре", "Икры", nil, 25, false},
		{"Подъёмы на носки сидя в тренажёре", "Икры", nil, 15, false},
		{"Подъёмы на носки стоя со штангой", "Икры", nil, 15, false},
		{"Подъёмы на носки стоя с гирей или гантелей на одной ноге", "Икры", nil, 3, false},
		{"Жим носками в тренажёре для жима ног", "Икры", nil, 40, false},
		{"Скручивания в тренажёре", "Пресс", nil, 10, false},
		{"Молитва", "Пресс", nil, 10, false},
	}

	sqlQuery := `INSERT INTO exercises (exercises_name, muscular_group, muscular_subgroup, working_weights, safe_for_injuries) VALUES `
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

func String(s string) *string {
	return &s
}
