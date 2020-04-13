package models

import (
	"strings"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models/verify"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// Verify verifies an User's contents.
func (r *User) Verify() error {
	return verify.All(map[string]error{
		"ID":           verify.Names(r.ID),
		"DisplayName":  verify.Names(r.DisplayName),
		"Organization": verify.StringEmptyOr(verify.StringMaxLength(64))(r.Organization),
	})
}

// BatchAddUsers adds users in batch.
func BatchAddUsers(db db.DBContext, reset bool, users ...*User) error {
	if reset {
		if _, err := db.Exec("DELETE FROM users"); err != nil {
			return errors.WithStack(err)
		}
	}
	if len(users) == 0 {
		return nil
	}
	var usernames []string
	for _, u := range users {
		if err := u.Verify(); err != nil {
			return errors.Wrapf(err, "user %s", u.ID)
		}
		usernames = append(usernames, u.ID)
	}
	// Query for user collisions
	{
		query, args, err := sqlx.In("SELECT id FROM users WHERE id IN (?)", usernames)
		if err != nil {
			return errors.WithStack(err)
		}
		var items []struct {
			ID string `sql:"id"`
		}
		if err := db.Select(&items, query, args...); err != nil {
			return errors.WithStack(err)
		}
		if len(items) > 0 {
			var usernames []string
			for _, item := range items {
				usernames = append(usernames, item.ID)
			}
			return verify.Errorf("The following usernames already existed: %s", strings.Join(usernames, ", "))
		}
	}
	// Build the huge query
	const field = "(?, ?, ?, ?, ?)"
	var fields []string
	var args []interface{}
	for _, u := range users {
		fields = append(fields, field)
		args = append(args, u.ID, u.DisplayName, u.Organization, u.Password, u.Hidden)
	}
	if _, err := db.Exec("INSERT INTO users(id, display_name, organization, password, hidden) VALUES "+strings.Join(fields, ", "), args...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}
