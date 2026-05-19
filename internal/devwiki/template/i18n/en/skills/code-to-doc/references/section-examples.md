# Workflow Section Examples

These are not full documents. They show the expected level of detail for each workflow section.

### Entry Points

```md
## Entry Points

| Entry | Type | Notes |
|---|---|---|
| `POST /api/users` | HTTP API | Create-user request entry |
| `UserService#create_user` | service method | Main business logic after controller validation |
```

### Call Chain

```md
## Call Chain

1. `users_controller#create` parses request parameters and runs basic validation.
2. `UserService#create_user` handles uniqueness checks, user creation, and default role assignment.
3. `AuditService#record` writes the audit event.
4. `NotificationClient#send_user_created` sends the notification.
```

### Key Logic

```md
## Key Logic

- Username and email uniqueness must be checked before create.
- User creation and default role assignment share one transaction boundary.
- Notification failure does not roll back creation, but writes a retry job.
```

### Code References

```yaml
code_refs:
  - path: "app/controllers/users_controller.rb"
    kind: function
    symbol: "create"
    note: "HTTP create entry"
    confidence: 0.95
  - path: "app/services/user_service.rb"
    kind: function
    symbol: "create_user"
    note: "Main user creation logic"
    confidence: 0.95
```

### Change Impact

```md
## Change Impact

- Changing parameter validation affects the create API and bulk import.
- Changing default role assignment affects initial permission state.
- Changing transaction boundaries requires failure rollback and audit consistency tests.
```
