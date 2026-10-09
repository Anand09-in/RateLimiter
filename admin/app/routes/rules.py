from uuid import UUID
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session

from ..database import get_db
from ..models import Rule, AuditLog
from ..schemas import RuleCreate, RuleUpdate, RuleOut

router = APIRouter(prefix="/rules", tags=["rules"])


def _rule_dict(rule: Rule) -> dict:
    return {
        "id": str(rule.id),
        "name": rule.name,
        "algorithm": rule.algorithm,
        "limit_count": rule.limit_count,
        "window_secs": rule.window_secs,
        "endpoint": rule.endpoint,
        "is_active": rule.is_active,
    }


def _audit(db: Session, action: str, rule: Rule, before: dict | None, after: dict | None):
    db.add(AuditLog(
        action=action,
        entity_type="rule",
        entity_id=rule.id,
        payload={"before": before, "after": after},
    ))


@router.post("", response_model=RuleOut, status_code=201)
def create_rule(body: RuleCreate, db: Session = Depends(get_db)):
    rule = Rule(**body.model_dump())
    db.add(rule)
    db.flush()
    _audit(db, "CREATE", rule, None, _rule_dict(rule))
    db.commit()
    db.refresh(rule)
    return rule


@router.get("", response_model=list[RuleOut])
def list_rules(db: Session = Depends(get_db)):
    return db.query(Rule).filter(Rule.is_active == True).all()


@router.put("/{rule_id}", response_model=RuleOut)
def update_rule(rule_id: UUID, body: RuleUpdate, db: Session = Depends(get_db)):
    rule = db.query(Rule).filter(Rule.id == rule_id, Rule.is_active == True).first()
    if not rule:
        raise HTTPException(status_code=404, detail="rule not found")
    before = _rule_dict(rule)
    for field, value in body.model_dump(exclude_none=True).items():
        setattr(rule, field, value)
    db.flush()
    _audit(db, "UPDATE", rule, before, _rule_dict(rule))
    db.commit()
    db.refresh(rule)
    return rule


@router.delete("/{rule_id}", status_code=204)
def delete_rule(rule_id: UUID, db: Session = Depends(get_db)):
    rule = db.query(Rule).filter(Rule.id == rule_id, Rule.is_active == True).first()
    if not rule:
        raise HTTPException(status_code=404, detail="rule not found")
    before = _rule_dict(rule)
    rule.is_active = False
    db.flush()
    _audit(db, "DELETE", rule, before, _rule_dict(rule))
    db.commit()
