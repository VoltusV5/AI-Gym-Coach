import logging
from logging.handlers import RotatingFileHandler

logger = logging.getLogger("chat")
logger.setLevel(logging.INFO)


logger.propagate = False

formatter = logging.Formatter(
    "%(asctime)s | %(name)s | [ %(levelname)s ] | %(message)s"
)

console = logging.StreamHandler()
console.setFormatter(formatter)


file_logger = RotatingFileHandler(
    filename="ai/chat/logs/chat.log",
    maxBytes=1_000_000,
    backupCount=3,
    encoding="utf-8"
)

file_logger.setFormatter(formatter)

error_logger = RotatingFileHandler(
    filename="ai/chat/logs/error.log",
    maxBytes=1_000_000,
    backupCount=2,
    encoding="utf-8"
)
error_logger.setLevel(logging.ERROR)
error_logger.setFormatter(formatter)

logger.addHandler(console)
logger.addHandler(error_logger)
logger.addHandler(file_logger)

