package models

// User represents a registered user in the system.
// It holds the core identity fields used across the application.
type User struct {
	// Id is the unique identifier for each user.
	Id string

	// Name is the full name of the user.
	Name string
}
