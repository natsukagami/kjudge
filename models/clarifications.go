package models

import (
	"fmt"

	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models/verify"
	"github.com/pkg/errors"
)

// Verify verifies Clarification's content.
func (c *Clarification) Verify() error {
	if c.Response != nil && len(c.Response) == 0 {
		return verify.Errorf("field Response: cannot be empty")
	}
	if len(c.Content) > 2048 {
		return verify.Errorf("Content must be at most 2048 characters")
	}
	if len(c.Response) > 2048 {
		return verify.Errorf("Response must be at most 2048 characters")
	}
	return verify.All(map[string]error{
		"content": verify.NotNull(c.Content),
	})
}

// AdminLink returns the link to the Clarification in the Admin Panel.
func (c *Clarification) AdminLink() string {
	return fmt.Sprintf("/admin/clarifications/%d", c.ID)
}

// GetContestUserClarifications returns the clarifications of a contest for an user.
func GetContestUserClarifications(db db.DBContext, contestID int, userID string) ([]*Clarification, error) {
	var res []*Clarification
	if err := db.Select(&res, "SELECT * FROM clarifications WHERE contest_id = ? AND user_id = ?"+queryClarificationOrderBy, contestID, userID); err != nil {
		return nil, errors.WithStack(err)
	}
	return res, nil
}

// GetUnreadClarifications returns unread clarifications later than the given ID.
func GetUnreadClarifications(db db.DBContext, contestID int, userID string, sinceID int) ([]*Clarification, error) {
	var res []*Clarification
	if err := db.Select(&res, "SELECT * FROM clarifications WHERE contest_id = ? AND user_id = ? AND id > ? AND response IS NOT NULL"+queryClarificationOrderBy, contestID, userID, sinceID); err != nil {
		return nil, errors.WithStack(err)
	}
	return res, nil
}
