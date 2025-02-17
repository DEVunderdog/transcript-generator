from google.cloud.storage import Client, transfer_manager
from typing import List
from gcp.cloud_config import _credentials
from logger import logger
from constants import constants


class CloudStorage:
    def __init__(self, project_id: str, bucket_name: str):
        self.destination_directory = constants.download_file_path
        self.project_id = project_id
        self.bucket_name = bucket_name


    def download_audio_files(self, blob_names: List[str]) -> tuple[str, str]:
        storage_client = Client(credentials=_credentials, project=self.project_id)
        bucket = storage_client.bucket(self.bucket_name)
        results = transfer_manager.download_many_to_path(
            bucket=bucket,
            blob_names=blob_names,
            destination_directory=self.destination_directory,
            max_workers=8,
        )

        file = None
        file_name = None

        for name, result in zip(blob_names, results):
            if isinstance(result, Exception):
                logger.error(f"failed to download {name} due to exception: {result}")
            else:
                logger.info(f"downloaded {name} to {self.destination_directory}")
                file = self.destination_directory + "/" + name
                file_name = name.split("/", 1)[1]

        return file, file_name
