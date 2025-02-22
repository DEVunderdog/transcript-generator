import signal

from config import settings
from logger import logger
from service import Service

if __name__ == "__main__":
    execute_service = Service(
        project_id=settings.project_id,
        bucket_name=settings.bucket_name,
        subscription_id=settings.subscription_id,
        sender_email=settings.sender_email,
        sender_password=settings.sender_password,
    )

    signal.signal(signal.SIGINT, execute_service.signal_handler)
    signal.signal(signal.SIGTERM, execute_service.signal_handler)

    try:
        execute_service.run_service()
    except Exception as e:
        logger.info(f"error occurred: {e}")
    finally:
        execute_service.cleanup()
