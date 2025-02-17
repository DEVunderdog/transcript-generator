from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", case_sensitive=False)

    service_account_key_path: str
    bucket_name: str
    project_id: str
    subscription_id: str
    sender_email: str
    sender_password: str


settings = Settings(_env_file="../.env/transcript_service.env")
