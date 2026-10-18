"""
Объединяет первый и второй этап в себе, на выходе получаем 5-10 планов,
которые летят в LLM.
"""
from chat.recommended_system.sql_worker import SqlFillterPlans
from chat.recommended_system.model_worker import CrossEncoderRanker
import logging

logger = logging.getLogger(__name__)


class RecommendationPipline:
    __COUNT_SQL: int | None = None
    __COUNT_MODEL:int | None = None

    def __init__(self, ranker:CrossEncoderRanker, fillter:SqlFillterPlans, count_sql:int=30, count_model:int=10) -> None:
        pass