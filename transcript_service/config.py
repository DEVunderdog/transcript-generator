from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(case_sensitive=False)

    bucket_name: str
    project_id: str
    subscription_id: str
    sender_email: str
    sender_password: str


settings = Settings()
