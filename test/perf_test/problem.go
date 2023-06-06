// Package perf_test provides performance testing
package perf_test

import (
	"database/sql"
	"fmt"
	"math/rand"
	"time"

	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/pkg/errors"
)

// TODO: Output, Memory, Calculate, TLE

type PerfTestSet struct {
	Name          string
	ExpectedTime  int                     // Expected running time of each test in ms
	CapTime       int                     // Time limit sent to sandbox
	TestGenerator func(*rand.Rand) []byte // Returns input
	TestCode      []byte                  // Solution to tested problem
}

// Generates problem and returns problem ID
func (r *PerfTestSet) AddToDB(db db.DBContext, seed int64, index int, contestID int, expectedTime int) (int, error) {
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
	for i := 1; i*r.ExpectedTime < expectedTime; i++ {
		test := &models.Test{
			ID:          0,
			Input:       r.TestGenerator(rng),
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

// Generates contest and returns contest ID
func GenerateContest(db db.DBContext) (int, error) {
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
