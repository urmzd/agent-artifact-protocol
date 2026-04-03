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

<aap:target id="schemas">
class UserBase(BaseModel):
    email: EmailStr
    name: str
    role: str = "user"

class UserCreate(UserBase):
    password: str

class UserUpdate(BaseModel):
    name: Optional[str] = None
    role: Optional[str] = None
    is_active: Optional[bool] = None

class UserResponse(UserBase):
    id: int
    is_active: bool
    created_at: datetime

    class Config:
        orm_mode = True
</aap:target>

<aap:target id="crud">
def get_user(db: Session, user_id: int):
    user = db.query(User).filter(User.id == user_id).first()
    if not user:
        raise HTTPException(status_code=404, detail="User not found")
    return user

def list_users(db: Session, skip: int = 0, limit: int = 100):
    return db.query(User).offset(skip).limit(limit).all()

def create_user(db: Session, user: UserCreate):
    db_user = User(email=user.email, name=user.name, hashed_password=user.password + "_hashed")
    db.add(db_user)
    db.commit()
    db.refresh(db_user)
    return db_user

def update_user(db: Session, user_id: int, user_data: UserUpdate):
    db_user = get_user(db, user_id)
    for key, value in user_data.dict(exclude_unset=True).items():
        setattr(db_user, key, value)
    db.commit()
    return db_user

def delete_user(db: Session, user_id: int):
    db_user = get_user(db, user_id)
    db.delete(db_user)
    db.commit()
    return {"detail": "User deleted"}
</aap:target>

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