package performance

import (
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/server/auth"
	"github.com/natsukagami/kjudge/worker"
	"github.com/natsukagami/kjudge/worker/queue"
	"github.com/natsukagami/kjudge/worker/sandbox"
	"github.com/pkg/errors"
)

type BenchmarkContext struct {
	tdir     string
	db       *db.DB
	user     *models.User
	contest  *models.Contest
	problems map[string]*models.Problem
}

func NewBenchmarkContext(tmpDir string) (*BenchmarkContext, error) {
	benchDB, err := db.New(filepath.Join(tmpDir, "bench.db"))
	if err != nil {
		err2 := os.RemoveAll(tmpDir)
		if err2 != nil {
			return nil, errors.Wrapf(err2, "while handling %v", err)
		}
		return nil, err
	}

	ctx := &BenchmarkContext{tdir: tmpDir, db: benchDB, problems: make(map[string]*models.Problem)}

	if err := ctx.writeContest(); err != nil {
		return nil, err
	}

	if err := ctx.writeUser(); err != nil {
		return nil, err
	}

	return ctx, nil
}

func (ctx *BenchmarkContext) Close() error {
	if err := os.RemoveAll(ctx.tdir); err != nil {
		return err
	}
	if err := ctx.db.Close(); err != nil {
		return err
	}
	return nil
}

func (ctx *BenchmarkContext) writeContest() error {
	ctx.contest = &models.Contest{
		ContestType:          "weighted",
		StartTime:            time.Now().AddDate(0, 0, -1),
		EndTime:              time.Now().AddDate(0, 0, 1),
		ID:                   0,
		Name:                 "Performance Testing",
		ScoreboardViewStatus: models.ScoreboardViewStatusPublic,
	}
	return errors.Wrapf(ctx.contest.Write(ctx.db), "creating contest")
}

func (ctx *BenchmarkContext) writeUser() error {
	password, err := auth.PasswordHash("bigquestions")
	if err != nil {
		return errors.Wrap(err, "hashing password")
	}
	ctx.user = &models.User{
		ID:           "Iroh",
		Password:     string(password),
		DisplayName:  "The Dragon of the West",
		Organization: "Order of the White Lotus",
	}
	return errors.Wrap(ctx.user.Write(ctx.db), "creating user")
}

func (ctx *BenchmarkContext) writeProblem(testset *PerfTestSet) error {
	problem, err := testset.AddToDB(ctx.db, 2403, len(ctx.problems)+1, ctx.contest.ID)
	if err != nil {
		return errors.Wrapf(err, "creating testset %v's problem", testset.Name)
	}

	ctx.problems[testset.Name] = problem
	return nil
}

const testSolution = `#include "solution.hpp"
`

func (ctx *BenchmarkContext) writeSolutions(N int, problemName string) error {
	problem := ctx.problems[problemName]
	for i := 0; i < N; i++ {
		sub := models.Submission{
			ProblemID:   problem.ID,
			UserID:      ctx.user.ID,
			Source:      []byte(testSolution),
			Language:    models.LanguageCpp,
			SubmittedAt: time.Now(),
			Verdict:     models.VerdictIsInQueue,
		}
		if err := sub.Write(ctx.db); err != nil {
			return err
		}

		job := models.NewJobScore(sub.ID)
		if err := job.Write(ctx.db); err != nil {
			return err
		}
	}
	return nil
}

func RunSingleTest(b *testing.B, ctx *BenchmarkContext, testset *PerfTestSet, sandboxName string) {
	log.Printf("running %v %v %v times", testset.Name, sandboxName, b.N)
	sandbox, err := worker.NewSandbox(
		sandboxName,
		sandbox.IgnoreWarnings(true),
		sandbox.EnableSandboxLogs(false))

	if err != nil {
		b.Fatal(err)
	}

	log.Printf("Generating %v solutions", b.N)
	for i := 0; i < b.N; i++ {
		ctx.writeSolutions(b.N, testset.Name)
	}
	log.Printf("Generated solutions")

	queue := queue.NewQueue(ctx.db, sandbox, queue.CompileLogs(false), queue.RunLogs(false), queue.ScoreLogs(false))

	b.ResetTimer()
	queue.Run()
	b.StopTimer()

	if err := ctx.assertRunComplete(testset); err != nil {
		b.Fatal(err)
	}
}

// TODO
func (ctx *BenchmarkContext) assertRunComplete(testset *PerfTestSet) error {
	return nil
}
