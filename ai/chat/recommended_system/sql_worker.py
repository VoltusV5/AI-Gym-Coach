"""
Первый этап: достаёт данные из бд и фильтрует их при помощи фильтров.
"""
from abc import ABC, abstractmethod
from sqlalchemy.engine import Engine 

class BaseRecommendedSQL(ABC):

    @abstractmethod
    def filltering_sql(self, query) -> list:
        pass


class SqlFillterPlans(BaseRecommendedSQL):
    __DATABASEURL:str|None = None
    __ENGINE:Engine|None = None

    def __init__(self, DATABASEURL, ENGINE) -> None:
        self.__DATABASEURL = DATABASEURL
        self.__ENGINE = ENGINE

    def filltering_sql(self, query) -> list:
        """Метод в котором происходит отсечение планов по фильтрам"""
        return []
