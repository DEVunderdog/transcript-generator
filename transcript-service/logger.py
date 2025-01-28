import logging
import os
from logging.handlers import TimedRotatingFileHandler

def setup_logger():
    current_dir = os.path.dirname(os.path.abspath(__file__))

    log_dir = os.path.join(os.path.dirname(current_dir), "logs", "transcript-service")
    if not os.path.exists(log_dir):
        os.makedirs(log_dir)

    logger = logging.getLogger()
    logger.setLevel(logging.INFO)

    formatter = logging.Formatter("%(asctime)s - %(levelname)s - %(message)s")

    console_handler = logging.StreamHandler()
    console_handler.setFormatter(formatter)
    logger.addHandler(console_handler)

    file_handler = TimedRotatingFileHandler(
        filename=os.path.join(log_dir, "transcript-service.log"),
        when="midnight",
        interval=1,
        backupCount=30,
        encoding="utf-8"
    )
    file_handler.setFormatter(formatter)
    file_handler.suffix = "%Y-%m-%d"
    logger.addHandler(file_handler)

    return logger

logger = setup_logger()