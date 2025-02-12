from google.cloud.storage import Client, transfer_manager
from typing import List, Dict
from gcp.cloud_config import _credentials
from transcript_service.config import settings
from transcript_service.logger import logger


class CloudStorage:
    def __init__(self, destination_directory: str):
        self.destination_directory = destination_directory
        self.credentials = _credentials
        self.project_id = settings.project_id
        self.bucket_name = settings.bucket_name

    def download_audio_files(self, blob_names: List[str]) -> Dict[str, str]:
        storage_client = Client(credentials=self.credentials, project=self.project_id)
        bucket = storage_client.bucket(self.bucket_name)
        results = transfer_manager.download_many_to_path(
            bucket=bucket,
            blob_names=blob_names,
            destination_directory=self.destination_directory,
            max_workers=8,
        )

        files = {}

        for name, result in zip(blob_names, results):
            if isinstance(result, Exception):
                logger.error(f"failed to download {name} due to exception: {result}")
            else:
                logger.info(f"downloaded {name} to {self.destination_directory}")
                files[self.destination_directory] = name

        return files
