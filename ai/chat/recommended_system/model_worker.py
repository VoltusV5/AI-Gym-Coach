"""
Второй этап: берёт список отфильтрованных данных на предыдущем этапе(sql_worker.py).
Используя легковестную модель отбирает 5-10 вариантов.
"""
from abc import ABC, abstractmethod
import logging

logger = logging.getLogger(__name__)

class BaseRecommendedModel(ABC):

    @abstractmethod
    def cross_encoder(self, plans:list)->list:
        pass


class CrossEncoderRanker(BaseRecommendedModel):
    __PATH_MODEL:str|None = None
    #__MODEL ?

    def __init__(self, path_model) -> None:
        self.__PATH_MODEL = path_model

    def cross_encoder(self, plans: list) -> list:
        """
        Метод возвращающий 5-10 финальных планов, которые пойдут в LLM
        """
        return []

    