package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onnwee/subcults/internal/middleware"
	"github.com/onnwee/subcults/internal/scene"
)

func TestCreateOrUpdateRSVP_Success(t *testing.T) {
	rsvpRepo := scene.NewInMemoryRSVPRepository()
	eventRepo := scene.NewInMemoryEventRepository()
	handlers := NewRSVPHandlers(rsvpRepo, eventRepo)

	// Create a future event
	futureTime := time.Now().Add(24 * time.Hour)
	event := &scene.Event{
		ID:            "event-1",
		SceneID:       "scene-1",
		Title:         "Test Event",
		CoarseGeohash: "dr5regw",
		StartsAt:      futureTime,
	}
	if err := eventRepo.Insert(event); err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	// Create request
	reqBody := RSVPRequest{Status: "going"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/events/event-1/rsvp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	// Add user DID to context (simulating auth middleware)
	ctx := middleware.SetUserDID(req.Context(), "did:plc:user1")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Call handler
	handlers.CreateOrUpdateRSVP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify RSVP was created
	stored, err := rsvpRepo.GetByEventAndUser("event-1", "did:plc:user1")
	if err != nil {
		t.Fatalf("Failed to get RSVP: %v", err)
	}
	if stored.Status != "going" {
		t.Errorf("Expected status 'going', got %s", stored.Status)
	}
}

func TestCreateOrUpdateRSVP_UpdateStatus(t *testing.T) {
	rsvpRepo := scene.NewInMemoryRSVPRepository()
	eventRepo := scene.NewInMemoryEventRepository()
	handlers := NewRSVPHandlers(rsvpRepo, eventRepo)

	// Create a future event
	futureTime := time.Now().Add(24 * time.Hour)
	event := &scene.Event{
		ID:            "event-1",
		SceneID:       "scene-1",
		Title:         "Test Event",
		CoarseGeohash: "dr5regw",
		StartsAt:      futureTime,
	}
	if err := eventRepo.Insert(event); err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	// Create initial RSVP with "maybe"
	initialRSVP := &scene.RSVP{
		EventID: "event-1",
		UserID:  "did:plc:user1",
		Status:  "maybe",
	}
	if err := rsvpRepo.Upsert(initialRSVP); err != nil {
		t.Fatalf("Failed to create initial RSVP: %v", err)
	}

	// Update to "going"
	reqBody := RSVPRequest{Status: "going"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/events/event-1/rsvp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := middleware.SetUserDID(req.Context(), "did:plc:user1")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.CreateOrUpdateRSVP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify status was updated
	stored, err := rsvpRepo.GetByEventAndUser("event-1", "did:plc:user1")
	if err != nil {
		t.Fatalf("Failed to get RSVP: %v", err)
	}
	if stored.Status != "going" {
		t.Errorf("Expected status 'going', got %s", stored.Status)
	}
}

func TestCreateOrUpdateRSVP_InvalidStatus(t *testing.T) {
	rsvpRepo := scene.NewInMemoryRSVPRepository()
	eventRepo := scene.NewInMemoryEventRepository()
	handlers := NewRSVPHandlers(rsvpRepo, eventRepo)

	// Create request with invalid status
	reqBody := RSVPRequest{Status: "invalid"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/events/event-1/rsvp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := middleware.SetUserDID(req.Context(), "did:plc:user1")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.CreateOrUpdateRSVP(w, req)

	// Verify error response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreateOrUpdateRSVP_PastEvent(t *testing.T) {
	rsvpRepo := scene.NewInMemoryRSVPRepository()
	eventRepo := scene.NewInMemoryEventRepository()
	handlers := NewRSVPHandlers(rsvpRepo, eventRepo)

	// Create a past event
	pastTime := time.Now().Add(-24 * time.Hour)
	event := &scene.Event{
		ID:            "event-1",
		SceneID:       "scene-1",
		Title:         "Past Event",
		CoarseGeohash: "dr5regw",
		StartsAt:      pastTime,
	}
	if err := eventRepo.Insert(event); err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	// Try to RSVP to past event
	reqBody := RSVPRequest{Status: "going"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/events/event-1/rsvp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := middleware.SetUserDID(req.Context(), "did:plc:user1")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.CreateOrUpdateRSVP(w, req)

	// Verify error response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestCreateOrUpdateRSVP_EventNotFound(t *testing.T) {
	rsvpRepo := scene.NewInMemoryRSVPRepository()
	eventRepo := scene.NewInMemoryEventRepository()
	handlers := NewRSVPHandlers(rsvpRepo, eventRepo)

	// Try to RSVP to non-existent event
	reqBody := RSVPRequest{Status: "going"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/events/non-existent/rsvp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	ctx := middleware.SetUserDID(req.Context(), "did:plc:user1")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.CreateOrUpdateRSVP(w, req)

	// Verify error response
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestCreateOrUpdateRSVP_Unauthenticated(t *testing.T) {
	rsvpRepo := scene.NewInMemoryRSVPRepository()
	eventRepo := scene.NewInMemoryEventRepository()
	handlers := NewRSVPHandlers(rsvpRepo, eventRepo)

	// Create request without user DID in context
	reqBody := RSVPRequest{Status: "going"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/events/event-1/rsvp", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handlers.CreateOrUpdateRSVP(w, req)

	// Verify error response
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestDeleteRSVP_Success(t *testing.T) {
	rsvpRepo := scene.NewInMemoryRSVPRepository()
	eventRepo := scene.NewInMemoryEventRepository()
	handlers := NewRSVPHandlers(rsvpRepo, eventRepo)

	// Create a future event
	futureTime := time.Now().Add(24 * time.Hour)
	event := &scene.Event{
		ID:            "event-1",
		SceneID:       "scene-1",
		Title:         "Test Event",
		CoarseGeohash: "dr5regw",
		StartsAt:      futureTime,
	}
	if err := eventRepo.Insert(event); err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	// Create RSVP
	rsvp := &scene.RSVP{
		EventID: "event-1",
		UserID:  "did:plc:user1",
		Status:  "going",
	}
	if err := rsvpRepo.Upsert(rsvp); err != nil {
		t.Fatalf("Failed to create RSVP: %v", err)
	}

	// Delete RSVP
	req := httptest.NewRequest("DELETE", "/events/event-1/rsvp", nil)
	ctx := middleware.SetUserDID(req.Context(), "did:plc:user1")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.DeleteRSVP(w, req)

	// Verify response
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d: %s", w.Code, w.Body.String())
	}

	// Verify RSVP was deleted
	_, err := rsvpRepo.GetByEventAndUser("event-1", "did:plc:user1")
	if err != scene.ErrRSVPNotFound {
		t.Errorf("Expected ErrRSVPNotFound, got %v", err)
	}
}

func TestDeleteRSVP_NotFound(t *testing.T) {
	rsvpRepo := scene.NewInMemoryRSVPRepository()
	eventRepo := scene.NewInMemoryEventRepository()
	handlers := NewRSVPHandlers(rsvpRepo, eventRepo)

	// Create a future event
	futureTime := time.Now().Add(24 * time.Hour)
	event := &scene.Event{
		ID:            "event-1",
		SceneID:       "scene-1",
		Title:         "Test Event",
		CoarseGeohash: "dr5regw",
		StartsAt:      futureTime,
	}
	if err := eventRepo.Insert(event); err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	// Try to delete non-existent RSVP
	req := httptest.NewRequest("DELETE", "/events/event-1/rsvp", nil)
	ctx := middleware.SetUserDID(req.Context(), "did:plc:user1")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.DeleteRSVP(w, req)

	// Verify error response
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestDeleteRSVP_PastEvent(t *testing.T) {
	rsvpRepo := scene.NewInMemoryRSVPRepository()
	eventRepo := scene.NewInMemoryEventRepository()
	handlers := NewRSVPHandlers(rsvpRepo, eventRepo)

	// Create a past event
	pastTime := time.Now().Add(-24 * time.Hour)
	event := &scene.Event{
		ID:            "event-1",
		SceneID:       "scene-1",
		Title:         "Past Event",
		CoarseGeohash: "dr5regw",
		StartsAt:      pastTime,
	}
	if err := eventRepo.Insert(event); err != nil {
		t.Fatalf("Failed to insert event: %v", err)
	}

	// Create RSVP
	rsvp := &scene.RSVP{
		EventID: "event-1",
		UserID:  "did:plc:user1",
		Status:  "going",
	}
	if err := rsvpRepo.Upsert(rsvp); err != nil {
		t.Fatalf("Failed to create RSVP: %v", err)
	}

	// Try to delete RSVP for past event
	req := httptest.NewRequest("DELETE", "/events/event-1/rsvp", nil)
	ctx := middleware.SetUserDID(req.Context(), "did:plc:user1")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.DeleteRSVP(w, req)

	// Verify error response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestDeleteRSVP_Unauthenticated(t *testing.T) {
	rsvpRepo := scene.NewInMemoryRSVPRepository()
	eventRepo := scene.NewInMemoryEventRepository()
	handlers := NewRSVPHandlers(rsvpRepo, eventRepo)

	// Try to delete without user DID in context
	req := httptest.NewRequest("DELETE", "/events/event-1/rsvp", nil)

	w := httptest.NewRecorder()
	handlers.DeleteRSVP(w, req)

	// Verify error response
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}
