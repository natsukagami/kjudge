package models

import "github.com/pkg/errors"

// JobType determines the type of the job.
// This can be:
// - Compile: highest priority. Compiles a submission into executable bytecode.
// - Test: run a test.
// - Score: recalculate the score.
type JobType string

// Possible values of JobType.
const (
	JobTypeCompile JobType = "compile"
	JobTypeTest    JobType = "test"
	JobTypeScore   JobType = "score"
)

// Verify verifies whether a job is a legit job.
func (r *Job) Verify() error {
	switch r.Type {
	case JobTypeCompile:
		if !r.SubmissionID.Valid {
			return errors.New("compile submission_id: missing")
		}
	case JobTypeTest:
		if !r.SubmissionID.Valid {
			return errors.New("test submission_id: missing")
		}
		if !r.TestID.Valid {
			return errors.New("test test_id: missing")
		}
	case JobTypeScore:
		if !r.ProblemID.Valid {
			return errors.New("score problem_id: missing")
		}
		if !r.UserID.Valid {
			return errors.New("score user_id: missing")
		}
	default:
		return errors.New("type: invalid value")
	}
	return nil
}
