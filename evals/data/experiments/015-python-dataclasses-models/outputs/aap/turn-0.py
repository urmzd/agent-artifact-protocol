from dataclasses import dataclass, field
from datetime import datetime
from enum import Enum
from typing import List, Optional, Dict
from uuid import UUID, uuid4

<aap:target id="core-models">
class Priority(Enum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    URGENT = "urgent"

class Status(Enum):
    TODO = "todo"
    IN_PROGRESS = "in_progress"
    REVIEW = "review"
    DONE = "done"

@dataclass(frozen=True)
class TimeEstimate:
    <aap:target id="time-estimate-value">hours: float</aap:target>

@dataclass(frozen=True)
class DateRange:
    <aap:target id="date-range-start">start: datetime</aap:target>
    <aap:target id="date-range-end">end: datetime</aap:target>

@dataclass
class BaseEntity:
    <aap:target id="base-fields">
    id: UUID = field(default_factory=uuid4)
    created_at: datetime = field(default_factory=datetime.utcnow)
    updated_at: datetime = field(default_factory=datetime.utcnow)
    </aap:target>

@dataclass
class Project(BaseEntity):
    <aap:target id="project-fields">
    name: str
    owner_id: UUID
    team_id: UUID
    </aap:target>

@dataclass
class Sprint(BaseEntity):
    <aap:target id="sprint-fields">
    project_id: UUID
    duration: DateRange
    status: Status = Status.TODO
    </aap:target>

@dataclass
class Task(BaseEntity):
    <aap:target id="task-fields">
    project_id: UUID
    sprint_id: Optional[UUID]
    assignee_id: Optional[UUID]
    title: str
    priority: Priority = Priority.MEDIUM
    status: Status = Status.TODO
    estimate: Optional[TimeEstimate] = None
    </aap:target>

@dataclass
class User(BaseEntity):
    <aap:target id="user-fields">
    username: str
    email: str
    </aap:target>

@dataclass
class Team(BaseEntity):
    <aap:target id="team-fields">
    name: str
    member_ids: List[UUID] = field(default_factory=list)
    </aap:target>

@dataclass
class Comment(BaseEntity):
    <aap:target id="comment-fields">
    task_id: UUID
    author_id: UUID
    text: str
    </aap:target>

class ModelFactory:
    <aap:target id="factory-methods">
    @staticmethod
    def create_task(title: str, project_id: UUID) -> Task:
        if not title:
            raise ValueError("Task title cannot be empty")
        return Task(title=title, project_id=project_id)
    </aap:target>
</aap:target>