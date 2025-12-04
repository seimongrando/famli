package models

// Settings stores configurable options for a user.
type Settings struct {
	UserID                   string `json:"-"`
	EmergencyProtocolEnabled bool   `json:"emergency_protocol_enabled"`
}
