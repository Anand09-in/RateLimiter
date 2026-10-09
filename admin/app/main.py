from fastapi import FastAPI, Depends
from sqlalchemy.orm import Session
from sqlalchemy import text

from .database import get_db, engine
from .models import Base, AuditLog
from .schemas import AuditLogOut
from .routes import rules, clients

Base.metadata.create_all(bind=engine)

app = FastAPI(title="Rate Limiter Admin API")

app.include_router(rules.router)
app.include_router(clients.router)


@app.get("/audit", response_model=list[AuditLogOut])
def list_audit(limit: int = 50, offset: int = 0, db: Session = Depends(get_db)):
    return db.query(AuditLog).order_by(AuditLog.created_at.desc()).offset(offset).limit(limit).all()


@app.get("/health")
def health(db: Session = Depends(get_db)):
    try:
        db.execute(text("SELECT 1"))
        db_status = "ok"
    except Exception:
        db_status = "error"
    return {"status": "ok", "db": db_status}
