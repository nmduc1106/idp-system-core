import time
import json
import signal
import sys
import threading
import pika
import redis
from http.server import HTTPServer, BaseHTTPRequestHandler
from src.config import Config
from src.database import update_job_status, get_document_path
from src.processor import OCRProcessor
import structlog
import os
import uuid
from opentelemetry import trace, propagate
from opentelemetry.trace.status import Status, StatusCode
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.otlp.proto.grpc.trace_exporter import OTLPSpanExporter
from opentelemetry.sdk.resources import Resource

logger = structlog.get_logger()

# --- 0. Setup Tracing ---
def setup_tracing():
    resource = Resource(attributes={
        "service.name": "ocr-worker",
        "service.instance.id": os.getenv("HOSTNAME", str(uuid.uuid4()))
    })
    provider = TracerProvider(resource=resource)
    
    # OTLP gRPC Exporter — reads from Docker Compose env, falls back to localhost for local dev
    otel_endpoint = os.getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "http://localhost:4317")
    exporter = OTLPSpanExporter(endpoint=otel_endpoint, insecure=True)
    processor = BatchSpanProcessor(
        exporter,
        max_queue_size=2048,
        max_export_batch_size=512
    )
    
    provider.add_span_processor(processor)
    trace.set_tracer_provider(provider)
    
    return trace.get_tracer(__name__)

tracer = setup_tracing()

# --- 1. Healthcheck Server (Để Docker kiểm tra sống/chết) ---
class HealthHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        self.send_response(200)
        self.end_headers()
        self.wfile.write(b"OK")

def run_health_server():
    server = HTTPServer(('0.0.0.0', 8000), HealthHandler)
    logger.info("health_check_started", port=8000)
    server.serve_forever()

# --- 2. RabbitMQ Consumer ---
class RabbitMQConsumer:
    def __init__(self):
        self.connection = None
        self.channel = None
        self.processor = OCRProcessor() # Gọi Singleton
        self.redis_client = None
        try:
            self.redis_client = redis.from_url(Config.REDIS_URL)
            logger.info("redis_connected")
        except Exception as e:
            logger.error("redis_connection_failed", error=str(e))

    def connect(self):
        """Kết nối RabbitMQ với cơ chế Retry vô hạn"""
        while True:
            try:
                params = pika.URLParameters(Config.RABBITMQ_URL)
                self.connection = pika.BlockingConnection(params)
                self.channel = self.connection.channel()
                
                # Setup Topology (Đảm bảo Queue tồn tại)
                self.channel.exchange_declare(exchange='doc_exchange', exchange_type='direct', durable=True)
                self.channel.queue_declare(queue=Config.QUEUE_NAME, durable=True)
                # FIX 1: Routing key phải khớp với Go
                self.channel.queue_bind(exchange='doc_exchange', queue=Config.QUEUE_NAME, routing_key='ocr_queue')                
                
                # QoS: Chỉ nhận 1 job/lần (Tránh quá tải RAM)
                self.channel.basic_qos(prefetch_count=1)
                
                self.channel.basic_consume(
                    queue=Config.QUEUE_NAME, 
                    on_message_callback=self.on_message
                )
                logger.info("rabbitmq_connected")
                return
            except pika.exceptions.AMQPConnectionError:
                logger.error("rabbitmq_connection_failed", retry_in=5)
                time.sleep(5)

    def on_message(self, ch, method, properties, body):
        """Xử lý từng message"""
        # FIX 2: Decode bytes sang string để OTel đọc được Context
        raw_headers = properties.headers if properties.headers else {}
        clean_headers = {}
        for k, v in raw_headers.items():
            clean_headers[k] = v.decode('utf-8') if isinstance(v, bytes) else str(v)
            
        ctx = propagate.extract(carrier=clean_headers)

        with tracer.start_as_current_span("process_ocr_job", context=ctx) as span:
            try:
                data = json.loads(body)
                job_id = data.get('job_id')
                doc_id = data.get('doc_id')
                
                if not job_id or not doc_id:
                    logger.warning("invalid_message", body=body)
                    span.set_attribute("error", "invalid message format")
                    span.set_status(Status(StatusCode.ERROR, "Invalid message payload"))
                    ch.basic_ack(delivery_tag=method.delivery_tag)
                    return

                span.set_attribute("job_id", job_id)
                span.set_attribute("doc_id", doc_id)

                logger.info("job_received", job_id=job_id)
                update_job_status(job_id, "PROCESSING")

                # A. Lấy đường dẫn file
                object_name = get_document_path(doc_id)
                if not object_name:
                    raise ValueError("Document not found in DB")

                # B. Chạy Pipeline
                result = self.processor.process(object_name)

                # C. Update Thành công
                update_job_status(job_id, "COMPLETED", result=result)
                
                # Publish to Redis for SSE
                if self.redis_client:
                    try:
                        payload = json.dumps({
                            "job_id": job_id,
                            "status": "COMPLETED",
                            "result": result
                        })
                        self.redis_client.publish(f"job_status:{job_id}", payload)
                    except Exception as redis_err:
                        logger.error("redis_publish_failed", error=str(redis_err))

                ch.basic_ack(delivery_tag=method.delivery_tag)
                logger.info("job_completed", job_id=job_id)

            except Exception as e:
                span.record_exception(e)
                span.set_status(Status(StatusCode.ERROR, str(e)))
                logger.error("job_failed", error=str(e), job_id=job_id if 'job_id' in locals() else 'unknown')
                
                # Xử lý lỗi: Cập nhật DB là FAILED
                if 'job_id' in locals():
                    update_job_status(job_id, "FAILED", error=str(e))
                    
                    # Publish to Redis for SSE
                    if self.redis_client:
                        try:
                            payload = json.dumps({
                                "job_id": job_id,
                                "status": "FAILED",
                                "error_message": str(e)
                            })
                            self.redis_client.publish(f"job_status:{job_id}", payload)
                        except Exception as redis_err:
                            logger.error("redis_publish_failed", error=str(redis_err))
                
                # Reject message (Không requeue để tránh lặp vô tận)
                ch.basic_nack(delivery_tag=method.delivery_tag, requeue=False)

    def start(self):
        self.connect()
        try:
            logger.info("worker_started_consuming")
            self.channel.start_consuming()
        except KeyboardInterrupt:
            self.stop()
        except pika.exceptions.ConnectionClosedByBroker:
            logger.warning("connection_closed_by_broker_reconnecting...")
            self.connect()

    def stop(self):
        logger.info("worker_stopping...")
        if self.connection and self.connection.is_open:
            self.connection.close()
        sys.exit(0)

# --- 3. Main Entry ---
if __name__ == "__main__":
    consumer = RabbitMQConsumer()
    
    # Xử lý tín hiệu tắt (Ctrl+C)
    def signal_handler(sig, frame):
        consumer.stop()
    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    # Chạy Healthcheck ở background thread
    health_thread = threading.Thread(target=run_health_server, daemon=True)
    health_thread.start()

    # Chạy Consumer
    consumer.start()