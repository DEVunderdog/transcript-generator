from google.oauth2 import service_account
from transcript_service.config import settings

_credentials = service_account.Credentials.from_service_account_file(
    settings.service_account_key_path
)


