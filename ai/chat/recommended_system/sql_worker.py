"""
Первый этап: достаёт данные из бд и фильтрует их при помощи фильтров.
TODO:
2. Сделать сортировку по рейтингу - когда он появиться:)
"""
from abc import ABC, abstractmethod
from sqlalchemy.engine import Engine
from sqlalchemy import text
from sqlalchemy import create_engine, inspect
from ..logger_config import logger
import pandas as pd
import time


class BaseRecommendedSQL(ABC):
    """
    Базовый интерфейс для классов фильтрации и подбора программ через SQL.
    """
    @abstractmethod
    def filltering_sql(self) -> pd.DataFrame|None:
        """
        Выполняет пайплайн фильтрации тренировочных программ.

        Returns:
            pd.DataFrame | None: Таблица с подходящими программами 
            (колонки: program_id, description) либо None в случае ошибки/пустого результата.
        """
        pass


class SqlFillterPlans(BaseRecommendedSQL):
    """
    Класс для загрузки программ из БД и их последующей фильтрации
    на основе параметров анкеты пользователя.

    Attributes:
        __FORM_USER (dict): Данные анкеты пользователя с целевыми параметрами.
        __ENGINE (Engine): SQLAlchemy engine для подключения к базе данных.
    """

    __FORM_USER: dict = {}
    __ENGINE = 0

    def __init__(self, form_user, ENGINE):
        """
        Инициализирует экземпляр фильтра программ.

        Args:
            form_user (dict): Словарь с параметрами пользователя. Ожидаемые ключи:
                - "goal": цель тренировок.
                - "fitness_level": уровень физической подготовки.
                - "training_days_map": коллекция/маппинг тренировочных дней.
            ENGINE: SQLAlchemy Engine для выполнения запросов к БД.
        """
        self.__FORM_USER = form_user
        self.__ENGINE = ENGINE


    def filltering_sql(self) -> pd.DataFrame|None:
        """
        Выполняет полный цикл отбора программ: SQL-запрос с последующей фильтрацией в Pandas.

        Сначала выгружает опубликованные версии активных программ из БД (до 50 записей),
        затем отсекает неподходящие под анкету пользователя варианты и возвращает
        только итоговые поля.

        Returns:
            pd.DataFrame | None: DataFrame с колонками ['program_id', 'description'],
            либо None, если на одном из этапов произошла ошибка или данные отсутствуют.
        """

        t1 = time.perf_counter()
        filter_table_df = self.__filter_sql_query_to_pd()
        if filter_table_df is None:
            logger.warning("Пустая таблица пришла из sql")
            return None
        
        filter_table_df = self.__filter_with_user_form(filter_table_df)
        if filter_table_df is None:
            return None
        t2 = time.perf_counter()
        logger.info("Фильтрация 1 этап SQL завершилась %0.3f сек", t2 - t1)
        return filter_table_df[["program_id", "description"]]


    def __filter_sql_query_to_pd(self):
        """
        Выполняет SQL-запрос к БД и загружает результаты в DataFrame.

        Выбирает опубликованные версии активных программ со связкой параметров:
        цель, уровень и количество тренировок в неделю (ограничение 50 записей).

        Returns:
            pd.DataFrame | None: DataFrame с сырыми данными из БД 
            (program_id, description, goal, level, sessions_per_week) 
            либо None при сбое подключения/выполнения запроса.
        """

        t1 = time.perf_counter()
        filtered_table = None
        try:
            with self.__ENGINE.connect() as conn:
                query = text("""
                            SELECT pr_ver.program_id, pr_ver.description, pr_ver.goal, pr_ver.level, pr_ver.sessions_per_week
                            FROM programs pr
                            LEFT JOIN program_versions pr_ver ON pr_ver.program_id = pr.id 
                            AND pr.status = "active" 
                            AND pr_ver.status = "published"
                            LIMIT 50;
                        """)
                filtered_table = pd.read_sql_query(query,conn)
        except Exception:
            filtered_table = None
            logger.exception("Ошибка фильтрации таблиц в sql!")

        t2 = time.perf_counter()
        logger.info("Фильтрация таблиц в sql %0.3f сек - завершена", t2-t1)
        return filtered_table



    def __filter_with_user_form(self, filtered_table):
        """
        Фильтрует DataFrame по параметрам из пользовательской анкеты.

        Сопоставляет колонки DataFrame со значениями анкеты:
        - `goal` == `form['goal']`
        - `level` == `form['fitness_level']`
        - `sessions_per_week` == `len(form['training_days_map'])`

        Args:
            filtered_table (pd.DataFrame): Исходный DataFrame, полученный из БД.

        Returns:
            pd.DataFrame | None: Отфильтрованный DataFrame либо None, 
            если в процессе фильтрации возникла ошибка (например, отсутствует нужный ключ в анкете).
        """
        
        form = self.__FORM_USER
        t1 = time.perf_counter()
        try:
            filtered_table = filtered_table[(filtered_table["goal"] == form["goal"]) &
                                            (filtered_table["level"] == form["fitness_level"]) &
                                            (filtered_table["sessions_per_week"] == len(form["training_days_map"]))]
            
        except:
            filtered_table = None
            logger.exception("Ошибка фильтрации на уровне pandas")
        t2 = time.perf_counter()
        logger.info("Фильтрация в pandas завершилась %0.3f сек",t2-t1)
        return filtered_table


if __name__ == "__main__":
    DATABASE_URL = "sqlite:///marketplace.db"
    engine = create_engine(DATABASE_URL)
    form_user = {
    "version": 1,
    "age": 26,
    "height_cm": 175,
    "weight_kg": 75,
    "goal": "Набрать мышцы",
    "fitness_level": "новичок",
    "training_days_map": [
        "Пн",
        "Ср",
        "Пт"
    ]
    }
    filtery = SqlFillterPlans(form_user, engine)
    t1= filtery.filltering_sql()
    print(t1)

    