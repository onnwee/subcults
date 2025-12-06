# Audit Logging

The audit logging package provides comprehensive access tracking for sensitive endpoints and operations, supporting compliance requirements and incident response.

## Overview

Audit logs record access events with:
- User identity (DID)
- Entity type and ID accessed
- Action performed
- Timestamp (UTC)
- Request metadata (request ID, IP address, user agent)

## Database Schema

The `audit_logs` table includes:
- Primary key: UUID
- Indexed columns for efficient querying:
  - `entity_type`, `entity_id`, `created_at` (composite index)
  - `user_did`, `created_at`
  - `action`, `created_at`
  - `created_at` (for retention policy queries)

## Usage Examples

### Basic Logging with Context

```go
import (
    "github.com/onnwee/subcults/internal/audit"
    "github.com/onnwee/subcults/internal/middleware"
)

// In a handler function with context
func handlePreciseLocationAccess(ctx context.Context, repo audit.Repository) error {
    // Log access to precise location
    err := audit.LogAccess(
        ctx,
        repo,
        "scene",                      // entity type
        "scene-123",                  // entity ID
        "access_precise_location",    // action
    )
    if err != nil {
        return err
    }
    
    // Continue with actual access...
    return nil
}
```

### Logging with HTTP Request Metadata

```go
// In an HTTP handler
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Log access with full request metadata
    // IP address handling:
    // - Uses X-Forwarded-For header's first IP (original client) if present
    // - Falls back to RemoteAddr if X-Forwarded-For is empty or invalid
    err := audit.LogAccessFromRequest(
        r,
        h.auditRepo,
        "admin_panel",
        "privacy_settings",
        "view_admin_panel",
    )
    if err != nil {
        // Handle error...
    }
    
    // Continue with handler logic...
}
```

### Querying Audit Logs

```go
// Query by entity (e.g., all access to a specific scene)
logs, err := repo.QueryByEntity("scene", "scene-123", 10) // limit to 10 most recent
if err != nil {
    return err
}

for _, log := range logs {
    fmt.Printf("User %s accessed %s at %s\n", 
        log.UserDID, log.Action, log.CreatedAt)
}

// Query by user (e.g., all access by a specific user)
userLogs, err := repo.QueryByUser("did:web:example.com:user123", 0) // no limit
if err != nil {
    return err
}
```

## Common Actions

Standard action names for consistency:

### Location Access
- `access_precise_location` - Viewing precise coordinates
- `access_coarse_location` - Viewing coarse/fuzzy location

### Administrative
- `view_admin_panel` - Accessing admin interface
- `view_privacy_settings` - Viewing privacy configuration
- `modify_privacy_settings` - Changing privacy settings

### Scene/Event Management
- `view_scene_details` - Viewing scene information
- `view_event_details` - Viewing event information
- `export_member_data` - Exporting user data

## Integration Points

Audit logging should be invoked at:

1. **Precise Location Endpoints** - Any API endpoint that returns precise geographic coordinates
2. **Admin Privacy Panel** - When administrators access privacy-related settings
3. **Data Export** - When user data is exported or downloaded
4. **Permission Changes** - When location consent or privacy settings are modified

## Testing

The package includes comprehensive tests with in-memory repository implementation:

```bash
go test -v ./internal/audit/...
```

## Performance Considerations

- Indexes are created for all common query patterns
- Logs are sorted by time (newest first) in query results
- Use limit parameter in queries to control result size
- Consider implementing retention policies (not yet implemented)

## Future Enhancements

- Retention policy automation (e.g., delete logs older than X days)
- Postgres repository implementation for production use
- Audit log export functionality
- Real-time monitoring/alerting for suspicious access patterns
