from enum import Enum
from typing import Literal

# ... existing code ...

class UserRole(str, Enum):
    ADMIN = "admin"
    EDITOR = "editor"
    VIEWER = "viewer"

class UserCreate(UserBase):
    password: str
    role: UserRole = UserRole.VIEWER

# ... rest of the code ...