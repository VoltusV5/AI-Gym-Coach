import asyncio
from aiogram import Bot, Dispatcher, types, F
from aiogram.filters import Command
from aiogram.types import InlineKeyboardMarkup, InlineKeyboardButton
from aiogram.fsm.context import FSMContext
from aiogram.fsm.state import State, StatesGroup
import os
from dotenv import load_dotenv
import gspread
import json

load_dotenv()
API_TOKEN = os.getenv("BOT_TOKEN")

bot = Bot(token=API_TOKEN)
dp = Dispatcher()

gc = gspread.service_account(filename='client_secret.json')
sh = gc.open_by_url(os.getenv("GOOGLE"))
worksheet = sh.sheet1

class EditState(StatesGroup):
    waiting_for_fix = State()


INDEX_FILE = 'progress.txt'

def save_index(index):
    """Сохраняет индекс в файл"""
    with open(INDEX_FILE, 'w') as f:
        f.write(str(index))
    print(f"💾 Индекс {index} сохранен")

def load_index():
    """Загружает индекс из файла"""
    if os.path.exists(INDEX_FILE):
        with open(INDEX_FILE, 'r') as f:
            index = int(f.read().strip())
            print(f"📂 Загружен индекс {index}")
            return index
    print("📂 Файл не найден, старт с 0")
    return 0





# Хранилище
current_index = load_index()
rows_data = []


def format_request(input_json):
    """Форматирует запрос в краткий вид"""
    try:
        if isinstance(input_json, str):
            data = json.loads(input_json)
        else:
            data = input_json
        
        # Берем только главное
        parts = []
        if 'пол' in data:
            parts.append(f"{data['пол']}")
        if 'возраст' in data:
            parts.append(f"{data['возраст']} лет")
        if 'цель' in data:
            parts.append(f"цель: {data['цель']}")
        if 'уровень_подготовки' in data:
            parts.append(f"ур: {data['уровень_подготовки']}")
        if 'дни_тренировок' in data:
            parts.append(f"дни: {len(data['дни_тренировок'])}")
        if "вес" in data:
            parts.append(f"вес: {data["вес"]}")
        if "рост" in data:
            parts.append(f"рост {data["рост"]}")
        if "тип_активности" in data:
            parts.append(f"тип активности: {data["тип_активности"]}")
        if "травмы_или_болезни" in data:
            parts.append(f"травмы_или_болезни {data["травмы_или_болезни"]}")
        
        
        
        return " | ".join(parts)
    
    except:
        # Если не спарсилось - просто обрезаем
        return str(input_json)[:100] + "..." if len(str(input_json)) > 100 else str(input_json)
    



def load_all_rows():
    """Загружает все строки из таблицы"""
    global rows_data
    rows_data = []
    all_rows = worksheet.get_all_values()
    
    for i, row in enumerate(all_rows[1:], start=2):
        if len(row) >= 2:
            rows_data.append({
                'row': i,
                'input': row[0],
                'output': row[1]
            })

def parse_workout_plan(output_json):
    """Парсит JSON и возвращает красивый текст плана тренировки"""
    try:
        if isinstance(output_json, str):
            data = json.loads(output_json)
        else:
            data = output_json
        
        text = "🏋️ ПЛАН ТРЕНИРОВКИ\n\n"
        
        for day, workouts in data['план_тренировок'].items():
            text += f"📅 День {day}\n"
            text += "═══════════════\n"
            
            for workout in workouts:
                text += f"\n🔹 {workout['группа']}\n"
                
                for ex in workout['упражнения']:
                    text += f"  • {ex['основное']}\n"
                    
                    if ex['вариации']:
                        variations = ", ".join(ex['вариации'])
                        text += f"    └ вариации: {variations}\n"
                
                text += f"\n    🔸 {workout['подходы']} x {workout['повторения']}"
                text += f" | отдых: {workout['отдых']}\n"
            
            text += "\n"
        
        if 'начальные_веса' in data and data['начальные_веса']:
            text += "⚖️ НАЧАЛЬНЫЕ ВЕСА\n"
            text += "═══════════════\n"
            for ex, weight in data['начальные_веса'].items():
                text += f"  • {ex}: {weight}\n"
        
        return text
    
    except Exception as e:
        return f"❌ Ошибка парсинга: {e}\n\n```json\n{output_json}\n```"

@dp.message(Command('start'))
async def cmd_start(message: types.Message):
    global current_index
    load_all_rows()
    
    if not rows_data:
        await message.answer("✅ Таблица пуста!")
        return
    
    await show_current(message)

async def show_current(message):
    global current_index, rows_data
    
    if current_index >= len(rows_data):
        await message.answer("✅ Все записи обработаны!")
        return
    
    data = rows_data[current_index]
    short_request = format_request(data['input'])
    # Формируем текст с красивым планом
    header = f"📝 СТРОКА {data['row']} ВСЕГО: 1501\n\n"
    header += f"📥 ЗАПРОС:\n{short_request}\n\n"
    header += f"📤 СГЕНЕРИРОВАННЫЙ ПЛАН:\n"
    
    plan_text = parse_workout_plan(data['output'])
    full_text = header + plan_text
    await message.answer(full_text)
    
    # Кнопки отдельно
    keyboard = InlineKeyboardMarkup(inline_keyboard=[
        [
            InlineKeyboardButton(text="✅ Одобрено", callback_data="approve"),
            InlineKeyboardButton(text="✏️ Предложить правку", callback_data="fix")
        ]
    ])
    await message.answer("Выбери действие:", reply_markup=keyboard)

@dp.callback_query(F.data == "approve")
async def approve(callback: types.CallbackQuery):
    global current_index
    current_index += 1
    save_index(current_index)
    await callback.message.delete()
    await show_current(callback.message)
    await callback.answer()

@dp.callback_query(F.data == "fix")
async def fix_plan(callback: types.CallbackQuery, state: FSMContext):
    await callback.message.edit_text(
        f"Строка {rows_data[current_index]['row']}\n\n"
        "Опиши текстом, что нужно исправить в плане тренировки:\n"
        "(Например: 'увеличить отдых до 3 минут', 'заменить жим на отжимания')"
    )
    await state.set_state(EditState.waiting_for_fix)
    await callback.answer()

@dp.message(EditState.waiting_for_fix)
async def receive_fix(message: types.Message, state: FSMContext):
    global current_index
    
    row = rows_data[current_index]['row']
    fix_text = message.text
    
    # Вместо worksheet.update(f'C{row}', fix_text)
    worksheet.update_cell(row, 3, fix_text)  # row, колонка (3 = C), значение
    
    await message.answer("✅ Правка сохранена в колонку C!")
    await state.clear()
    
    current_index += 1
    save_index(current_index)
    await show_current(message)

@dp.message(Command('next'))
async def cmd_next(message: types.Message):
    global current_index
    current_index += 1
    save_index(current_index)
    await show_current(message)

async def main():
    await dp.start_polling(bot)

if __name__ == '__main__':
    asyncio.run(main())