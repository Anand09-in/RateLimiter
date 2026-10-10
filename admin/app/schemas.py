from datetime import datetime
from typing import Literal
from uuid import UUID

from pydantic import BaseModel, Field


class RuleCreate(BaseModel):
    name: str
    algorithm: Literal["fixed", "sliding", "token"]
    limit_count: int = Field(gt=0)
    window_secs: int = Field(gt=0)
    endpoint: str = "*"


class RuleUpdate(BaseModel):
    name: str | None = None
    algorithm: Literal["fixed", "sliding", "token"] | None = None
    limit_count: int | None = Field(default=None, gt=0)
    window_secs: int | None = Field(default=None, gt=0)
    endpoint: str | None = None


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
    custom_limit: int | None = Field(default=None, gt=0)


class ClientOut(BaseModel):
    id: UUID
    name: str
    api_key: str
    rule_id: UUID
    custom_limit: int | None
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
