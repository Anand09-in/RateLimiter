import secrets
from uuid import UUID
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Client, AuditLog
from ..schemas import ClientCreate, ClientOut

router = APIRouter(prefix="/clients", tags=["clients"])


def _client_dict(c: Client) -> dict:
    return {
        "id": str(c.id),
        "name": c.name,
        "api_key": c.api_key,
        "rule_id": str(c.rule_id),
        "custom_limit": c.custom_limit,
        "is_active": c.is_active,
    }


def _audit(db: Session, action: str, client: Client, before: dict | None, after: dict | None):
    db.add(AuditLog(
        action=action,
        entity_type="client",
        entity_id=client.id,
        payload={"before": before, "after": after},
    ))


@router.post("", response_model=ClientOut, status_code=201)
def create_client(body: ClientCreate, db: Session = Depends(get_db)):
    client = Client(**body.model_dump(), api_key=secrets.token_urlsafe(32))
    db.add(client)
    db.flush()
    _audit(db, "CREATE", client, None, _client_dict(client))
    db.commit()
    db.refresh(client)
    return client


@router.get("", response_model=list[ClientOut])
def list_clients(db: Session = Depends(get_db)):
    return db.query(Client).filter(Client.is_active == True).all()


@router.get("/{client_id}", response_model=ClientOut)
def get_client(client_id: UUID, db: Session = Depends(get_db)):
    client = db.query(Client).filter(Client.id == client_id).first()
    if not client:
        raise HTTPException(status_code=404, detail="client not found")
    return client


@router.delete("/{client_id}", status_code=204)
def delete_client(client_id: UUID, db: Session = Depends(get_db)):
    client = db.query(Client).filter(Client.id == client_id, Client.is_active == True).first()
    if not client:
        raise HTTPException(status_code=404, detail="client not found")
    before = _client_dict(client)
    client.is_active = False
    db.flush()
    _audit(db, "DELETE", client, before, _client_dict(client))
    db.commit()
