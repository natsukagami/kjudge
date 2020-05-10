package models

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models/verify"
	"github.com/pkg/errors"
)

// ScoringMode dictates how the best submission is chosen.
// There are:
// - Min: The submission with the lowest score is chosen. If on a tie, choose the one with the highest penalty
// - Best: The submission with the highest score is chosen. If on a tie, choose the one with lowest penalty.
// - Once: The first (successfully compiled) submission is the best one.
// - Last: The last submission is the best one.
// - Decay: The last submission is the best one. The score is modified by the number of submissions before it (0.1 * count), and the
// time passed (0.7 * time passed in %), to a minimum of 0.3 times the original.
type ScoringMode string

// Defined values for ScoringMode.
const (
	ScoringModeMin   ScoringMode = "min"
	ScoringModeBest  ScoringMode = "best"
	ScoringModeOnce  ScoringMode = "once"
	ScoringModeLast  ScoringMode = "last"
	ScoringModeDecay ScoringMode = "decay"
)

func (s ScoringMode) verify() error {
	return verify.String(string(s), verify.Enum(string(ScoringModeMin), string(ScoringModeBest), string(ScoringModeOnce), string(ScoringModeLast), string(ScoringModeDecay)))
}

// PenaltyPolicy dictates how the penalty is calculated.
// There are:
// - None: no penalty at all (IOI - like).
// - SubmitTime: minutes passed from the start of the contest, of the submission time.
// - ICPC: SubmitTime + 20 * (# of past submissions)
type PenaltyPolicy string

// Defined values for PenaltyPolicy.
const (
	PenaltyPolicyNone       PenaltyPolicy = "none"
	PenaltyPolicySubmitTime PenaltyPolicy = "submit_time"
	PenaltyPolicyICPC       PenaltyPolicy = "icpc"
)

func (p PenaltyPolicy) verify() error {
	return verify.String(string(p), verify.Enum(string(PenaltyPolicyNone), string(PenaltyPolicySubmitTime), string(PenaltyPolicyICPC)))
}

// Verify verifies a Problem's content.
func (r *Problem) Verify() error {
	return verify.All(map[string]error{
		"DisplayName":               verify.Names(r.DisplayName),
		"Name":                      verify.Names(r.Name),
		"ScoringMode":               r.ScoringMode.verify(),
		"PenaltyPolicy":             r.PenaltyPolicy.verify(),
		"TimeLimit":                 verify.IntPositive(r.TimeLimit),
		"MemoryLimit":               verify.IntPositive(r.MemoryLimit),
		"MaxSubmissionsCount":       verify.IntMin(0)(r.MaxSubmissionsCount),
		"SecondsBetweenSubmissions": verify.IntMin(0)(r.SecondsBetweenSubmissions),
	})
}

// AdminLink is the link to the problem in the admin panel.
func (r *Problem) AdminLink() string {
	return fmt.Sprintf("/admin/problems/%d", r.ID)
}

// Link is the link to the problem.
func (r *Problem) Link() string {
	return fmt.Sprintf("/contests/%d/problems/%s", r.ContestID, r.Name)
}

// ProblemWithTestGroups is a problem with attached test groups,
// that will provide score-related statistics.
type ProblemWithTestGroups struct {
	*Problem
	TestGroups []*TestGroup
}

// TotalScore returns the problem's maximal total score.
func (p *ProblemWithTestGroups) TotalScore() float64 {
	total := 0.0
	for _, tg := range p.TestGroups {
		if tg.Score >= 0 {
			total += tg.Score
		}
	}
	return total
}

// SubtaskScores returns the problem's test group scores as a list seperated by forward slash.
func (p *ProblemWithTestGroups) SubtaskScores() string {
	first := true
	var res strings.Builder
	for _, tg := range p.TestGroups {
		if tg.Score < 0 {
			continue
		}
		if !first {
			res.WriteString("/")
		}
		first = false
		res.WriteString(fmt.Sprintf("%.2f", tg.Score))
	}
	return res.String()
}

// CollectTestGroups collects the test groups for a list of problems.
func CollectTestGroups(db db.DBContext, problems []*Problem, private bool) ([]*ProblemWithTestGroups, error) {
	if len(problems) == 0 {
		return nil, nil
	}
	pMap := make(map[int]*ProblemWithTestGroups)
	var IDs []int
	for _, p := range problems {
		pMap[p.ID] = &ProblemWithTestGroups{Problem: p}
		IDs = append(IDs, p.ID)
	}
	privateQuery := ""
	if !private {
		privateQuery = " AND score >= 0"
	}
	var tgs []*TestGroup
	query, args, err := sqlx.In("SELECT * FROM test_groups WHERE problem_id IN (?)"+privateQuery+queryTestGroupOrderBy, IDs)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if err := db.Select(&tgs, query, args...); err != nil {
		return nil, errors.WithStack(err)
	}
	for _, tg := range tgs {
		p := pMap[tg.ProblemID]
		p.TestGroups = append(p.TestGroups, tg)
	}
	var ps []*ProblemWithTestGroups
	for _, p := range problems {
		ps = append(ps, pMap[p.ID])
	}
	return ps, nil
}

// GetProblemByName gets a Problem from the Database by its name and contest.
func GetProblemByName(db db.DBContext, contestID int, name string) (*Problem, error) {
	var result Problem
	if err := db.Get(&result, "SELECT * FROM problems WHERE contest_id = ? AND name = ?", contestID, name); err != nil {
		return nil, errors.WithStack(err)
	}
	return &result, nil
}
