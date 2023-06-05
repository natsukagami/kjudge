// Package perf_test provides performance testing
package perf_test

import (
	"math/rand"

	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
)

func generateProblem(db db.DBContext, contestID int, displayName string, name string) *models.Problem {
	return &models.Problem{
		ContestID:                 contestID,
		DisplayName:               displayName,
		ID:                        0,       // Auto generate ID
		MaxSubmissionsCount:       0,       // No limit
		MemoryLimit:               1 << 20, // 1GB
		Name:                      name,
		PenaltyPolicy:             "none",
		ScoringMode:               "best",
		SecondsBetweenSubmissions: 0,
		TimeLimit:                 5,
	}
}

type PerfTest struct {
	Input  []byte
	Output []byte
}

type PerfTestSet struct {
	Name          string
	ExpectedTime  int // Expected running time of each test in ms
	TestGenerator func(*rand.Rand) *PerfTest
	TestCode      []byte // Solution to tested problem
	ExpectedAC    bool
}

func GenerateContest() {

}
