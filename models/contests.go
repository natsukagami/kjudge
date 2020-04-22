package models

import (
	"fmt"

	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models/verify"
	"github.com/pkg/errors"
)

// ContestType is the enum representing the contest type.
// The contest type determines ONLY how the scoreboard is rendered.
type ContestType string

const (
	// The contestants are sorted by number of "solved" problems, then by total penalty.
	ContestTypeUnweighted ContestType = "unweighted"
	// Problems each have different scores, and penalty only serves as tiebreakers.
	ContestTypeWeighted ContestType = "weighted"
)

// ScoreboardViewStatus is the enum representing the scoreboard view status.
// scoreboard view status type determines how the scoreboard is accessed.
type ScoreboardViewStatus string

const (
	// This allowes everyone to see the scoreboard
	ScoreboardViewStatusPublic ScoreboardViewStatus = "public"
	// This only allowes registered users to see the scoreboard
	ScoreboardViewStatusUser ScoreboardViewStatus = "user"
	// There is no scoreboard rendered during the contest
	ScoreboardViewStatusNoScoreboard ScoreboardViewStatus = "no_scoreboard"
)

// Verify tries to verify the values of a Contest struct.
func (c *Contest) Verify() error {
	if err := verify.Names(c.Name); err != nil {
		return errors.WithMessage(err, "name: ")
	}
	if !c.StartTime.Before(c.EndTime) {
		return errors.New("start time: must be before end time")
	}
	if c.ContestType != ContestTypeUnweighted &&
		c.ContestType != ContestTypeWeighted {
		return errors.New("contest type: invalid value")
	}
	return nil
}

// Link returns the HTTP link to the contest.
func (c *Contest) Link() string {
	return fmt.Sprintf("/contests/%d", c.ID)
}

// AdminLink returns the link to the contest in the Admin Panel.
func (c *Contest) AdminLink() string {
	return fmt.Sprintf("/admin/contests/%d", c.ID)
}

// GetContestsUnfinished gets a list of contests that are unfinished (upcoming or pending).
func GetContestsUnfinished(db db.DBContext) ([]*Contest, error) {
	var res []*Contest
	if err := db.Select(&res, "SELECT * FROM contests WHERE datetime(end_time) > datetime('now')"+queryContestOrderBy); err != nil {
		return nil, errors.WithStack(err)
	}
	return res, nil
}

// GetContestsFinished gets a list of contests that are finished (past contests).
func GetContestsFinished(db db.DBContext) ([]*Contest, error) {
	var res []*Contest
	if err := db.Select(&res, "SELECT * FROM contests WHERE datetime(end_time) <= datetime('now')"+queryContestOrderBy); err != nil {
		return nil, errors.WithStack(err)
	}
	return res, nil
}

// GetContests returns a list of all contests.
func GetContests(db db.DBContext) ([]*Contest, error) {
	var res []*Contest
	if err := db.Select(&res, "SELECT * FROM contests ORDER BY id DESC"); err != nil {
		return nil, errors.WithStack(err)
	}
	return res, nil
}

// GetNearestOngoingContest return the nearest ongoing contest
func GetNearestOngoingContest(db db.DBContext) (*Contest, error) {
	var res Contest
	if err := db.Get(&res, "SELECT * FROM contests WHERE datetime(end_time) > datetime('now') AND date(start_time) <= datetime('now')"+queryContestOrderBy+" LIMIT 1"); err != nil {
		return nil, errors.WithStack(err)
	}
	return &res, nil
}
