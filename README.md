# Intelligent Document Processing (IDP) System 🚀

An enterprise-grade, microservices-based Document Processing System that leverages Multimodal LLMs (Anthropic Claude 3.5) to automatically extract structured JSON data from raw images and documents.

## 🏗️ System Architecture

The system is built with a highly resilient, asynchronous event-driven architecture:

1. **API Gateway (Golang):** Provides high-performance REST APIs for file uploads and status polling. Includes self-healing RabbitMQ connection management.
2. **Message Broker (RabbitMQ):** Decouples ingestion from processing, ensuring no jobs are lost during traffic spikes.
3. **OCR AI Worker (Python):** Consumes messages, fetches files from MinIO, encodes them to Base64, and uses **Anthropic Claude Vision API** with strict Pydantic schemas for zero-hallucination data extraction.
4. **Storage & DB:** MinIO (S3-compatible) for object storage and PostgreSQL for job state management.

## 📂 Project Structure

- `/api-gateway` - Golang REST API & Producer.
- `/ocr-worker` - Python Worker & AI Engine (See `ocr-worker/README_SERVICE.md` for deep dive).
- `docker-compose.yml` - Local infrastructure orchestration.

## 🛠️ CI/CD
This repository includes a GitHub Actions pipeline (`.github/workflows/ci.yml`) that automatically runs parallel linting and build checks for both Go and Python codebases on every push to the `main` branch.