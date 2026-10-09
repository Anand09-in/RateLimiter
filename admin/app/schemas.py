from typing import Literal, Optional
from uuid import UUID
from datetime import datetime
from pydantic import BaseModel, Field


class RuleCreate(BaseModel):
    name: str
    algorithm: Literal["fixed", "sliding", "token"]
    limit_count: int = Field(gt=0)
    window_secs: int = Field(gt=0)
    endpoint: str = "*"


class RuleUpdate(BaseModel):
    name: Optional[str] = None
    algorithm: Optional[Literal["fixed", "sliding", "token"]] = None
    limit_count: Optional[int] = Field(default=None, gt=0)
    window_secs: Optional[int] = Field(default=None, gt=0)
    endpoint: Optional[str] = None


class RuleOut(BaseModel):
    id: UUID
    name: str
    algorithm: str
    limit_count: int
    window_secs: int
    endpoint: str
    is_active: bool
    created_at: datetime
    updated_at: datetime

    model_config = {"from_attributes": True}


class ClientCreate(BaseModel):
    name: str
    rule_id: UUID
    custom_limit: Optional[int] = Field(default=None, gt=0)


class ClientOut(BaseModel):
    id: UUID
    name: str
    api_key: str
    rule_id: UUID
    custom_limit: Optional[int]
    is_active: bool
    created_at: datetime

    model_config = {"from_attributes": True}


class AuditLogOut(BaseModel):
    id: UUID
    action: str
    entity_type: str
    entity_id: UUID
    changed_by: str
    payload: dict
    created_at: datetime

    model_config = {"from_attributes": True}
