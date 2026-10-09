"""initial schema

Revision ID: 0001
Revises:
Create Date: 2026-10-09

"""
from typing import Sequence, Union

import sqlalchemy as sa
from sqlalchemy.dialects.postgresql import UUID, JSONB
from alembic import op

revision: str = "0001"
down_revision: Union[str, None] = None
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.execute('CREATE EXTENSION IF NOT EXISTS "pgcrypto"')

    op.create_table(
        "rules",
        sa.Column("id", UUID(as_uuid=True), primary_key=True, server_default=sa.text("gen_random_uuid()")),
        sa.Column("name", sa.Text, nullable=False),
        sa.Column("algorithm", sa.Text, nullable=False),
        sa.Column("limit_count", sa.Integer, nullable=False),
        sa.Column("window_secs", sa.Integer, nullable=False),
        sa.Column("endpoint", sa.Text, nullable=False, server_default="*"),
        sa.Column("is_active", sa.Boolean, nullable=False, server_default="true"),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False, server_default=sa.text("NOW()")),
        sa.Column("updated_at", sa.DateTime(timezone=True), nullable=False, server_default=sa.text("NOW()")),
        sa.CheckConstraint("algorithm IN ('fixed', 'sliding', 'token')", name="ck_rules_algorithm"),
    )

    op.create_table(
        "clients",
        sa.Column("id", UUID(as_uuid=True), primary_key=True, server_default=sa.text("gen_random_uuid()")),
        sa.Column("name", sa.Text, nullable=False),
        sa.Column("api_key", sa.Text, nullable=False, unique=True),
        sa.Column("rule_id", UUID(as_uuid=True), sa.ForeignKey("rules.id"), nullable=False),
        sa.Column("custom_limit", sa.Integer, nullable=True),
        sa.Column("is_active", sa.Boolean, nullable=False, server_default="true"),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False, server_default=sa.text("NOW()")),
    )

    op.create_table(
        "audit_log",
        sa.Column("id", UUID(as_uuid=True), primary_key=True, server_default=sa.text("gen_random_uuid()")),
        sa.Column("action", sa.Text, nullable=False),
        sa.Column("entity_type", sa.Text, nullable=False),
        sa.Column("entity_id", UUID(as_uuid=True), nullable=False),
        sa.Column("changed_by", sa.Text, nullable=False, server_default="api"),
        sa.Column("payload", JSONB, nullable=False),
        sa.Column("created_at", sa.DateTime(timezone=True), nullable=False, server_default=sa.text("NOW()")),
    )

    op.create_index("idx_clients_api_key", "clients", ["api_key"])
    op.create_index("idx_audit_entity", "audit_log", ["entity_type", "entity_id"])


def downgrade() -> None:
    op.drop_index("idx_audit_entity", table_name="audit_log")
    op.drop_index("idx_clients_api_key", table_name="clients")
    op.drop_table("audit_log")
    op.drop_table("clients")
    op.drop_table("rules")
