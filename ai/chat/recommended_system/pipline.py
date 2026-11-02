"""
Объединяет первый и второй этап в себе, на выходе получаем 5-10 планов,
которые летят в LLM.
"""
from ai.chat.recommended_system.sql_worker import SqlFillterPlans
from ai.chat.recommended_system.model_worker import CrossEncoderRanker
import pandas as pd
from ..logger_config import logger


class RecommendationPipline:

    def __init__(self, ranker:CrossEncoderRanker, fillter:SqlFillterPlans) -> None:
        self.ranker = ranker
        self.fillter = fillter

    def pipline(self, query) -> pd.DataFrame|None:
        table = self.fillter.filltering_sql()
        if table is None:
            logger.error("Вернулась пустая таблица после фильтрации на уровне SQL")
            return None
        table = self.ranker.cross_encoder(table, query)
        if table.empty:
            logger.error("Вурнулась пустая таблица после работы модели")
            return None
        return table