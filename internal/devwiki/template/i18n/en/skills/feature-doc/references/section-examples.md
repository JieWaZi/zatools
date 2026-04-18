# Feature Document Section Examples

These are not full documents. They show the expected level of detail for each section.

### Example 1: Core capability subsection

````md
### 2.1 Automatic group synchronization

- Capability summary: The system automatically synchronizes member relationships and permission scopes after a group changes.
- Entry trigger: `GroupSyncService#sync_group_members`
- Key processing: Load the target group, diff the expected members, apply relationship changes, and refresh permission cache.
- Output / side effects: Updates membership records, refreshes cache, and writes an audit event.
- Key constraints: Disabled groups do not trigger sync; manually locked member records are preserved.
- Key code: `group_sync_service.rb#sync_group_members`
````

### Example 2: Module design subsection

````md
### 5.1 GroupSyncService Module

**Module responsibility**: Execute group-member synchronization, refresh permissions, and persist sync results.

**Process / entry information**:

- Process name: `app:group_sync_worker`
- Startup mode: background job trigger
- Entry file: `group_sync_service.rb`
- Entry function: `sync_group_members`

**Core methods**:

|Method|Responsibility|Key design|
|---|---|---|
|`load_target_members`|Load the target group and target member list|Paged reads, skip disabled members|
|`apply_member_delta`|Apply membership delta and refresh cache|Batch writes, rollback on failure, audit record|
````

### Example 3: Interface and message format

````md
### 6.2 External Interfaces / Message Formats

**Request or message format**:

```json
{
  "group_id": "group-001",
  "member_ids": ["u-1001", "u-1002"],
  "operator": "admin@example.com"
}
```

**Response or result format**:

```json
{
  "status": "ok",
  "synced_count": 2
}
```
````

### Example 4: Specification and reliability

````md
### 7.2 Performance / Timing Specifications

|Metric|Default|Notes|
|---|---|---|
|Sync batch size|200 records|Upper bound of members processed in one batch|
|Max retries|3|Upper bound for cache refresh retries|

### 8.1 Reliability Design

- Deduplication or idempotency: Use a sync job ID to prevent duplicate execution.
- Retry and backoff: Retry cache refresh with 1s, 2s, 3s linear backoff.
- Persistence / backup: Store sync results and audit events for replay and investigation.
````
