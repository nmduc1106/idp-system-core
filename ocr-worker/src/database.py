from sqlalchemy import create_engine, text
from src.config import Config
import structlog

logger = structlog.get_logger()

# Tạo Engine với Connection Pool (Tự động reconnect DB nếu rớt)
engine = create_engine(
    Config.DATABASE_URL, 
    pool_size=5, 
    max_overflow=10, 
    pool_pre_ping=True
)

def update_job_status(job_id, state, result=None, error=None):
    """Hàm helper để update status job an toàn"""
    try:
        with engine.begin() as conn:
            if result:
                query = text("UPDATE jobs SET state=:state, result=:result, finished_at=NOW() WHERE id=:id")
                conn.execute(query, {"state": state, "result": result, "id": job_id})
            elif error:
                query = text("UPDATE jobs SET state=:state, error_message=:error WHERE id=:id")
                conn.execute(query, {"state": state, "error": error, "id": job_id})
            else:
                query = text("UPDATE jobs SET state=:state WHERE id=:id")
                conn.execute(query, {"state": state, "id": job_id})
            
            logger.info("db_updated", job_id=job_id, state=state)
    except Exception as e:
        logger.error("db_update_failed", error=str(e))
        raise e

def get_document_path(doc_id):
    with engine.connect() as conn:
        result = conn.execute(text("SELECT storage_path FROM documents WHERE id=:id"), {"id": doc_id})
        row = result.fetchone()
        return row[0] if row else None