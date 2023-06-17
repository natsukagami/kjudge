package queue

import (
	"log"
	"time"

	"github.com/mattn/go-sqlite3"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/worker"
	"github.com/natsukagami/kjudge/worker/sandbox"
	"github.com/pkg/errors"
)

// Queue implements a queue that runs each job one by one.
type Queue struct {
	DB       *db.DB
	Sandbox  sandbox.Runner
	Settings Settings
}

func NewQueue(db *db.DB, sandbox sandbox.Runner, options ...Option) Queue {
	setting := DefaultSettings
	for _, option := range options {
		setting = option(setting)
	}
	return Queue{DB: db, Sandbox: sandbox, Settings: setting}
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
		_ = job.Delete(q.DB)
	}
}

// Run starts the queue, solves all pending jobs, then returns
func (q *Queue) Run() {
	for {
		job, err := models.FirstJob(q.DB)
		if err != nil {
			log.Printf("[WORKER] Fetching job failed: %+v\n", err)
			continue
		}
		if job == nil {
			return
		}

		if err := q.HandleJob(job); err != nil {
			log.Printf("[WORKER] Handling job failed: %+v\n", err)
		}
		_ = job.Delete(q.DB)
	}
}

// HandleJob dispatches a job.
func (q *Queue) HandleJob(job *models.Job) error {
	// Start a job with a context and submission
	tx, err := q.DB.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer db.Rollback(tx)

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
		if _, err := worker.Compile(&worker.CompileContext{
			DB: tx, Sub: sub, Problem: problem, AllowLogs: q.Settings.LogCompile}); err != nil {
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
		if err := worker.Run(q.Sandbox, &worker.RunContext{
			DB: tx, Sub: sub, Problem: problem, TestGroup: tg, Test: test, AllowLogs: q.Settings.LogRun}); err != nil {
			return err
		}
	case models.JobTypeScore:
		contest, err := models.GetContest(tx, problem.ContestID)
		if err != nil {
			return err
		}
		if err := worker.Score(&worker.ScoreContext{
			DB: tx, Sub: sub, Problem: problem, Contest: contest, AllowLogs: q.Settings.LogScore}); err != nil {
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
	go func() {
		for range time.Tick(3 * time.Second) {
			toUpdate <- struct{}{}
		}
	}()
	return toUpdate
}
