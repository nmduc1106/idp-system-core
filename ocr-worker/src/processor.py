"""
OCR Processor - Anthropic Claude Multimodal LLM Pipeline.

Uses Claude's Vision API to extract structured data from document images.
V1: Supports images (PNG/JPEG/WEBP/GIF) only.
"""
import base64
import json
import mimetypes
from typing import Optional

import anthropic
import structlog
from minio import Minio
from pydantic import BaseModel, Field
from tenacity import retry, stop_after_attempt, wait_exponential

from src.config import Config

logger = structlog.get_logger()


# --- Pydantic Schema for Structured Output ---
class ExtractedData(BaseModel):
    """Schema for structured data extracted from invoice/receipt documents."""
    date: Optional[str] = Field(
        description="The invoice or document date (Ngày). Format: DD/MM/YYYY or as printed."
    )
    total_amount: Optional[str] = Field(
        description="The total amount or sum (Tổng tiền / Tổng cộng / Total Amount) including currency symbol if present."
    )
    tax_id: Optional[str] = Field(
        description="The tax identification number (Mã số thuế / MST / Tax ID)."
    )
    invoice_number: Optional[str] = Field(
        description="The invoice or receipt number (Số hóa đơn / Invoice No.)."
    )
    vendor_name: Optional[str] = Field(
        description="The name of the vendor/seller/company issuing the document."
    )


# Generate the JSON schema string once for the system prompt
_SCHEMA_JSON = json.dumps(ExtractedData.model_json_schema(), indent=2)

_SYSTEM_PROMPT = f"""You are an expert document extraction AI specializing in Vietnamese and English invoices, receipts, and business documents.

Your task: Analyze the provided image and extract structured data.

RULES:
1. Return ONLY valid JSON matching the schema below. No markdown, no explanation, no conversational text.
2. If a field is not present in the image, you MUST set its value to null.
3. Do not invent or guess any values.

JSON Schema:
{_SCHEMA_JSON}

Example output:
{{"date": "05/03/2026", "total_amount": "1,500,000 VND", "tax_id": "0123456789", "invoice_number": "INV-001", "vendor_name": "Công ty ABC"}}"""


class OCRProcessor:
    """
    Production-grade document processor using Anthropic Claude.

    Singleton pattern ensures the Claude client and MinIO connection
    are initialized exactly once across the worker's lifetime.
    """
    _instance = None

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super(OCRProcessor, cls).__new__(cls)
            cls._instance._initialize()
        return cls._instance

    def _initialize(self) -> None:
        """Initialize MinIO client and Anthropic client (called once)."""
        logger.info("initializing_processor_resources")

        # 1. Setup MinIO Client
        self.minio_client = Minio(
            Config.MINIO_ENDPOINT,
            access_key=Config.MINIO_ACCESS_KEY,
            secret_key=Config.MINIO_SECRET_KEY,
            secure=Config.MINIO_SECURE
        )

        # 2. Setup Anthropic Client
        if not Config.ANTHROPIC_API_KEY:
            logger.error("anthropic_api_key_missing")
            raise ValueError("ANTHROPIC_API_KEY environment variable is not set.")

        self.client = anthropic.Anthropic(api_key=Config.ANTHROPIC_API_KEY)
        self.model = Config.ANTHROPIC_MODEL
        logger.info("anthropic_client_initialized", model=self.model)

        # 3. Test MinIO connection
        try:
            if not self.minio_client.bucket_exists(Config.MINIO_BUCKET):
                logger.warning("bucket_not_found", bucket=Config.MINIO_BUCKET)
        except Exception as e:
            logger.error("minio_connection_error", error=str(e))

    @retry(
        wait=wait_exponential(multiplier=1, min=2, max=10),
        stop=stop_after_attempt(5),
        reraise=True
    )
    def _call_llm_vision(self, image_bytes: bytes, mime_type: str) -> ExtractedData:
        """
        Send an image to Claude and extract structured data.

        Decorated with @retry for exponential backoff against 429/529 errors.
        Uses system prompt to enforce JSON-only output matching the Pydantic schema.
        """
        logger.info("calling_claude_vision", mime_type=mime_type, model=self.model)

        # Encode image as base64 for Claude Messages API
        image_b64 = base64.standard_b64encode(image_bytes).decode("utf-8")

        message = self.client.messages.create(
            model=self.model,
            max_tokens=1024,
            system=_SYSTEM_PROMPT,
            messages=[
                {
                    "role": "user",
                    "content": [
                        {
                            "type": "image",
                            "source": {
                                "type": "base64",
                                "media_type": mime_type,
                                "data": image_b64,
                            },
                        },
                        {
                            "type": "text",
                            "text": "Extract all structured data from this document image.",
                        },
                    ],
                }
            ],
        )

        # Extract text from Claude's response
        raw_text = message.content[0].text

        # Strip any accidental markdown wrappers (```json ... ```)
        clean_text = raw_text.strip().removeprefix("```json").removesuffix("```").strip()
        parsed = json.loads(clean_text)
        return ExtractedData.model_validate(parsed)

    def process(self, object_name: str) -> str:
        """
        Production pipeline: Download -> Claude Vision -> Pydantic Validation -> JSON.

        V1: Supports image files (PNG, JPEG, WEBP, GIF) only.
        """
        try:
            # Step 1: Download from MinIO
            logger.info("downloading_file", object_name=object_name)
            response = self.minio_client.get_object(Config.MINIO_BUCKET, object_name)
            file_bytes = response.read()
            response.close()
            response.release_conn()

            # Step 2: Determine MIME type and validate format
            mime_type, _ = mimetypes.guess_type(object_name)
            supported_types = {"image/png", "image/jpeg", "image/webp", "image/gif"}

            if mime_type not in supported_types:
                error_msg = f"Unsupported file type: {mime_type}. V1 supports images only (PNG/JPEG/WEBP/GIF)."
                logger.warning("unsupported_file_type", object_name=object_name, mime_type=mime_type)
                raise ValueError(error_msg)

            # Step 3: Call Claude Vision
            logger.info("processing_image_with_claude", object_name=object_name)
            extracted_data = self._call_llm_vision(file_bytes, mime_type)

            # Step 4: Build final result (Pydantic -> dict -> JSON)
            result = {
                "extracted_data": extracted_data.model_dump()
            }
            logger.info("claude_processing_complete", object_name=object_name)
            return json.dumps(result, ensure_ascii=False)

        except Exception as e:
            logger.error("processing_error", error=str(e), object_name=object_name)
            raise e