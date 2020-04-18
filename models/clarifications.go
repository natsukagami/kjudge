package models

import (
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models/verify"
	"github.com/pkg/errors"
)

// Verify verifies Clarification's content.
func (c *Clarification) Verify() error {
	if c.Response != nil && len(c.Response) == 0 {
		return verify.Errorf("field Response: cannot be empty")
	}
	return verify.All(map[string]error{
		"content": verify.NotNull(c.Content),
	})
}

// GetUnreadClarifications returns unread clarifications later than the given ID.
func GetUnreadClarifications(db db.DBContext, contestID int, userID int, sinceID int) ([]*Clarification, error) {
	var res []*Clarification
	if err := db.Select(&res, "SELECT * FROM clarifications WHERE contest_id = ? AND user_id = ? AND id > ? AND response IS NOT NULL"+queryClarificationOrderBy, contestID, userID, sinceID); err != nil {
		return nil, errors.WithStack(err)
	}
	return res, nil
}
