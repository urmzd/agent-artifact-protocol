@router.patch("/{id}/deactivate", response_model=UserResponse)
def deactivate_user(id: int, db: Session = Depends(get_db)):
    db_user = get_user_by_id(db, id)
    if not db_user:
        raise HTTPException(status_code=404, detail="User not found")
    
    db_user.is_active = False
    db.commit()
    db.refresh(db_user)
    return db_user