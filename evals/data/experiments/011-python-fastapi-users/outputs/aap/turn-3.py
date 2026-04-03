{
  "protocol": "aap/0.1",
  "id": "user-module",
  "version": 2,
  "name": "edit",
  "operation": {"direction": "input", "format": "text/html"},
  "content": [
    {
      "op": "insert_after",
      "target": {"type": "id", "value": "router"},
      "content": "

@router.patch(\"/{id}/deactivate\", response_model=UserResponse)
def deactivate(id: int, db: Session = Depends()):
    user = get_user(db, id)
    if not user:
        raise HTTPException(status_code=404, detail=\"User not found\")
    user.is_active = False
    db.commit()
    db.refresh(user)
    return user"
    }
  ]
}