<aap:target id="user-module">
from typing import List, Optional
from datetime import datetime
from fastapi import APIRouter, Depends, HTTPException, status
from sqlalchemy.orm import Session
from sqlalchemy import Column, Integer, String, Boolean, DateTime
from sqlalchemy.ext.declarative import declarative_base
from pydantic import BaseModel, EmailStr, Field

Base = declarative_base()

<aap:target id="models">
class User(Base):
    __tablename__ = "users"
    id = Column(Integer, primary_key=True, index=True)
    email = Column(String, unique=True, index=True, nullable=False)
    name = Column(String, nullable=False)
    hashed_password = Column(String, nullable=False)
    role = Column(String, default="user")
    is_active = Column(Boolean, default=True)
    created_at = Column(DateTime, default=datetime.utcnow)
</aap:target>

<aap:target id="schemas">    role: UserRole = UserRole.viewer</aap:target>

<aap:target id="crud">def list_users(db: Session, skip: int = 0, limit: int = 100, role: Optional[str] = None, is_active: Optional[bool] = None):
    query = db.query(User)
    if role:
        query = query.filter(User.role == role)
    if is_active is not None:
        query = query.filter(User.is_active == is_active)
    return query.offset(skip).limit(limit).all()</aap:target>

<aap:target id="router">
router = APIRouter(prefix="/users", tags=["users"])

@router.post("/", response_model=UserResponse)
def create(user: UserCreate, db: Session = Depends()):
    return create_user(db, user)

@router.get("/", response_model=List[UserResponse])
def read_all(skip: int = 0, limit: int = 100, db: Session = Depends()):
    return list_users(db, skip, limit)

@router.get("/{id}", response_model=UserResponse)
def read_one(id: int, db: Session = Depends()):
    return get_user(db, id)

@router.put("/{id}", response_model=UserResponse)
def update(id: int, user: UserUpdate, db: Session = Depends()):
    return update_user(db, id, user)

@router.delete("/{id}")
def delete(id: int, db: Session = Depends()):
    return delete_user(db, id)
</aap:target>
</aap:target>