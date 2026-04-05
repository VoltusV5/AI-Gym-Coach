/**
 * Ответ POST /api/v1/plans/generate для моков и превью (см. docs/TZ-backend-trenirovka-uprazhneniya.md).
 * plan[].exercises[j] — слот j, внутри массив вариаций { id, exercise_name, weight }.
 */
export function getMockPlanGenerateResponse() {
  return {
    split: 'fullbody',
    plan: [
      {
        day: 'A',
        day_name: 'Фулбоди А',
        exercises: [
          [
            { id: 142, exercise_name: 'Жим штанги лёжа', weight: 42 },
            { id: 143, exercise_name: 'Жим в тренажёре (грудь)', weight: 42 },
            { id: 144, exercise_name: 'Жим гантелей лёжа', weight: 42 }
          ],
          [
            { id: 89, exercise_name: 'Жим гантелей на наклонной', weight: 40 },
            { id: 90, exercise_name: 'Жим в Смите на наклонной', weight: 40 }
          ],
          [
            { id: 215, exercise_name: 'Тяга верхнего блока к груди', weight: 35 },
            { id: 216, exercise_name: 'Подтягивания с резинкой', weight: 35 },
            { id: 217, exercise_name: 'Тяга вертикального блока узким', weight: 35 }
          ],
          [
            { id: 301, exercise_name: 'Тяга Т-грифа', weight: 32 },
            { id: 302, exercise_name: 'Тяга горизонтального блока', weight: 32 }
          ],
          [
            { id: 401, exercise_name: 'Жим ногами в тренажёре', weight: 100 },
            { id: 402, exercise_name: 'Приседания в Смите', weight: 60 }
          ],
          [
            { id: 501, exercise_name: 'Ягодичный мост со штангой', weight: 50 },
            { id: 502, exercise_name: 'Отведения ноги в кроссовере', weight: 25 }
          ],
          [
            { id: 601, exercise_name: 'Скручивания на пресс', weight: 0 },
            { id: 602, exercise_name: 'Планка', weight: 0 }
          ]
        ]
      },
      {
        day: 'B',
        day_name: 'Фулбоди B',
        exercises: [
          [{ id: 701, exercise_name: 'Жим штанги сидя', weight: 30 }],
          [{ id: 702, exercise_name: 'Махи гантелей в стороны', weight: 8 }],
          [{ id: 703, exercise_name: 'Разведения в наклоне', weight: 12 }],
          [
            { id: 801, exercise_name: 'Румынская тяга', weight: 55 },
            { id: 802, exercise_name: 'Сгибания ног лёжа', weight: 40 }
          ],
          [
            { id: 901, exercise_name: 'Сгибания рук со штангой', weight: 20 },
            { id: 902, exercise_name: 'Молотковые сгибания', weight: 14 }
          ],
          [{ id: 601, exercise_name: 'Скручивания на пресс', weight: 0 }]
        ]
      },
      {
        day: 'C',
        day_name: 'Фулбоди C',
        exercises: [
          [
            { id: 145, exercise_name: 'Разводка гантелей лёжа', weight: 14 },
            { id: 146, exercise_name: 'Отжимания на брусьях', weight: 0 }
          ],
          [{ id: 218, exercise_name: 'Тяга нижнего блока сидя', weight: 40 }],
          [{ id: 403, exercise_name: 'Выпады с гантелями', weight: 16 }],
          [{ id: 1001, exercise_name: 'Подъёмы на носки стоя', weight: 60 }],
          [{ id: 603, exercise_name: 'Подъёмы ног в висе', weight: 0 }]
        ]
      }
    ]
  }
}
