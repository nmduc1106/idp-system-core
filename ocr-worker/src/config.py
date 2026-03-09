import os
from typing import Optional
from dotenv import load_dotenv

# Load file .env but do NOT override existing environment variables.
# This ensures Docker Compose injected variables take highest precedence.
load_dotenv(override=False)


def _read_secret(secret_name: str) -> Optional[str]:
    """Read a Docker secret from /run/secrets/ (mounted read-only by Docker)."""
    secret_path = f"/run/secrets/{secret_name}"
    try:
        with open(secret_path, "r") as f:
            return f.read().strip()
    except FileNotFoundError:
        return None


class Config:
    # --- Anthropic Claude ---
    # Priority: Docker Secret > Environment Variable > .env file
    ANTHROPIC_API_KEY = _read_secret("anthropic_api_key") or os.getenv("ANTHROPIC_API_KEY")
    ANTHROPIC_MODEL = os.getenv("ANTHROPIC_MODEL", "claude-sonnet-4-5-20250929")

    # --- RabbitMQ ---
    # docker-compose.yml uses RABBITMQ_URL, falling back to AMQP_URL for legacy support
    RABBITMQ_URL = os.getenv("RABBITMQ_URL", os.getenv("AMQP_URL", "amqp://guest:guest@localhost:5672/"))
    QUEUE_NAME = "ocr_queue"
    DLQ_NAME = "ocr_dlq"

    # --- Database (SQLAlchemy URL) ---
    DATABASE_URL = os.getenv("DATABASE_URL", "postgresql://idp_user:secret_password@localhost:5432/idp_db")

    # --- Redis ---
    REDIS_URL = os.getenv("REDIS_URL", "redis://localhost:6379/0")

    # --- MinIO ---
    MINIO_ENDPOINT = os.getenv("MINIO_ENDPOINT", "localhost:9000")

    # Ensure Docker Compose variable names match Python env lookups
    MINIO_ACCESS_KEY = os.getenv("MINIO_ACCESS_KEY", os.getenv("MINIO_ROOT_USER", "minio_admin"))
    MINIO_SECRET_KEY = os.getenv("MINIO_SECRET_KEY", os.getenv("MINIO_ROOT_PASSWORD", "minio_secret_key"))
    MINIO_BUCKET = os.getenv("MINIO_BUCKET", "documents")
    MINIO_SECURE = os.getenv("MINIO_SECURE", os.getenv("MINIO_USE_SSL", "false")).lower() == "true"