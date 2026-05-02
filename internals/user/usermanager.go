package user

import (
	"errors"
	"fmt"

	"github.com/rivando-al-rasyid/vanwallet-backend/internals/models"
)

// UserManager holds a collection of User model pointers.
// It acts as an in-memory store for managing user data.
type UserManager struct {
	data []*models.User
}

// NewUserManager initializes and returns a new empty UserManager instance.
// This should be called before performing any user operations.
func NewUserManager() *UserManager {
	return &UserManager{}
}

// AddUser creates a new user with the given id and name, then appends it to the store.
// Returns an error if a user with the same id already exists.
func (um *UserManager) AddUser(id, nama string) error {
	// Check for duplicate user id before adding
	for _, user := range um.data {
		if user.Id == id {
			return errors.New("error: user with id '" + id + "' is already registered")
		}
	}

	// Append the new user to the in-memory slice
	um.data = append(um.data, &models.User{Id: id, Name: nama})
	fmt.Printf("User '%s' with id '%s' has been successfully added.\n", nama, id)
	return nil
}

// GetUser searches for a user by their id and returns a pointer to the User model.
// Returns an error if no user with the given id is found.
func (um *UserManager) GetUser(id string) (*models.User, error) {
	// Iterate through the store to find a matching user id
	for _, user := range um.data {
		if user.Id == id {
			return user, nil
		}
	}

	return nil, errors.New("error: user with id '" + id + "' not found")
}
