package models

import (
	"github.com/jmoiron/sqlx"
	"github.com/natsukagami/kjudge/db"
	"github.com/pkg/errors"
)

// This file handles the DB interactions related to rejudging a set of submissions.
// Since this can be costly, we would want to avoid the N+1 thing.

// Quickly create a bunch of score jobs.
func batchScoreJobs(subIDs ...int) []*Job {
	var jobs []*Job
	for _, id := range subIDs {
		jobs = append(jobs, NewJobScore(id))
	}
	return jobs
}

// Remove the submission's `score`, `penalty` and `verdict`.
func resetScore(db db.DBContext, subIDs ...int) error {
	query, params, err := sqlx.In(`UPDATE submissions SET score = NULL, penalty = NULL, verdict = "..." WHERE id IN (?)`, subIDs)
	if err != nil {
		return errors.WithStack(err)
	}
	if _, err := db.Exec(query, params...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// RejudgeScore re-scores all submissions given.
func RejudgeScore(db db.DBContext, subIDs ...int) error {
	if len(subIDs) == 0 {
		return nil
	}
	if err := resetScore(db, subIDs...); err != nil {
		return err
	}
	if err := BatchInsertJobs(db, batchScoreJobs(subIDs...)...); err != nil {
		return err
	}
	return nil
}

// Remove all test results.
func resetTests(db db.DBContext, subIDs ...int) error {
	query, params, err := sqlx.In("DELETE FROM test_results WHERE submission_id IN (?)", subIDs)
	if err != nil {
		return errors.WithStack(err)
	}
	if _, err := db.Exec(query, params...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// RejudgeRun re-runs all tests for the submissions given.
func RejudgeRun(db db.DBContext, subIDs ...int) error {
	if len(subIDs) == 0 {
		return nil
	}
	if err := resetTests(db, subIDs...); err != nil {
		return err
	}
	return RejudgeScore(db, subIDs...)
}

// Reset the compilation output.
func resetCompileOutput(db db.DBContext, subIDs ...int) error {
	query, params, err := sqlx.In(`UPDATE submissions SET compiler_output = NULL, compiled_source = NULL WHERE id IN (?)`, subIDs)
	if err != nil {
		return errors.WithStack(err)
	}
	if _, err := db.Exec(query, params...); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

// RejudgeCompile re-compiles all submissions given.
func RejudgeCompile(db db.DBContext, subIDs ...int) error {
	if len(subIDs) == 0 {
		return nil
	}
	if err := resetCompileOutput(db, subIDs...); err != nil {
		return err
	}
	return RejudgeRun(db, subIDs...)
}
