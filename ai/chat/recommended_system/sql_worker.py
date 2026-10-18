"""
Первый этап: достаёт данные из бд и фильтрует их при помощи фильтров.
"""
from abc import ABC, abstractmethod
from sqlalchemy.engine import Engine 
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
    __ENGINE: Engine|None = None
    __TABLES: list[str] = ['program_days', 'program_versions', 'programs']

    def __init__(self, form_user, ENGINE) -> None:
        self.__FORM_USER = form_user
        self.__ENGINE = ENGINE


    def filltering_sql(self) -> pd.DataFrame|None:
        """Метод в котором происходит отсечение планов по фильтрам"""
        t1 = time.perf_counter()
        if self.__ENGINE is not None:
            reader = self.__read_sql_to_pandas(self.__ENGINE)
            if reader is None:
                return None
            else:
                programs, program_versions, program_days = reader
        else:
            logger.error("Проблема с движком __ENGINE в SqlFillterPlans")
            return None
        
        try:
            programs, program_versions = self.__filtering_all_tables(programs, program_versions)
        except Exception:
            logger.exception("Ошибка на первом этапе sql-фильтрация")
            return None

        try:
            res = self.__create_final_df(programs, program_versions, program_days)
        except Exception:
            logger.exception("Ошибка на первом этапе sql-фильтрация")
            return None
        
        t2 = time.perf_counter()
        logger.info("Успешная фильтрация бд %.3f сек", t2-t1)

        if res is None:
            logger.error("Вернулся None в результате первого этапа!")
            return None
        return res
        

    def __create_final_df(self, programs, program_versions,program_days):
        programs_id = programs[["id", "published_version_id"]]
        days = program_days[program_days["program_version_id"].isin(programs_id["published_version_id"])]
        days = (
            days.groupby("program_version_id")["title"]
            .apply(lambda titles: ", ".join(titles.dropna()))
            .reset_index()
        )
        program_description = program_versions[program_versions["id"].isin(programs_id["published_version_id"])][["id", "program_id","description"]]

        try:

            res = pd.merge(program_description, days, how="left",left_on="id", right_on="program_version_id")
            res["description"] = res["description"].fillna("")
            res["title"] = res["title"].fillna(" Нет данных о названии дня.")
            res["full_info"] = res["description"] + res["title"] 
            return res
        except Exception:
                logger.error("Ошибка при left join")
                return None
    
                
    
    def __filtering_all_tables(self,programs, program_versions):
        """
        Фильтрация каждой таблицы
        TODO: Провекрки на пустоты после фильтрации
        """
        if not programs.empty:
            programs = programs[(programs["status"] == "active")]
            logger.info("Успешная фильтрация таблицы programs %d штук", len(programs))
        else:
            logger.warning("Таблица programs пустая")
        
        if not program_versions.empty:
            program_versions = program_versions[(program_versions["status"] == "published") & 
                                               (program_versions["goal"] == self.__FORM_USER["goal"]) &
                                               (program_versions["level"] == self.__FORM_USER["fitness_level"]) &
                                               (program_versions["sessions_per_week"] == len(self.__FORM_USER["training_days_map"]))]

            logger.info("Успешная фильтрация таблицы program_versions %d штук", len(program_versions))
        else:
            logger.warning("Таблица program_versions пустая")

        return programs, program_versions

    def __read_sql_to_pandas(self, engine:Engine):
        programs = pd.DataFrame()
        program_versions = pd.DataFrame()
        program_days = pd.DataFrame()
        for table_name in self.__TABLES:
            if table_name == "programs":
                programs = pd.read_sql_table(table_name=table_name, con=engine)
                if programs.empty:
                    logger.error("Таблица programs пустая, при считывании из БД")
                    programs = None
            elif table_name == "program_versions":
                program_versions = pd.read_sql_table(table_name=table_name, con=engine)
                if program_versions.empty:
                    program_versions = None
                    logger.warning("Таблица program_versions пустая, при считывании из БД")
            elif table_name == "program_days":
                program_days = pd.read_sql_table(table_name=table_name, con=engine)
                if program_days.empty:
                    program_days = None
                    logger.warning("Таблица program_days пустая, при считывании из БД")
        if programs is None:
            logger.error("Пустая таблица маркетплейса")
            return None
        if program_days is not None and program_versions is not None and programs is not None:
            logger.info("Успешное считываение из бд в pandas programs: %d стр, program_versions %d стр, program_days %d стр", len(programs), len(program_versions), len(program_days))
        return programs, program_versions, program_days


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

    