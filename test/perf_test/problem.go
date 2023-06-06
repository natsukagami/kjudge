// Package perf_test provides performance testing
package perf_test

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
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

func (r *PerfTestSet) addSolution(db db.DBContext, problemID int, N int) error {

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
func createUser(db db.DBContext) (int, error) {
	
}

func generateDB(dbFile string, N int, testList ...*PerfTestSet) error {
	benchDB, err := db.New(dbFile)
	if err != nil {
		return errors.Wrap(err, "creating DB")
	}

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
		testset.addSolution(benchDB, problemID, N)
	}
	return nil
}

func generateTestSuite(b *testing.B){
	tmpDir, err := os.MkdirTemp(os.TempDir(), "kjudge_bench")
	if err != nil {
		log.Panic("cannot create temp dir:", err)
	}
	defer os.RemoveAll(tmpDir)
	dbFile := filepath.Join(tmpDir, "kjudge.db")
	generateDB(dbFile, b.N, BigInputProblem(), SpawnTimeProblem())
}
