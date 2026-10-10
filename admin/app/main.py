from fastapi import FastAPI
from sqlalchemy import text
from sqlalchemy.exc import SQLAlchemyError

from .database import DbDep
from .models import AuditLog
from .routes import clients, rules
from .schemas import AuditLogOut

app = FastAPI(title="Rate Limiter Admin API")

app.include_router(rules.router)
app.include_router(clients.router)


@app.get("/audit", response_model=list[AuditLogOut])
def list_audit(limit: int = 50, offset: int = 0, db: DbDep = None):
    return db.query(AuditLog).order_by(AuditLog.created_at.desc()).offset(offset).limit(limit).all()


@app.get("/health")
def health(db: DbDep = None):
    try:
        db.execute(text("SELECT 1"))
        db_status = "ok"
    except SQLAlchemyError:
        db_status = "error"
    return {"status": "ok", "db": db_status}
