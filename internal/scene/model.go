// Package scene provides models and repository for managing scenes and events
// with location privacy controls.
package scene

// Point represents a geographic coordinate with latitude and longitude.
type Point struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Palette represents the color scheme for a scene's visual identity.
type Palette struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
}

// Scene represents a subcultural scene with optional precise location data.
// The precise_point field is only persisted when allow_precise consent is true.
type Scene struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	AllowPrecise  bool     `json:"allow_precise"`
	PrecisePoint  *Point   `json:"precise_point,omitempty"`
	CoarseGeohash string   `json:"coarse_geohash"`       // Required for privacy-conscious discovery
	Tags          []string `json:"tags,omitempty"`       // Categorization tags
	Visibility    string   `json:"visibility,omitempty"` // public, private, unlisted
	Palette       *Palette `json:"palette,omitempty"`    // Color scheme
	OwnerUserID   *string  `json:"owner_user_id,omitempty"` // FK to users table
	
	// AT Protocol record tracking
	RecordDID  *string `json:"record_did,omitempty"`
	RecordRKey *string `json:"record_rkey,omitempty"`
}

// Event represents an event within a scene with optional precise location data.
// The precise_point field is only persisted when allow_precise consent is true.
type Event struct {
	ID           string `json:"id"`
	SceneID      string `json:"scene_id"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	AllowPrecise bool   `json:"allow_precise"`
	PrecisePoint *Point `json:"precise_point,omitempty"`
	
	// AT Protocol record tracking
	RecordDID  *string `json:"record_did,omitempty"`
	RecordRKey *string `json:"record_rkey,omitempty"`
}

// EnforceLocationConsent clears PrecisePoint if AllowPrecise is false.
// This ensures that precise location data is never stored without consent.
// Returns the scene for chaining.
func (s *Scene) EnforceLocationConsent() *Scene {
	if !s.AllowPrecise {
		s.PrecisePoint = nil
	}
	return s
}

// EnforceLocationConsent clears PrecisePoint if AllowPrecise is false.
// This ensures that precise location data is never stored without consent.
// Returns the event for chaining.
func (e *Event) EnforceLocationConsent() *Event {
	if !e.AllowPrecise {
		e.PrecisePoint = nil
	}
	return e
}
