package user

import (
	"errors"
	"fmt"

	"github.com/rivando-al-rasyid/vanwallet-backend/internals/models"
)

type UserManager struct {
	data []*models.User
}

func NewUserManager() *UserManager {
	return &UserManager{}
}

func (um *UserManager) AddUser(id, nama string) error {
	for _, user := range um.data {
		if user.Id == id {
			return errors.New("error: user dengan id '" + id + "' sudah terdaftar")
		}
	}
	um.data = append(um.data, &models.User{ID: id, Nama: nama})
	fmt.Printf("User '%s' dengan id '%s' berhasil ditambahkan.\n", nama, id)
	return nil
}

func (um *UserManager) GetUser(id string) (*models.User, error) {
	for _, user := range um.data {
		if user.Id == id {
			return user, nil
		}
	}
	return nil, errors.New("error: user dengan id '" + id + "' tidak ditemukan")
}
