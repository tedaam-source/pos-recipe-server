package entity

import "errors"

// User represents a user in the system.
type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

var (
	ErrUserNotFound = errors.New("user not found")
)
