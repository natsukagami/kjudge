// Package perf_test provides performance testing
package performance

import (
	"database/sql"
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/server/auth"
	"github.com/natsukagami/kjudge/worker"
	"github.com/pkg/errors"
)

// TODO: Output, Memory, Calculate, TLE

type PerfTestSet struct {
	Name          string
	Count  		  int                    
	CapTime       int                     // Time limit sent to sandbox
	Generator func(*rand.Rand) []byte // Returns input
	Solution      []byte                  // Solution to tested problem
	
}

// Generates problem and returns id
func (r *PerfTestSet) addProblem(db db.DBContext, seed int64, index int, contestID int) (int, error) {
	// Creates problem
	problem := &models.Problem{
		ContestID:                 contestID,
		DisplayName:               r.Name,
		ID:                        0,
		MaxSubmissionsCount:       0,
		MemoryLimit:               1 << 20, // 1GB
		Name:                      fmt.Sprintf("%v", index),
		PenaltyPolicy:             models.PenaltyPolicyNone,
		ScoringMode:               models.ScoringModeLast,
		SecondsBetweenSubmissions: 0,
		TimeLimit:                 r.CapTime,
	}
	if err := problem.Write(db); err != nil {
		return 0, errors.Wrapf(err, "problem %v", r.Name)
	}

	// Creates test group
	testGroup := &models.TestGroup{
		ID:          0,
		MemoryLimit: sql.NullInt64{Valid: false}, // nil
		Name:        "main",
		ProblemID:   problem.ID,
		Score:       100,
		ScoringMode: models.TestScoringModeSum,
		TimeLimit:   sql.NullInt64{Valid: false}, // nil
	}
	if err := testGroup.Write(db); err != nil {
		return 0, errors.Wrapf(err, "test group %v", testGroup.Name)
	}

	// Creates tests.
	rng := rand.New(rand.NewSource(seed))
	for i := 1; i < r.Count; i++ {
		test := &models.Test{
			ID:          0,
			Input:       r.Generator(rng),
			Name:        fmt.Sprintf("%v", i),
			Output:      []byte(""),
			TestGroupID: testGroup.ID,
		}
		if err := test.Write(db); err != nil {
			return 0, errors.Wrapf(err, "test %v", i)
		}
	}

	return problem.ID, nil
}

func (r *PerfTestSet) addSolution(db db.DBContext, problemID int, userID string) error {
	sub := models.Submission{
		ProblemID:   problemID,
		UserID:      userID,
		Source:      r.Solution,
		Language:    models.LanguageCpp,
		SubmittedAt: time.Now(),
		Verdict:     models.VerdictIsInQueue,
	}
	if err := sub.Write(db); err != nil {
		return err
	}

	job := models.NewJobScore(sub.ID)
	if err := job.Write(db); err != nil {
		return err
	}
	
	return nil
}

// Generates contest and returns contest ID
func createContest(db db.DBContext) (int, error) {
	contest := &models.Contest{
		ContestType:          "weighted",
		StartTime:            time.Now().AddDate(0, 0, -1),
		EndTime:              time.Now().AddDate(0, 0, 1),
		ID:                   0,
		Name:                 "Performance Testing",
		ScoreboardViewStatus: models.ScoreboardViewStatusPublic,
	}
	if err := contest.Write(db); err != nil {
		return 0, err
	}
	return contest.ID, nil
}

// Generates user and returns user ID
func createUser(db db.DBContext) (string, error) {
	password, err := auth.PasswordHash("bigquestions")
	if err != nil {
		return "", errors.Wrap(err, "while hashing password")
	}
	user := &models.User{
		ID: "Iroh",
		Password: string(password),
		DisplayName: "The Dragon of the West",
		Organization: "Order of the White Lotus",
	}
	if err := user.Write(db); err != nil {
		return "", errors.Wrap(err, "while creating user")
	}
	return user.ID, nil
}

func writeTestDB(benchDB db.DBContext, N int, testList ...*PerfTestSet) error {
	contestID, err := createContest(benchDB)
	if err != nil {
		return errors.Wrap(err, "creating contest")
	}

	userID, err := createUser(benchDB)
	if err != nil {
		return errors.Wrap(err, "creating user")
	}

	for idx, testset := range testList {
		problemID, err := testset.addProblem(benchDB, 2403, idx+1, contestID);
		if err != nil {
			return errors.Wrapf(err, "creating testset %v's problem", testset.Name)
		}
		for i := 0; i < N; i++ {
			testset.addSolution(benchDB, problemID, userID)
		}
	}
	return nil
}

// Copy a file by 4096 bytes chunk
func StreamCopy(src string, dst string) error {
	inf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer inf.Close()

	ouf, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer ouf.Close()
	
	buf := make([]byte, 4096)
	for {
		readLen, err := inf.Read(buf)
		lastIter := false
		if err == io.EOF {
			lastIter = true
		} else if err != nil {
			return err
		}
		
		if _, err := ouf.Write(buf[:readLen]); err != nil {
			return err
		}
		if lastIter {
			break
		}
	}
	return nil
}

func RunSingleTest(b *testing.B, tmpDir string, testset *PerfTestSet, sandboxName string) {
	dbFile := filepath.Join(tmpDir, fmt.Sprintf("%v_%v_%v.db", testset.Name, sandboxName, b.N))
	
	benchDB, err := db.New(dbFile)
	if err != nil {
		b.Error(err)
		b.FailNow()
	}
	defer benchDB.Close()
	
	b.Logf("Generating %v test suite", testset.Name)
	if err := writeTestDB(benchDB, b.N, testset); err != nil {
		b.Error(err)
		b.FailNow()
	}
	defer benchDB.Close()

	sandbox, err := worker.NewSandbox(sandboxName)
	if err != nil {
		b.Error(err)
		b.FailNow()
	}

	b.ResetTimer()
	queue := &worker.Queue{Sandbox: sandbox, DB: benchDB}
	queue.Run()
}
