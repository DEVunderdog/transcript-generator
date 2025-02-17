from service import Service
from config import settings

if __name__ == "__main__":
    execute_service = Service(project_id=settings.project_id, bucket_name=settings.bucket_name, subscription_id=settings.subscription_id)
    execute_service.run_service()