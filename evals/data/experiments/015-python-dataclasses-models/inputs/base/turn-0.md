Create Python dataclass models for a project management system.

Include:
- Base classes with common fields (id, created_at, updated_at)
- Entity models: Project, Sprint, Task, User, Team, Comment
- Value objects: Priority (enum), Status (enum), TimeEstimate, DateRange
- Relationships between models using IDs
- Validation methods and factory classmethods
- Type hints and docstrings

Use section IDs: base, entities, value-objects

Use AAP section markers to delineate each major code block.
Wrap each logical section with `<aap:section id="ID">` and `</aap:section>`.


Output raw code only. No markdown fences, no explanation.