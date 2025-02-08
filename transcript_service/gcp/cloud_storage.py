from google.cloud.storage import Client, transfer_manager
from typing import List, Dict
from gcp.cloud_config import _credentials
from transcript_service.config import settings
from transcript_service.logger import logger


def download_audio_files(
    blob_names: List[str],
    destination_directory: str,
) -> Dict[str, str]:
    storage_client = Client(credentials=_credentials, project=settings.project_id)
    bucket = storage_client.bucket(settings.bucket_name)
    results = transfer_manager.download_many_to_path(
        bucket=bucket,
        blob_names=blob_names,
        destination_directory=destination_directory,
        max_workers=8,
    )

    files = {}

    for name, result in zip(blob_names, results):
        if isinstance(result, Exception):
            logger.error(f"failed to download {name} due to exception: {result}")
        else:
            logger.info(f"downloaded {name} to {destination_directory}")
            files[destination_directory] = name

    return files
