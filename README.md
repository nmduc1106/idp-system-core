# Intelligent Document Processing (IDP) System 🚀

An enterprise-grade, microservices-based Document Processing System that leverages Multimodal LLMs (Anthropic Claude 3.5) to automatically extract structured JSON data from raw images and documents with **Real-time Event Streaming**.

## 🏗️ System Architecture

The system is built with a highly resilient, asynchronous event-driven architecture:

1. **Frontend (React + TS):** A modern dashboard built with **TypeScript** and **Tailwind CSS**. It features a secure **HttpOnly Cookie** authentication flow and real-time status updates via **Server-Sent Events (SSE)**.
2. **API Gateway (Golang):** High-performance gateway managing JWT sessions, file uploads, and **SSE streaming**. It acts as a bridge between Redis Pub/Sub and the Client.
3. **Redis Pub/Sub:** Serves as the real-time backbone, allowing the Python Worker to notify the Go Gateway immediately upon job completion.
4. **Message Broker (RabbitMQ):** Decouples ingestion from processing, ensuring no jobs are lost during traffic spikes.
5. **OCR AI Worker (Python):** Uses **Anthropic Claude Vision API** with Pydantic schemas for structured extraction. Publishes results to Redis for instant frontend updates.



## 🔐 Security Features

- **HttpOnly & Secure Cookies:** Protects JWT tokens from XSS attacks by keeping them inaccessible to JavaScript.
- **Data Ownership Isolation:** Strict database-level checks ensure users can only stream and view their own documents.
- **Zero-Token URL Policy:** SSE streams use native browser cookie mechanics instead of exposing tokens in query strings.

## 📂 Project Structure

- `/api-gateway` - Golang REST API, SSE Engine & Producer.
- `/frontend` - React TypeScript SPA (Real-time Dashboard).
- `/ocr-worker` - Python Worker & AI Engine.
- `docker-compose.yml` - Infrastructure (Postgres, Redis, RabbitMQ, MinIO).

## 🚀 Getting Started

1. **Infrastructure:** `docker-compose up -d`
2. **Backend:** `cd api-gateway && go run cmd/api/main.go`
3. **Worker:** (Ensure `.env` is set) `cd ocr-worker && python src/worker.py`
4. **Frontend:** `cd frontend && npm install && npm run dev`

## 🛠️ CI/CD
This repository includes a GitHub Actions pipeline (`.github/workflows/ci.yml`) that automatically runs parallel linting and build checks for both Go and Python codebases on every push to the `main` branch.