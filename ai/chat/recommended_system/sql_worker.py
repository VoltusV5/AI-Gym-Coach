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

    @abstractmethod
    def filltering_sql(self) -> pd.DataFrame|None:
        pass


class SqlFillterPlans(BaseRecommendedSQL):
    __FORM_USER: dict = {}
    __ENGINE = 0

    def __init__(self, form_user, ENGINE) -> None:
        self.__FORM_USER = form_user
        self.__ENGINE = ENGINE


    def filltering_sql(self) -> pd.DataFrame|None:
        """Метод в котором происходит отсечение планов по фильтрам"""
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
        """Функция, которая фильтрует данные в sql"""
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
        """Фильтрация получившейся таблицы под конкретного user"""
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

    