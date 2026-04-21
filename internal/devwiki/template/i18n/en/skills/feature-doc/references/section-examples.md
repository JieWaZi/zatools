# Feature Document Section Examples

These are not full documents. They show the expected level of detail for each section.

### Example 1: Business scenario overview

````md
### 1.3 Business Scenario Overview

|Scenario|Trigger|Core object / data|Result / impact scope|
|---|---|---|---|
|User creation|An administrator submits the create-user form|User profile, organization, initial role|Creates the user record, links the user to the organization, and enables login|
|User disablement|An administrator disables a user|User status and active sessions|Updates the user state, blocks future login, and clears current sessions|
````

### Example 2: Core capability subsection

````md
### 2.1 User creation

- Capability summary: The system lets administrators create new users and initialize their basic identity data.
- Business trigger: An administrator submits the create-user form in the management console.
- Core object / data: User profile, organization, initial roles, and login credentials.
- Result / impact scope: Creates the user record, links the user to the organization, and enables system login.
- Key processing: Validate required fields and uniqueness, persist the user, assign default roles, and send an initialization notification.
- Key constraints: Username and email must be unique; disabled organizations cannot create new users.
- Supporting implementation: `UserService#create_user`
````

### Example 3: Module design subsection

````md
### 5.1 UserService Module

**Module responsibility**: Handle user creation, updates, disablement, and lookup.

**Process / entry information**:

- Process name: `web`
- Startup mode: HTTP request trigger
- Entry file: `user_service.rb`
- Entry function: `create_user`

**Design purpose**: Keep user lifecycle operations in the service layer instead of spreading business rules across controllers.

**Core methods**:

|Method|Responsibility|Key design|
|---|---|---|
|`create_user`|Create a user and initialize default state|Validate uniqueness, persist, then assign default roles|
|`disable_user`|Disable a user and block later login|Update state and clear active sessions|
````

### Example 4: API interface

````md
### 6.1 API Interfaces

|Interface|URL|Method|Description|
|---|---|---|---|
|`create_user`|`/api/users`|`POST`|Create a new user and return the user identifier|

**Request format**:

```json
{
  "username": "alice",
  "email": "alice@example.com",
  "organization_id": "org-001",
  "role_ids": ["role-admin"],
  "operator": "admin@example.com"
}
```

**Response format**:

```json
{
  "status": "ok",
  "user_id": "user-1001"
}
```
````

### Example 5: Message / subscription interface

````md
### 6.2 Message / Subscription Interfaces

|Message or topic|Producer / consumer|Trigger|Description|
|---|---|---|---|
|`user.created`|`UserService -> AuditSubscriber`|After user creation succeeds|Tell the audit module to persist a user-created event|

**Message structure**:

```json
{
  "event": "user.created",
  "user_id": "user-1001",
  "organization_id": "org-001",
  "operator": "admin@example.com"
}
```
````

### Example 6: Function interface

````md
### 6.3 Function Interfaces

|Function|Caller|Description|
|---|---|---|
|`create_user(params)`|`UserController`|Validate input, create the user, and trigger follow-up initialization logic|

**Input parameters**:

```json
{
  "username": "alice",
  "email": "alice@example.com",
  "organization_id": "org-001",
  "role_ids": ["role-admin"]
}
```
````

### Example 7: Specification and reliability

````md
### 7.2 Performance / Timing Specifications

|Metric|Default|Notes|
|---|---|---|
|Create-user timeout|`5s`|Upper bound of one user-creation request|
|Max retries|3|Upper bound for notification or audit retries|

### 8.1 Reliability Design

- Deduplication or idempotency: Use username and email uniqueness to prevent duplicate creation.
- Retry and backoff: Retry notification delivery with 1s, 2s, 3s linear backoff.
- Persistence / backup: Store user records and audit events for investigation and replay.
````
