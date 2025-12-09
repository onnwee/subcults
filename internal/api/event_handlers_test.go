package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/onnwee/subcults/internal/middleware"
	"github.com/onnwee/subcults/internal/scene"
)

// TestCreateEvent_Success tests successful event creation.
func TestCreateEvent_Success(t *testing.T) {
	eventRepo := scene.NewInMemoryEventRepository()
	sceneRepo := scene.NewInMemorySceneRepository()
	handlers := NewEventHandlers(eventRepo, sceneRepo)

	// Create a scene first
	testScene := &scene.Scene{
		ID:            uuid.New().String(),
		Name:          "Test Scene",
		OwnerDID:      "did:plc:test123",
		CoarseGeohash: "dr5regw",
		CreatedAt:     &time.Time{},
	}
	if err := sceneRepo.Insert(testScene); err != nil {
		t.Fatalf("failed to insert scene: %v", err)
	}

	startsAt := time.Now().Add(24 * time.Hour)
	endsAt := startsAt.Add(2 * time.Hour)

	reqBody := CreateEventRequest{
		SceneID:       testScene.ID,
		Title:         "Test Event",
		Description:   "A test event",
		AllowPrecise:  true,
		PrecisePoint:  &scene.Point{Lat: 40.7128, Lng: -74.0060},
		CoarseGeohash: "dr5regw",
		Tags:          []string{"test", "example"},
		StartsAt:      startsAt,
		EndsAt:        &endsAt,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Set user DID in context
	ctx := middleware.SetUserDID(req.Context(), "did:plc:test123")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.CreateEvent(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var createdEvent scene.Event
	if err := json.NewDecoder(w.Body).Decode(&createdEvent); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if createdEvent.Title != "Test Event" {
		t.Errorf("expected title 'Test Event', got %s", createdEvent.Title)
	}
	if createdEvent.SceneID != testScene.ID {
		t.Errorf("expected scene_id '%s', got %s", testScene.ID, createdEvent.SceneID)
	}
	if createdEvent.PrecisePoint == nil {
		t.Error("expected precise_point to be set")
	}
	if createdEvent.Status != "scheduled" {
		t.Errorf("expected status 'scheduled', got %s", createdEvent.Status)
	}
}

// TestCreateEvent_InvalidTimeWindow tests rejection of invalid time windows.
func TestCreateEvent_InvalidTimeWindow(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		startsAt time.Time
		endsAt   *time.Time
		wantCode int
		wantErr  string
	}{
		{
			name:     "end before start",
			startsAt: now.Add(24 * time.Hour),
			endsAt:   func() *time.Time { t := now; return &t }(),
			wantCode: http.StatusBadRequest,
			wantErr:  ErrCodeInvalidTimeRange,
		},
		{
			name:     "same time",
			startsAt: now.Add(24 * time.Hour),
			endsAt:   func() *time.Time { t := now.Add(24 * time.Hour); return &t }(),
			wantCode: http.StatusBadRequest,
			wantErr:  ErrCodeInvalidTimeRange,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventRepo := scene.NewInMemoryEventRepository()
			sceneRepo := scene.NewInMemorySceneRepository()
			handlers := NewEventHandlers(eventRepo, sceneRepo)

			// Create a scene first
			testScene := &scene.Scene{
				ID:            uuid.New().String(),
				Name:          "Test Scene",
				OwnerDID:      "did:plc:test123",
				CoarseGeohash: "dr5regw",
			}
			if err := sceneRepo.Insert(testScene); err != nil {
				t.Fatalf("failed to insert scene: %v", err)
			}

			reqBody := CreateEventRequest{
				SceneID:       testScene.ID,
				Title:         "Test Event",
				CoarseGeohash: "dr5regw",
				StartsAt:      tt.startsAt,
				EndsAt:        tt.endsAt,
			}

			body, err := json.Marshal(reqBody)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			ctx := middleware.SetUserDID(req.Context(), "did:plc:test123")
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handlers.CreateEvent(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("expected status %d, got %d: %s", tt.wantCode, w.Code, w.Body.String())
			}

			var errResp ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
				t.Fatalf("failed to decode error response: %v", err)
			}

			if errResp.Error.Code != tt.wantErr {
				t.Errorf("expected error code '%s', got '%s'", tt.wantErr, errResp.Error.Code)
			}
		})
	}
}

// TestCreateEvent_MissingCoarseGeohash tests rejection when coarse_geohash is missing.
func TestCreateEvent_MissingCoarseGeohash(t *testing.T) {
	eventRepo := scene.NewInMemoryEventRepository()
	sceneRepo := scene.NewInMemorySceneRepository()
	handlers := NewEventHandlers(eventRepo, sceneRepo)

	// Create a scene first
	testScene := &scene.Scene{
		ID:            uuid.New().String(),
		Name:          "Test Scene",
		OwnerDID:      "did:plc:test123",
		CoarseGeohash: "dr5regw",
	}
	if err := sceneRepo.Insert(testScene); err != nil {
		t.Fatalf("failed to insert scene: %v", err)
	}

	reqBody := CreateEventRequest{
		SceneID:       testScene.ID,
		Title:         "Test Event",
		CoarseGeohash: "", // Empty geohash
		StartsAt:      time.Now().Add(24 * time.Hour),
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserDID(req.Context(), "did:plc:test123")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.CreateEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error.Code != ErrCodeValidation {
		t.Errorf("expected error code '%s', got '%s'", ErrCodeValidation, errResp.Error.Code)
	}
}

// TestCreateEvent_UnauthorizedCreate tests rejection when user doesn't own the scene.
func TestCreateEvent_UnauthorizedCreate(t *testing.T) {
	eventRepo := scene.NewInMemoryEventRepository()
	sceneRepo := scene.NewInMemorySceneRepository()
	handlers := NewEventHandlers(eventRepo, sceneRepo)

	// Create a scene with different owner
	testScene := &scene.Scene{
		ID:            uuid.New().String(),
		Name:          "Test Scene",
		OwnerDID:      "did:plc:owner123",
		CoarseGeohash: "dr5regw",
	}
	if err := sceneRepo.Insert(testScene); err != nil {
		t.Fatalf("failed to insert scene: %v", err)
	}

	reqBody := CreateEventRequest{
		SceneID:       testScene.ID,
		Title:         "Test Event",
		CoarseGeohash: "dr5regw",
		StartsAt:      time.Now().Add(24 * time.Hour),
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	// Set different user DID
	ctx := middleware.SetUserDID(req.Context(), "did:plc:different123")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.CreateEvent(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d: %s", w.Code, w.Body.String())
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error.Code != ErrCodeForbidden {
		t.Errorf("expected error code '%s', got '%s'", ErrCodeForbidden, errResp.Error.Code)
	}
}

// TestCreateEvent_PrivacyEnforcement tests that privacy is enforced on creation.
func TestCreateEvent_PrivacyEnforcement(t *testing.T) {
	eventRepo := scene.NewInMemoryEventRepository()
	sceneRepo := scene.NewInMemorySceneRepository()
	handlers := NewEventHandlers(eventRepo, sceneRepo)

	// Create a scene first
	testScene := &scene.Scene{
		ID:            uuid.New().String(),
		Name:          "Test Scene",
		OwnerDID:      "did:plc:test123",
		CoarseGeohash: "dr5regw",
	}
	if err := sceneRepo.Insert(testScene); err != nil {
		t.Fatalf("failed to insert scene: %v", err)
	}

	reqBody := CreateEventRequest{
		SceneID:       testScene.ID,
		Title:         "Private Event",
		AllowPrecise:  false, // Privacy not consented
		PrecisePoint:  &scene.Point{Lat: 40.7128, Lng: -74.0060}, // Should be cleared
		CoarseGeohash: "dr5regw",
		StartsAt:      time.Now().Add(24 * time.Hour),
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserDID(req.Context(), "did:plc:test123")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.CreateEvent(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var createdEvent scene.Event
	if err := json.NewDecoder(w.Body).Decode(&createdEvent); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if createdEvent.PrecisePoint != nil {
		t.Error("expected precise_point to be nil when allow_precise=false")
	}
}

// TestCreateEvent_TitleValidation tests title length validation.
func TestCreateEvent_TitleValidation(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		wantCode int
		wantErr  string
	}{
		{
			name:     "too short",
			title:    "ab",
			wantCode: http.StatusBadRequest,
			wantErr:  ErrCodeValidation,
		},
		{
			name:     "too long",
			title:    strings.Repeat("a", 81),
			wantCode: http.StatusBadRequest,
			wantErr:  ErrCodeValidation,
		},
		{
			name:     "valid minimum",
			title:    "abc",
			wantCode: http.StatusCreated,
			wantErr:  "",
		},
		{
			name:     "valid maximum",
			title:    strings.Repeat("a", 80),
			wantCode: http.StatusCreated,
			wantErr:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventRepo := scene.NewInMemoryEventRepository()
			sceneRepo := scene.NewInMemorySceneRepository()
			handlers := NewEventHandlers(eventRepo, sceneRepo)

			// Create a scene first
			testScene := &scene.Scene{
				ID:            uuid.New().String(),
				Name:          "Test Scene",
				OwnerDID:      "did:plc:test123",
				CoarseGeohash: "dr5regw",
			}
			if err := sceneRepo.Insert(testScene); err != nil {
				t.Fatalf("failed to insert scene: %v", err)
			}

			reqBody := CreateEventRequest{
				SceneID:       testScene.ID,
				Title:         tt.title,
				CoarseGeohash: "dr5regw",
				StartsAt:      time.Now().Add(24 * time.Hour),
			}

			body, err := json.Marshal(reqBody)
			if err != nil {
				t.Fatalf("failed to marshal request: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			ctx := middleware.SetUserDID(req.Context(), "did:plc:test123")
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handlers.CreateEvent(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("expected status %d, got %d: %s", tt.wantCode, w.Code, w.Body.String())
			}

			if tt.wantErr != "" {
				var errResp ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}

				if errResp.Error.Code != tt.wantErr {
					t.Errorf("expected error code '%s', got '%s'", tt.wantErr, errResp.Error.Code)
				}
			}
		})
	}
}

// TestUpdateEvent_Success tests successful event update.
func TestUpdateEvent_Success(t *testing.T) {
	eventRepo := scene.NewInMemoryEventRepository()
	sceneRepo := scene.NewInMemorySceneRepository()
	handlers := NewEventHandlers(eventRepo, sceneRepo)

	// Create a scene first
	testScene := &scene.Scene{
		ID:            uuid.New().String(),
		Name:          "Test Scene",
		OwnerDID:      "did:plc:test123",
		CoarseGeohash: "dr5regw",
	}
	if err := sceneRepo.Insert(testScene); err != nil {
		t.Fatalf("failed to insert scene: %v", err)
	}

	// Create an event
	now := time.Now()
	startsAt := now.Add(24 * time.Hour)
	existingEvent := &scene.Event{
		ID:            uuid.New().String(),
		SceneID:       testScene.ID,
		Title:         "Original Title",
		CoarseGeohash: "dr5regw",
		StartsAt:      startsAt,
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}
	if err := eventRepo.Insert(existingEvent); err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	newTitle := "Updated Title"
	newDesc := "Updated description"
	reqBody := UpdateEventRequest{
		Title:       &newTitle,
		Description: &newDesc,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, "/events/"+existingEvent.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserDID(req.Context(), "did:plc:test123")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.UpdateEvent(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var updatedEvent scene.Event
	if err := json.NewDecoder(w.Body).Decode(&updatedEvent); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if updatedEvent.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got %s", updatedEvent.Title)
	}
	if updatedEvent.Description != "Updated description" {
		t.Errorf("expected description 'Updated description', got %s", updatedEvent.Description)
	}
}

// TestUpdateEvent_CannotUpdatePastEvent tests that past events cannot have time updated.
func TestUpdateEvent_CannotUpdatePastEvent(t *testing.T) {
	eventRepo := scene.NewInMemoryEventRepository()
	sceneRepo := scene.NewInMemorySceneRepository()
	handlers := NewEventHandlers(eventRepo, sceneRepo)

	// Create a scene first
	testScene := &scene.Scene{
		ID:            uuid.New().String(),
		Name:          "Test Scene",
		OwnerDID:      "did:plc:test123",
		CoarseGeohash: "dr5regw",
	}
	if err := sceneRepo.Insert(testScene); err != nil {
		t.Fatalf("failed to insert scene: %v", err)
	}

	// Create a past event
	now := time.Now()
	pastTime := now.Add(-24 * time.Hour)
	existingEvent := &scene.Event{
		ID:            uuid.New().String(),
		SceneID:       testScene.ID,
		Title:         "Past Event",
		CoarseGeohash: "dr5regw",
		StartsAt:      pastTime,
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}
	if err := eventRepo.Insert(existingEvent); err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	newStartsAt := now.Add(48 * time.Hour)
	reqBody := UpdateEventRequest{
		StartsAt: &newStartsAt,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, "/events/"+existingEvent.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserDID(req.Context(), "did:plc:test123")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.UpdateEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error.Code != ErrCodeValidation {
		t.Errorf("expected error code '%s', got '%s'", ErrCodeValidation, errResp.Error.Code)
	}
}

// TestUpdateEvent_TimeWindowValidation tests time window validation on update.
func TestUpdateEvent_TimeWindowValidation(t *testing.T) {
	eventRepo := scene.NewInMemoryEventRepository()
	sceneRepo := scene.NewInMemorySceneRepository()
	handlers := NewEventHandlers(eventRepo, sceneRepo)

	// Create a scene first
	testScene := &scene.Scene{
		ID:            uuid.New().String(),
		Name:          "Test Scene",
		OwnerDID:      "did:plc:test123",
		CoarseGeohash: "dr5regw",
	}
	if err := sceneRepo.Insert(testScene); err != nil {
		t.Fatalf("failed to insert scene: %v", err)
	}

	// Create a future event
	now := time.Now()
	startsAt := now.Add(24 * time.Hour)
	endsAt := startsAt.Add(2 * time.Hour)
	existingEvent := &scene.Event{
		ID:            uuid.New().String(),
		SceneID:       testScene.ID,
		Title:         "Future Event",
		CoarseGeohash: "dr5regw",
		StartsAt:      startsAt,
		EndsAt:        &endsAt,
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}
	if err := eventRepo.Insert(existingEvent); err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	// Try to set end time before start time
	newEndsAt := startsAt.Add(-1 * time.Hour)
	reqBody := UpdateEventRequest{
		EndsAt: &newEndsAt,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, "/events/"+existingEvent.ID, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := middleware.SetUserDID(req.Context(), "did:plc:test123")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.UpdateEvent(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error.Code != ErrCodeInvalidTimeRange {
		t.Errorf("expected error code '%s', got '%s'", ErrCodeInvalidTimeRange, errResp.Error.Code)
	}
}

// TestGetEvent_Success tests successful event retrieval.
func TestGetEvent_Success(t *testing.T) {
	eventRepo := scene.NewInMemoryEventRepository()
	sceneRepo := scene.NewInMemorySceneRepository()
	handlers := NewEventHandlers(eventRepo, sceneRepo)

	// Create an event
	now := time.Now()
	testEvent := &scene.Event{
		ID:            uuid.New().String(),
		SceneID:       uuid.New().String(),
		Title:         "Test Event",
		CoarseGeohash: "dr5regw",
		StartsAt:      now.Add(24 * time.Hour),
		AllowPrecise:  true,
		PrecisePoint:  &scene.Point{Lat: 40.7128, Lng: -74.0060},
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}
	if err := eventRepo.Insert(testEvent); err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/events/"+testEvent.ID, nil)
	w := httptest.NewRecorder()

	handlers.GetEvent(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var foundEvent scene.Event
	if err := json.NewDecoder(w.Body).Decode(&foundEvent); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if foundEvent.ID != testEvent.ID {
		t.Errorf("expected ID '%s', got '%s'", testEvent.ID, foundEvent.ID)
	}
	if foundEvent.Title != "Test Event" {
		t.Errorf("expected title 'Test Event', got %s", foundEvent.Title)
	}
	if foundEvent.PrecisePoint == nil {
		t.Error("expected precise_point to be set")
	}
}

// TestGetEvent_NotFound tests 404 when event doesn't exist.
func TestGetEvent_NotFound(t *testing.T) {
	eventRepo := scene.NewInMemoryEventRepository()
	sceneRepo := scene.NewInMemorySceneRepository()
	handlers := NewEventHandlers(eventRepo, sceneRepo)

	req := httptest.NewRequest(http.MethodGet, "/events/"+uuid.New().String(), nil)
	w := httptest.NewRecorder()

	handlers.GetEvent(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error.Code != ErrCodeNotFound {
		t.Errorf("expected error code '%s', got '%s'", ErrCodeNotFound, errResp.Error.Code)
	}
}

// TestGetEvent_PrivacyEnforcement tests that precise_point is hidden when not allowed.
func TestGetEvent_PrivacyEnforcement(t *testing.T) {
	eventRepo := scene.NewInMemoryEventRepository()
	sceneRepo := scene.NewInMemorySceneRepository()
	handlers := NewEventHandlers(eventRepo, sceneRepo)

	// Create an event without precise location consent
	now := time.Now()
	testEvent := &scene.Event{
		ID:            uuid.New().String(),
		SceneID:       uuid.New().String(),
		Title:         "Private Event",
		CoarseGeohash: "dr5regw",
		StartsAt:      now.Add(24 * time.Hour),
		AllowPrecise:  false,
		PrecisePoint:  &scene.Point{Lat: 40.7128, Lng: -74.0060}, // Should be cleared
		CreatedAt:     &now,
		UpdatedAt:     &now,
	}
	if err := eventRepo.Insert(testEvent); err != nil {
		t.Fatalf("failed to insert event: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/events/"+testEvent.ID, nil)
	w := httptest.NewRecorder()

	handlers.GetEvent(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var foundEvent scene.Event
	if err := json.NewDecoder(w.Body).Decode(&foundEvent); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if foundEvent.PrecisePoint != nil {
		t.Error("expected precise_point to be nil when allow_precise=false")
	}
}
