"""
Второй этап: берёт список отфильтрованных данных на предыдущем этапе(sql_worker.py).
Используя легковестную модель отбирает 5-10 вариантов.
"""
from abc import ABC, abstractmethod
from sentence_transformers import CrossEncoder
from ai.chat.recommended_system.sql_worker import SqlFillterPlans
from sqlalchemy import create_engine
import pandas as pd
import os
import time
import torch
from ..logger_config import logger


class BaseRecommendedModel(ABC):
    """
    Базовый интерфейс для моделей семантического ранжирования тренировочных планов.
    """
    @abstractmethod
    def cross_encoder(self, plans:pd.DataFrame, query)->pd.DataFrame:
        """
        Выполняет ранжирование переданных планов относительно поискового запроса.

        Args:
            plans (pd.DataFrame): Таблица с кандидатами планов (должна содержать поле с описанием).
            query (str): Текстовый запрос пользователя/промпт для сопоставления.

        Returns:
            pd.DataFrame: Таблица с отранжированными планами и оценками релевантности.
        """
        pass


class CrossEncoderRanker(BaseRecommendedModel):
    """
    Класс для реранкинга планов с использованием архитектуры Cross-Encoder.

    Вычисляет семантическую релевантность пар (запрос, описание_плана) 
    и отбирает наиболее подходящие кандидаты для контекста LLM. 
    Модель инициализируется один раз при создании экземпляра.

    Attributes:
        __PATH_MODEL (str): Директория для локального кэширования весов модели.
        __MODEL_NAME (str): Название или HuggingFace-идентификатор используемой модели.
        model (CrossEncoder | None): Загруженный экземпляр модели или None при ошибке инициализации.
    """

    __PATH_MODEL:str = ""
    __MODEL_NAME = "BAAI/bge-reranker-v2-m3"

    def __init__(self, path_model="ai/chat/recommended_system/model/", model_name="BAAI/bge-reranker-v2-m3") -> None:
        """
        Инициализирует и загружает Cross-Encoder модель в память.

        Определяет доступное устройство (CUDA или CPU), загружает веса модели 
        в указанный кэш и переводит модель в состояние готовности к инференсу.

        Args:
            path_model (str, optional): Путь к директории для сохранения/загрузки кэша модели.
                По умолчанию "ai/chat/recommended_system/model/".
            model_name (str, optional): Имя предобученной модели реранкера в HuggingFace Hub.
                По умолчанию "BAAI/bge-reranker-v2-m3".
        """

        self.__PATH_MODEL = path_model
        self.__MODEL_NAME = model_name
        self.model = None
        try:
            device = "cuda" if torch.cuda.is_available() else "cpu"
            self.model = CrossEncoder(
                    model_name_or_path=model_name, 
                    max_length=256,
                    device=device,
                    model_kwargs={"cache_dir": path_model},
                    local_files_only= os.path.isdir(path_model) and bool(os.listdir(path_model))
                    )
            logger.info("Модель %s инициализирована используется %s", self.__MODEL_NAME, device)
        except Exception:
            logger.exception("Ошибка при инициализации модели")


    def cross_encoder(self, plans: pd.DataFrame, query) -> pd.DataFrame:
        """
        Ранжирует планы по релевантности текстовому запросу и возвращает топ-10 кандидатов.

        Формирует пары [запрос, описание], выполняет инференс на предварительно 
        загруженной модели, рассчитывает нормализованные скоры (0–100%) через Sigmoid 
        и отбирает до 10 наиболее релевантных планов.

        Args:
            plans (pd.DataFrame): DataFrame с планами. Обязательно наличие колонки 'description'.
            query (str): Текстовый запрос или критерии пользователя для сопоставления.

        Returns:
            pd.DataFrame: DataFrame из максимум 10 лучших планов с добавленной колонкой 'Score',
            отсортированных по убыванию релевантности. При ошибке или неинициализированной 
            модели возвращает пустой pd.DataFrame.
        """

        t1 = time.perf_counter()
        try:
            
            data = [[query, text] for text in plans["description"]]
            if self.model is None:
                logger.error("Модель не инициализировалась")
                return pd.DataFrame()
            
            pred = self.model.predict(data, activation_fn=torch.nn.Sigmoid())*100
            plans["Score"] = pred
            plans = plans.sort_values(by="Score", ascending=False)

        except Exception:
            logger.exception("Ошибка в ранжировании моделью")
            return pd.DataFrame()

        t2 = time.perf_counter()
        logger.info("Успешное ранжирование моделью %s за %0.3f сек", self.__MODEL_NAME, t2-t1)
        return plans.nlargest(10, columns="Score")


    def get_path_model(self) -> str:
        """
        Возвращает путь к директории кэша весов модели.

        Returns:
            str: Локальный путь к папке с кэшем модели.
        """
        return self.__PATH_MODEL






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

    cross = CrossEncoderRanker(model_name="cross-encoder/mmarco-mMiniLMv2-L12-H384-v1")
    if t1 is not None:
        print(t1)
        print(cross.cross_encoder(t1,"Хочу ноги накачать"))


