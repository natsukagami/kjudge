// Package auth implements the authentication functions for the server.
package auth

import (
	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

const hashCost = bcrypt.DefaultCost

// CheckPassword returns whether the raw password matches the hashed password.
func CheckPassword(raw, hashed string) (bool, error) {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), []byte(raw))
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false, nil
	} else if err != nil {
		return false, errors.WithStack(err)
	}
	return true, nil
}

// PasswordHash hashes a raw string into a hashed password.
func PasswordHash(raw string) ([]byte, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), hashCost)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return hashed, nil
}
