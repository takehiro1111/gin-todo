package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type PasswordManager interface {
	HashPassword(password string) (string, error)
	CheckPassword(hashedPassword, password string) error
}

type bcryptHasher struct{}

func NewBcryptHasher() PasswordManager {
	return &bcryptHasher{}
}

func (h *bcryptHasher) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("failed to generate hash password")
	}
	return string(hash), nil
}

func (h *bcryptHasher) CheckPassword(hashedPassword, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return errors.New("failed to compare hash password")
	}
	return nil
}
