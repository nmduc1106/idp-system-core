# OCR Worker Service

## Service Overview
The `ocr-worker` operates as an asynchronous background worker responsible for consuming document processing jobs from RabbitMQ, extracting structured data using **Anthropic Claude**, updating PostgreSQL, and explicitly triggering **Go API Webhooks** for cache invalidation.

## Tech Stack
- **Language:** Python 3.10
- **Core Libraries:**
  - `pika` — RabbitMQ consumer integration
  - `anthropic` — Anthropic Claude Multimodal Vision API SDK
  - `pydantic` — Structured output validation via JSON schemas
  - `tenacity` — Production-grade retry logic with exponential backoff
  - `minio` — S3-compatible Object Storage SDK
  - `sqlalchemy` & `psycopg2-binary` — PostgreSQL Database interactions
  - `structlog` — Structured JSON logging
  - `opentelemetry` — Distributed tracing (OTLP gRPC exporter)


## Asynchronous Processing Pipeline

```mermaid
sequenceDiagram
    participant RMQ as RabbitMQ
    participant Worker as OCR Worker
    participant DB as PostgreSQL
    participant Claude as Claude Vision API
    participant GoAPI as Go API Gateway

    RMQ->>Worker: Consume Message (job_id)
    Worker->>DB: update_job_status(PROCESSING)
    Worker->>Claude: messages.create(base64_image)
    Claude-->>Worker: JSON string
    Worker->>DB: update_job_status(COMPLETED, result)
    Worker->>GoAPI: POST /internal/webhook/job-completed (X-Webhook-Secret)
    GoAPI-->>Worker: 200 OK (Cache Cleared)
    Worker->>RMQ: basic_ack (Success)
```

## Multimodal LLM Pipeline (Claude Vision)
The extraction logic is encapsulated within the `OCRProcessor` singleton class in `processor.py`.

### Image Processing Flow
1. **Download**: Raw image bytes are fetched from MinIO via `get_object()`.
2. **MIME Validation**: The file extension is used to determine the MIME type. V1 supports `image/png`, `image/jpeg`, `image/webp`, and `image/gif`.
3. **Base64 Encoding**: Image bytes are encoded to a Base64 string using `base64.standard_b64encode()`.
4. **Claude Vision API**: The base64 image is sent to Claude's Messages API as an inline `source.type: "base64"` image block, alongside a text prompt requesting extraction.

### Structured Output via System Prompt
Since Claude does not have a native `response_schema` parameter, a detailed **system prompt** is used to enforce structured JSON output:
- The `ExtractedData` Pydantic model is serialized to JSON Schema via `model_json_schema()`.
- The schema is embedded directly in the system prompt with strict instructions:
  - Return **only** valid JSON, no markdown or conversational text.
  - Set missing fields to `null` — never invent data.
- The raw response text is stripped of any accidental ` ```json ` wrappers, parsed with `json.loads()`, and validated through `ExtractedData.model_validate()`.

### Pydantic Schema (`ExtractedData`)
| Field            | Type             | Description                                    |
|------------------|------------------|------------------------------------------------|
| `date`           | `Optional[str]`  | Invoice/document date (DD/MM/YYYY)             |
| `total_amount`   | `Optional[str]`  | Total amount with currency symbol              |
| `tax_id`         | `Optional[str]`  | Tax identification number (MST)                |
| `invoice_number` | `Optional[str]`  | Invoice/receipt number                         |
| `vendor_name`    | `Optional[str]`  | Vendor/seller/company name                     |

## Error Handling & Resiliency

### Tenacity `@retry` Decorator
The `_call_llm_vision` method is wrapped with `@retry(wait=wait_exponential(multiplier=1, min=2, max=10), stop=stop_after_attempt(5))`. This handles:
- **429 Rate Limit**: Exponential backoff (2s → 4s → 8s → 10s → 10s) before retrying.
- **529 API Overloaded / 5xx Server Errors**: Automatically retried with the same backoff.
- **`reraise=True`**: After 5 failed attempts, the original exception is re-raised to the caller.

### Connection Retries
The RabbitMQ connection logic (`connect()` in `worker.py`) uses an infinite `while True` loop to catch `AMQPConnectionError` and retry every 5 seconds.

### Resource Management
The `OCRProcessor` employs a **Singleton pattern** ensuring the Anthropic client and MinIO client are initialized exactly once, avoiding memory leaks. RabbitMQ QoS `prefetch_count=1` guarantees the worker processes one job at a time.

### Message Acknowledgment
- **`basic_ack`**: Sent only after the DB is updated with `COMPLETED` status and extracted data.
- **`basic_nack`**: On exception — the error is logged, DB status set to `FAILED`, and the message is negatively acknowledged with `requeue=False` to prevent poisoned-message loops.

## Configuration (`config.py`)
Environment variables are loaded with `load_dotenv(override=False)`, ensuring Docker Compose injected variables take absolute precedence over local `.env` files.

| Variable                      | Source Priority                          | Default Fallback                     |
|-------------------------------|------------------------------------------|--------------------------------------|
| `ANTHROPIC_API_KEY`           | Docker Secret → Env Var                  | —                                    |
| `ANTHROPIC_MODEL`             | Env Var                                  | `claude-sonnet-4-5-20250929`         |
| `RABBITMQ_URL`                | Env Var (`RABBITMQ_URL` → `AMQP_URL`)   | `amqp://guest:guest@localhost:5672/` |
| `DATABASE_URL`                | Env Var                                  | `postgresql://...@localhost:5432/...` |
| `MINIO_ENDPOINT`              | Env Var                                  | `localhost:9000`                     |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | Env Var (read in `worker.py`)           | `http://localhost:4317`              |

## Folder Map
```text
ocr-worker/src/
├── config.py       # Configuration manager. Loads env vars with Docker Compose priority (load_dotenv(override=False)).
│                   # Reads Docker Secrets for API keys via _read_secret() helper.
├── database.py     # SQLAlchemy ORM. Connection pooling, update_job_status(), get document paths.
├── processor.py    # The AI Engine. Singleton OCRProcessor: MinIO download, Base64 encoding,
│                   # Anthropic Claude Vision API call, Pydantic validation, Tenacity retry.
└── worker.py       # Main Application. Healthcheck server, OpenTelemetry tracing setup,
                    # RabbitMQ event loop (consume, ack/nack, orchestrate pipeline).
```
