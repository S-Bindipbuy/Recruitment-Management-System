package models

import "time"

type AuditLog struct {
	ID         int64     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	Action     string    `json:"action" db:"action"`
	EntityType string    `json:"entity_type" db:"entity_type"`
	EntityID   int64     `json:"entity_id" db:"entity_id"`
	OldValues  string    `json:"old_values" db:"old_values"`
	NewValues  string    `json:"new_values" db:"new_values"`
	IPAddress  string    `json:"ip_address" db:"ip_address"`
	UserAgent  string    `json:"user_agent" db:"user_agent"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
