package models

import (
	"git.nkagami.me/natsukagami/kjudge/models/verify"
	"github.com/pkg/errors"
)

// Verify verifies the content of ProblemResult.
func (r *ProblemResult) Verify() error {
	if (r.Solved || r.Score > 0 || r.Penalty > 0) && !r.BestSubmissionID.Valid {
		return errors.New("best submission: must be there when result is not zero")
	}
	return verify.All(map[string]error{
		"Penalty": verify.Float(r.Penalty, verify.FloatMin(0)),
		"Score":   verify.Float(r.Score, verify.FloatMin(0)),
	})
}
