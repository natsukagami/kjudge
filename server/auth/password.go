// Package auth implements the authentication functions for the server.
package auth

import (
	"crypto/rand"
	"fmt"

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
	if len(raw) < 6 {
		return nil, errors.New("Password too short")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(raw), hashCost)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return hashed, nil
}

// GeneratePassword generates random hex passwords of length 8.
func GeneratePassword(count int) ([]string, error) {
	b := make([]byte, 4*count)
	if _, err := rand.Read(b); err != nil {
		return nil, errors.WithStack(err)
	}
	var res []string
	for i := 0; i < count; i++ {
		pass := b[4*i : 4*(i+1)]
		res = append(res, fmt.Sprintf("%x", pass))
	}
	return res, nil
}
