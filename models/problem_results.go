package models

import (
	"github.com/jmoiron/sqlx"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models/verify"
	"github.com/pkg/errors"
)

// Verify verifies the content of ProblemResult.
func (r *ProblemResult) Verify() error {
	if (r.Solved || r.Score > 0 || r.Penalty > 0) && !r.BestSubmissionID.Valid {
		return errors.New("best submission: must be there when result is not zero")
	}
	return verify.All(map[string]error{
		"Penalty":        verify.Int(r.Penalty, verify.IntMin(0)),
		"Score":          verify.Float(r.Score, verify.FloatMin(0)),
		"FailedAttempts": verify.Int(r.FailedAttempts, verify.IntMin(0)),
	})
}

// CollectProblemResults collects an user's problem results for a contest.
// The result map's key is the problem ID.
func CollectUserProblemResults(db db.DBContext, userID string, problems []*Problem) (map[int]*ProblemResult, error) {
	if len(problems) == 0 {
		return make(map[int]*ProblemResult), nil
	}
	var ps []*ProblemResult
	var IDs []int
	for _, p := range problems {
		IDs = append(IDs, p.ID)
	}
	query, params, err := sqlx.In("SELECT * FROM problem_results WHERE problem_id IN (?) AND user_id = ?", IDs, userID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err := db.Select(&ps, query, params...); err != nil {
		return nil, errors.WithStack(err)
	}
	res := make(map[int]*ProblemResult)
	for _, p := range ps {
		res[p.ProblemID] = p
	}
	return res, nil
}

func CollectContestProblemResults(db db.DBContext, problems []*Problem) ([]*ProblemResult, error) {
	if len(problems) == 0 {
		return nil, nil
	}
	var res []*ProblemResult
	var IDs []int
	for _, p := range problems {
		IDs = append(IDs, p.ID)
	}
	query, params, err := sqlx.In("SELECT * FROM problem_results WHERE problem_id IN (?)", IDs)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err := db.Select(&res, query, params...); err != nil {
		return nil, errors.WithStack(err)
	}
	return res, nil
}
