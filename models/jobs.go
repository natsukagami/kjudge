package models

import (
	"database/sql"
	"strings"

	"git.nkagami.me/natsukagami/kjudge/db"
	"github.com/pkg/errors"
)

// JobType determines the type of the job.
// This can be:
// - Compile: highest priority. Compiles a submission into executable bytecode.
// - Test: run a test.
// - Score: recalculate the score.
type JobType string

// Possible values of JobType.
const (
	JobTypeCompile JobType = "compile"
	JobTypeRun     JobType = "run"
	JobTypeScore   JobType = "score"
)

const (
	compilePriority = 3
	runPriority     = 2
	scorePriority   = 1
)

// NewJobCompile creates a new Compile job.
func NewJobCompile(subID int) *Job {
	return &Job{
		Priority:     compilePriority,
		Type:         JobTypeCompile,
		SubmissionID: sql.NullInt64{Int64: int64(subID), Valid: true},
	}
}

// NewJobRun creates a new Run job.
func NewJobRun(subID int, testID int) *Job {
	return &Job{
		Priority:     runPriority,
		Type:         JobTypeRun,
		SubmissionID: sql.NullInt64{Int64: int64(subID), Valid: true},
		TestID:       sql.NullInt64{Int64: int64(testID), Valid: true},
	}
}

// NewJobScore creates a new Score job.
func NewJobScore(userID int, problemID int) *Job {
	return &Job{
		Priority:  scorePriority,
		Type:      JobTypeScore,
		UserID:    sql.NullInt64{Int64: int64(userID), Valid: true},
		ProblemID: sql.NullInt64{Int64: int64(problemID), Valid: true},
	}
}

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

// FirstJob returns the first job that needs to be done.
func FirstJob(db db.DBContext) (*Job, error) {
	var j *Job
	if err := db.Get(j, "SELECT * FROM jobs ORDER BY priority DESC, id ASC LIMIT 1"); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return j, nil
}

// BatchInsertJobs try to insert all given jobs.
func BatchInsertJobs(db db.DBContext, jobs ...*Job) error {
	if len(jobs) == 0 {
		return nil // No inserts needed
	}
	rowMarks := "(?, ?, ?, ?, ?, ?)"
	command := strings.Builder{}
	command.WriteString("INSERT INTO jobs(priority, problem_id, submission_id, test_id, type, user_id) VALUES ")
	var values []interface{}
	for id, r := range jobs {
		if id > 0 {
			command.WriteString(", ")
		}
		command.WriteString(rowMarks)
		values = append(values, r.Priority, r.ProblemID, r.SubmissionID, r.TestID, r.Type, r.UserID)
	}
	res, err := db.Exec(command.String(), values...)
	if err != nil {
		return errors.WithStack(err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return errors.WithStack(err)
	}
	for i, r := range jobs {
		r.ID = int(id) - len(jobs) + 1 + i
	}
	return nil
}
