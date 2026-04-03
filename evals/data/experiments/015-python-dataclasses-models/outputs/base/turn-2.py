class Priority(Enum):
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    URGENT = "urgent"
    CRITICAL = "critical"

    @property
    def color(self) -> str:
        """Returns the hex color code associated with the priority level."""
        colors = {
            Priority.LOW: "#808080",      # Gray
            Priority.MEDIUM: "#008000",   # Green
            Priority.HIGH: "#FFA500",     # Orange
            Priority.URGENT: "#FF4500",   # OrangeRed
            Priority.CRITICAL: "#FF0000"  # Red
        }
        return colors.get(self, "#000000")