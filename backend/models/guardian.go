package models

// Guardian represents a trusted contact.
type Guardian struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	UserID string `json:"user_id"`
}
