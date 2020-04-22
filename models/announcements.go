package models

import (
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models/verify"
	"github.com/pkg/errors"
)

// Verify verifies Announcement's content.
func (a *Announcement) Verify() error {
	if len(a.Content) > 2048 {
		return verify.Errorf("Content must be at most 2048 characters")
	}
	return verify.All(map[string]error{
		"content": verify.NotNull(a.Content),
	})
}

// GetUnreadAnnouncements returns the list of announcements later than the given id.
func GetUnreadAnnouncements(db db.DBContext, contestID int, sinceID int) ([]*Announcement, error) {
	var res []*Announcement
	if err := db.Select(&res, "SELECT * FROM announcements WHERE contest_id = ? AND id > ?"+queryAnnouncementOrderBy, contestID, sinceID); err != nil {
		return nil, errors.WithStack(err)
	}
	return res, nil
}
