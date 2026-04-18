package model

import "time"

type User struct {
	ID         int       `json:"id"`
	Provider   string    `json:"provider"`
	ProviderID string    `json:"provider_id"`
	Email      string    `json:"email"`
	Name       string    `json:"name"`
	CreatedAt  time.Time `json:"created_at"`
}
