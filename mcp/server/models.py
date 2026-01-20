from typing import List, Optional
from sqlmodel import Field, SQLModel, Relationship, Column, String
from datetime import datetime, timezone


class Note(SQLModel, table=True):
    __tablename__ = "notes"  # Match Go's table name

    # gorm.Model equivalent fields
    id: Optional[int] = Field(default=None, primary_key=True)
    created_at: Optional[datetime] = Field(
        default_factory=lambda: datetime.now(timezone.utc),
        sa_column=Column("created_at"),
    )
    updated_at: Optional[datetime] = Field(
        default_factory=lambda: datetime.now(timezone.utc),
        sa_column=Column("updated_at"),
    )
    deleted_at: Optional[datetime] = Field(default=None, sa_column=Column("deleted_at"))

    # Your custom fields
    content: str = Field(sa_column=Column("content", String))

    # Note topics removed - no Topic relationship


class NoteBase(SQLModel):
    content: str = Field(sa_column=Column("content", String))
