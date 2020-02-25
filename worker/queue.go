package worker

import (
	"log"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models"
	"github.com/mattn/go-sqlite3"
	"github.com/pkg/errors"
)

// Queue implements a queue that runs each job one by one.
type Queue struct {
	DB      *db.DB
	Sandbox Sandbox
}

// Start starts the queue. It is blocking, so might wanna "go run" it.
func (q *Queue) Start() {
	// Register the update callback
	toUpdate := q.startHook()

	for {
		// Get the newest job
		job, err := models.FirstJob(q.DB)
		if err != nil {
			log.Printf("[WORKER] Fetching job failed: %+v\n", err)
			continue
		}
		if job == nil {
			// Wait for at least one toUpdate before continuing
			<-toUpdate
			continue
		}

		if err := q.HandleJob(job); err != nil {
			log.Printf("[WORKER] Handling job failed: %+v\n", err)
		}
		job.Delete(q.DB)
	}
}

// HandleJob dispatches a job.
func (q *Queue) HandleJob(job *models.Job) error {
	// Start a job with a context and submission
	tx, err := q.DB.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer tx.Rollback()

	sub, err := models.GetSubmission(tx, job.SubmissionID)
	if err != nil {
		return err
	}
	problem, err := models.GetProblem(tx, sub.ProblemID)
	if err != nil {
		return err
	}
	switch job.Type {
	case models.JobTypeCompile:
		if _, err := Compile(&CompileContext{DB: tx, Sub: sub, Problem: problem}); err != nil {
			return err
		}
	case models.JobTypeRun:
		test, err := models.GetTest(tx, int(job.TestID.Int64))
		if err != nil {
			return err
		}
		tg, err := models.GetTestGroup(tx, test.TestGroupID)
		if err != nil {
			return err
		}
		if err := Run(q.Sandbox, &RunContext{
			DB: tx, Sub: sub, Problem: problem, TestGroup: tg, Test: test}); err != nil {
			return err
		}
	case models.JobTypeScore:
		contest, err := models.GetContest(tx, problem.ContestID)
		if err != nil {
			return err
		}
		if err := Score(&ScoreContext{DB: tx, Sub: sub, Problem: problem, Contest: contest}); err != nil {
			return err
		}
	}
	return errors.WithStack(tx.Commit())
}

// Starts a hook to be announced everytime jobs is inserted.
func (q *Queue) startHook() <-chan struct{} {
	toUpdate := make(chan struct{})
	q.DB.PersistentConn.RegisterUpdateHook(func(typ int, db, table string, rowID int64) {
		if typ == sqlite3.SQLITE_INSERT && table == "jobs" {
			toUpdate <- struct{}{}
		}
	})
	return toUpdate
}
