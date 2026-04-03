import uuid
from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import List, Optional, Dict, Any


class Priority(Enum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class Status(Enum):
    TODO = "todo"
    IN_PROGRESS = "in_progress"
    REVIEW = "review"
    DONE = "done"


@dataclass(frozen=True)
class TimeEstimate:
    hours: int
    minutes: int = 0

    def __post_init__(self):
        if self.hours < 0 or self.minutes < 0:
            raise ValueError("Time values cannot be negative")


@dataclass(frozen=True)
class DateRange:
    start_date: datetime
    end_date: datetime

    def __post_init__(self):
        if self.end_date < self.start_date:
            raise ValueError("End date must be after start date")


@dataclass
class BaseEntity:
    id: uuid.UUID = field(default_factory=uuid.uuid4)
    created_at: datetime = field(default_factory=datetime.now)
    updated_at: datetime = field(default_factory=datetime.now)

    def touch(self):
        self.updated_at = datetime.now()


@dataclass
class User(BaseEntity):
    username: str
    email: str


@dataclass
class Team(BaseEntity):
    name: str
    member_ids: List[uuid.UUID] = field(default_factory=list)


@dataclass
class Project(BaseEntity):
    name: str
    owner_id: uuid.UUID
    team_ids: List[uuid.UUID] = field(default_factory=list)


@dataclass
class Sprint(BaseEntity):
    project_id: uuid.UUID
    name: str
    duration: DateRange


@dataclass
class Task(BaseEntity):
    title: str
    description: str
    project_id: uuid.UUID
    sprint_id: Optional[uuid.UUID] = None
    assignee_id: Optional[uuid.UUID] = None
    priority: Priority = Priority.MEDIUM
    status: Status = Status.TODO
    estimate: Optional[TimeEstimate] = None

    @classmethod
    def create(cls, title: str, project_id: uuid.UUID, **kwargs):
        return cls(title=title, project_id=project_id, **kwargs)


@dataclass
class Comment(BaseEntity):
    task_id: uuid.UUID
    author_id: uuid.UUID
    content: str

    def __post_init__(self):
        if not self.content.strip():
            raise ValueError("Comment content cannot be empty")


class ModelFactory:
    """Factory to handle complex entity creation logic."""

    @staticmethod
    def create_task(title: str, project_id: uuid.UUID, **kwargs) -> Task:
        return Task(
            title=title,
            project_id=project_id,
            **kwargs
        )

    @staticmethod
    def create_user(username: str, email: str) -> User:
        return User(username=username, email=email)