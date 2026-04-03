def list_users(db: Session, skip: int = 0, limit: int = 10, role: Optional[str] = None, is_active: Optional[bool] = None):
    query = db.query(User)
    
    if role:
        query = query.filter(User.role == role)
    if is_active is not None:
        query = query.filter(User.is_active == is_active)
        
    return query.offset(skip).limit(limit).all()

# Updated Router Endpoint
@router.get("/", response_model=List[UserResponse])
def read_users(
    skip: int = 0, 
    limit: int = 10, 
    role: Optional[str] = None, 
    is_active: Optional[bool] = None, 
    db: Session = Depends(get_db)
):
    return list_users(db, skip=skip, limit=limit, role=role, is_active=is_active)