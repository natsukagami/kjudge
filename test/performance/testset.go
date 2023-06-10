// Package perf_test provides performance testing
package performance

import (
	"database/sql"
	"fmt"
	"math/rand"

	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/pkg/errors"
)

// TODO: Output, Memory, Calculate, TLE

type PerfTestSet struct {
	Name      string
	Count     int
	CapTime   int                     // Time limit sent to sandbox
	Generator func(*rand.Rand) []byte // Returns input
	Solution  []byte                  // Solution to tested problem

}

// Generates problem and returns id
func (r *PerfTestSet) AddToDB(db db.DBContext, seed int64, index int, contestID int) (*models.Problem, error) {
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
		return nil, errors.Wrapf(err, "problem %v", r.Name)
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
		return nil, errors.Wrapf(err, "test group %v", testGroup.Name)
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
			return nil, errors.Wrapf(err, "test %v", i)
		}
	}

	// Uploads solution
	solutionFile := &models.File{
		Content:   r.Solution,
		Filename:  "solution.hpp",
		ID:        0,
		ProblemID: problem.ID,
		Public:    true,
	}
	if err := problem.WriteFiles(db, []*models.File{solutionFile}); err != nil {
		return nil, errors.Wrap(err, "solution file")
	}

	return problem, nil
}
