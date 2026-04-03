@dataclass
class Milestone(BaseEntity):
    """
    Represents a significant point or event in a project.

    Attributes:
        name: The display name of the milestone.
        target_date: The deadline for reaching the milestone.
        status: The current status of the milestone.
        project_id: The ID of the project this milestone belongs to.
        task_ids: A list of IDs representing tasks associated with this milestone.
    """
    name: str
    target_date: datetime
    status: Status
    project_id: uuid.UUID
    task_ids: List[uuid.UUID] = field(default_factory=list)

    def __post_init__(self):
        if not self.name:
            raise ValueError("Milestone name cannot be empty")

    @classmethod
    def create(cls, name: str, target_date: datetime, project_id: uuid.UUID) -> 'Milestone':
        """Factory method to create a new milestone."""
        return cls(
            name=name,
            target_date=target_date,
            status=Status.TODO,
            project_id=project_id
        )

    def add_task(self, task_id: uuid.UUID):
        """Adds a task ID to the milestone association."""
        if task_id not in self.task_ids:
            self.task_ids.append(task_id)
            self.touch()