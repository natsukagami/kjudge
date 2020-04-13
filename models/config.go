package models

import (
	"crypto/rand"

	"git.nkagami.me/natsukagami/kjudge/db"
	"github.com/pkg/errors"
)

// Config is the configuation of the server.
type Config struct {
	SessionKey              []byte `db:"session_key"`
	EnableRegistration      bool   `db:"enable_registration"`
	EnableUserCustomization bool   `db:"enable_user_customization"`
}

// GenerateConfig generates a random configuration.
func GenerateConfig() (*Config, error) {
	c := Config{
		SessionKey:              make([]byte, 64),
		EnableRegistration:      true,
		EnableUserCustomization: true,
	}
	if _, err := rand.Read(c.SessionKey); err != nil {
		return nil, errors.WithStack(err)
	}
	return &c, nil
}

// GetConfig gets the configuration of the server.
func GetConfig(db db.DBContext) (*Config, error) {
	var c Config
	if err := db.Get(&c, "SELECT * FROM config"); err != nil {
		return nil, errors.WithStack(err)
	}
	return &c, nil
}

// Write writes to the database.
// It needs a root DB because we need a transaction.
func (c *Config) Write(db *db.DB) error {
	if err := c.Verify(); err != nil {
		return err
	}

	tx, err := db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer tx.Rollback()

	res, err := tx.Exec("UPDATE config SET session_key = ?, enable_registration = ?, enable_user_customization = ?", c.SessionKey, c.EnableRegistration, c.EnableUserCustomization)
	if err != nil {
		return errors.WithStack(err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return errors.WithStack(err)
	}
	if rowsAffected == 0 {
		// Gotta INSERT something I guess
		_, err := tx.Exec("INSERT INTO config(session_key, enable_registration, enable_user_customization) VALUES (?, ?, ?)", c.SessionKey, c.EnableRegistration, c.EnableUserCustomization)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return errors.WithStack(tx.Commit())
}

// Verify verifies a Config's content.
func (c *Config) Verify() error {
	if c.SessionKey == nil {
		return errors.New("keys must not be null")
	}
	if len(c.SessionKey) != 64 {
		return errors.New("keys must have length 64")
	}
	if c.EnableRegistration && !c.EnableUserCustomization {
		// If registration is enabled, so must user customization.
		return errors.New("cannot disable user customization when registration is enabled")
	}
	return nil
}
