package audit

import (
	"context"
	"net/http"
	"strings"

	"github.com/onnwee/subcults/internal/middleware"
)

// LogAccess is a helper function that records an access event to the audit log.
// It extracts user DID and request ID from the context if available.
// entityType: Type of entity accessed (e.g., "scene", "event", "admin_panel")
// entityID: ID of the entity accessed
// action: Action performed (e.g., "access_precise_location", "view_admin_panel")
func LogAccess(ctx context.Context, repo Repository, entityType, entityID, action string) error {
	entry := LogEntry{
		UserDID:    middleware.GetUserDID(ctx),
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		RequestID:  middleware.GetRequestID(ctx),
	}

	_, err := repo.LogAccess(entry)
	return err
}

// LogAccessFromRequest is a helper function that records an access event with HTTP request metadata.
// It extracts user DID, request ID, IP address, and user agent from the request/context.
func LogAccessFromRequest(r *http.Request, repo Repository, entityType, entityID, action string) error {
	// Extract IP address from request
	// X-Forwarded-For can contain multiple IPs (client, proxy1, proxy2, ...)
	// Use the first (leftmost) IP which represents the original client
	ipAddress := r.RemoteAddr
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// Split on comma and take the first IP, trimming whitespace
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 && strings.TrimSpace(ips[0]) != "" {
			ipAddress = strings.TrimSpace(ips[0])
		}
	}

	entry := LogEntry{
		UserDID:    middleware.GetUserDID(r.Context()),
		EntityType: entityType,
		EntityID:   entityID,
		Action:     action,
		RequestID:  middleware.GetRequestID(r.Context()),
		IPAddress:  ipAddress,
		UserAgent:  r.UserAgent(),
	}

	_, err := repo.LogAccess(entry)
	return err
}
