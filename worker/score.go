package worker

import (
	"database/sql"
	"log"
	"math"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/natsukagami/kjudge/models"
)

// ScoreContext is a context for calculating a submission's score
// and update the user's problem scores.
type ScoreContext struct {
	DB        *sqlx.Tx
	Sub       *models.Submission
	Problem   *models.Problem
	Contest   *models.Contest
	AllowLogs bool
}

func (s *ScoreContext) Log(format string, v ...interface{}) {
	if !s.AllowLogs {
		return
	}
	log.Printf(format, v...)
}

// Score does scoring on a submission and updates the user's ProblemResult.
func Score(s *ScoreContext) error {
	// Check if there's any test results missing.
	testResults, err := s.TestResults()
	if err != nil {
		return err
	}
	tests, err := models.GetProblemTestsMeta(s.DB, s.Problem.ID)
	if err != nil {
		return err
	}
	if compiled, source := s.CompiledSource(); !compiled {
		// Add a compilation job and re-add ourselves.
		s.Log("[WORKER] Submission %v not compiled, creating Compile job.\n", s.Sub.ID)
		return models.BatchInsertJobs(s.DB, models.NewJobCompile(s.Sub.ID), models.NewJobScore(s.Sub.ID))
	} else if source == nil {
		s.Log("[WORKER] Not running a submission that failed to compile.\n")
		s.Sub.Verdict = models.VerdictCompileError
		if err := s.Sub.Write(s.DB); err != nil {
			return err
		}
		// Update the ProblemResult
		subs, err := models.GetUserProblemSubmissions(s.DB, s.Sub.UserID, s.Problem.ID)
		if err != nil {
			return err
		}
		pr := s.CompareScores(subs)
		s.Log("[WORKER] Problem results updated for user %s, problem %d (score = %.1f, penalty = %d)\n", s.Sub.UserID, s.Problem.ID, pr.Score, pr.Penalty)

		return pr.Write(s.DB)
	}
	if missing := MissingTests(tests, testResults); len(missing) > 0 {
		s.Log("[WORKER] Submission %v needs to run %d tests before being scored.\n", s.Sub.ID, len(missing))
		var jobs []*models.Job
		for _, m := range missing {
			jobs = append(jobs, models.NewJobRun(s.Sub.ID, m.ID))
		}
		jobs = append(jobs, models.NewJobScore(s.Sub.ID))
		return models.BatchInsertJobs(s.DB, jobs...)
	}

	s.Log("[WORKER] Scoring submission %d\n", s.Sub.ID)
	// Calculate the score by summing scores on each test group.
	s.Sub.Score = sql.NullFloat64{Float64: 0.0, Valid: true}
	for _, tg := range tests {
		if !tg.Hidden() {
			s.Sub.Score.Float64 += tg.ComputeScore(testResults)
		}
	}
	// Calculate penalty too
	if err := s.ComputePenalties(s.Sub); err != nil {
		return err
	}
	// Verdict
	UpdateVerdict(tests, s.Sub)
	// Write the submission's score
	if err := s.Sub.Write(s.DB); err != nil {
		return err
	}
	s.Log("[WORKER] Submission %d scored (verdict = %s, score = %.1f). Updating problem results\n", s.Sub.ID, s.Sub.Verdict, s.Sub.Score.Float64)

	// Update the ProblemResult
	subs, err := models.GetUserProblemSubmissions(s.DB, s.Sub.UserID, s.Problem.ID)
	if err != nil {
		return err
	}
	pr := s.CompareScores(subs)
	s.Log("[WORKER] Problem results updated for user %s, problem %d (score = %.1f, penalty = %d)\n", s.Sub.UserID, s.Problem.ID, pr.Score, pr.Penalty)

	return pr.Write(s.DB)
}

// Update the submission's verdict.
func UpdateVerdict(tests []*models.TestGroupWithTests, sub *models.Submission) {
	score, _, counts := scoreOf(sub)
	if !counts {
		sub.Verdict = models.VerdictCompileError
		return
	}

	maxPossibleScore := 0.0
	for _, tg := range tests {
		if tg.Score > 0 {
			maxPossibleScore += tg.Score
		}
	}

	if score == maxPossibleScore {
		sub.Verdict = models.VerdictAccepted
	} else {
		sub.Verdict = models.VerdictScored
	}
}

// TestResults returns the submission's test results, mapped by the test's ID.
func (s *ScoreContext) TestResults() (map[int]*models.TestResult, error) {
	trs, err := models.GetSubmissionTestResults(s.DB, s.Sub.ID)
	if err != nil {
		return nil, err
	}
	res := make(map[int]*models.TestResult)
	for _, tr := range trs {
		res[tr.TestID] = tr
	}
	return res, nil
}

// ComputePenalties compute penalty values for each submission, based on the PenaltyPolicy.
func (s *ScoreContext) ComputePenalties(sub *models.Submission) error {
	value := 0
	switch s.Problem.PenaltyPolicy {
	case models.PenaltyPolicyNone:
	case models.PenaltyPolicyICPC:
		subs, err := models.GetUserProblemSubmissions(s.DB, s.Sub.UserID, s.Problem.ID)
		if err != nil {
			return err
		}
		for id, s := range subs {
			if sub.ID == s.ID {
				value = 20 * id
				break
			}
		}
		fallthrough // We also need the submit time
	case models.PenaltyPolicySubmitTime:
		// Sometimes the penalty can be messed up
		submitTimePenalty := int((sub.SubmittedAt.Sub(s.Contest.StartTime) + time.Minute - 1) / time.Minute)
		if submitTimePenalty >= 0 {
			value += submitTimePenalty
		}
	default:
		panic(s)
	}
	sub.Penalty = sql.NullInt64{Int64: int64(value), Valid: true}
	return nil
}

// Returns (score, penalty, should_count).
func scoreOf(sub *models.Submission) (float64, int, bool) {
	if sub == nil || sub.CompiledSource == nil || !sub.Score.Valid || !sub.Penalty.Valid {
		// Looks like a pending submission
		return 0, 0, false
	}
	return sub.Score.Float64, int(sub.Penalty.Int64), true
}

// CompareScores compare the submission results and return the best one.
// If nil is returned, then the problem result should just be removed.
// The submissions list passed in must be sorted in the OrderBy order.
func (s *ScoreContext) CompareScores(subs []*models.Submission) *models.ProblemResult {
	maxScore := 0.0
	var which *models.Submission
	contestTime := float64(s.Contest.EndTime.Sub(s.Contest.StartTime))
	counted := 0
	failedAttempts := 0

	// Since the submissions' order are by submit time desc, we need to reverse the list.
	for i, j := 0, len(subs)-1; i < j; i, j = i+1, j-1 {
		subs[i], subs[j] = subs[j], subs[i]
	}

getScoredSub:
	for _, sub := range subs {
		score, _, counts := scoreOf(sub)
		if !counts {
			continue
		}
		counted++
		switch s.Problem.ScoringMode {
		case models.ScoringModeOnce:
			which = sub
			maxScore = score
			break getScoredSub
		case models.ScoringModeLast:
			which = sub
			maxScore = score
		case models.ScoringModeDecay:
			score = score * math.Max(0.3,
				(1.0-0.7*float64(sub.SubmittedAt.Sub(s.Contest.StartTime))/contestTime)*
					(1.0-0.1*float64(counted)))
			fallthrough
		case models.ScoringModeBest:
			if which == nil || score > which.Score.Float64 {
				which = sub
				maxScore = score
			}
		case models.ScoringModeMin:
			if which == nil || score <= which.Score.Float64 {
				which = sub
				maxScore = score // this is literally min score
			}
		default:
			panic(s)
		}
	}

	for _, sub := range subs {
		_, _, counts := scoreOf(sub)
		if !counts {
			continue
		}
		if s.Problem.ScoringMode == models.ScoringModeMin {
			if sub == which {
				if sub.Verdict != models.VerdictAccepted {
					failedAttempts++
				}
				break
			}
		} else if sub.Verdict == models.VerdictAccepted {
			break
		}
		failedAttempts++
	}

	_, penalty, counts := scoreOf(which)
	if !counts {
		return &models.ProblemResult{
			BestSubmissionID: sql.NullInt64{},
			FailedAttempts:   failedAttempts,
			Penalty:          0,
			Score:            0.0,
			Solved:           false,
			ProblemID:        s.Problem.ID,
			UserID:           s.Sub.UserID,
		}
	}

	// Don't consider penalty in certain scenarios...
	contestType := s.Contest.ContestType
	if contestType == models.ContestTypeWeighted && maxScore == 0.0 {
		penalty = 0
	} else if contestType == models.ContestTypeUnweighted && which.Verdict != models.VerdictAccepted {
		penalty = 0
	}
	return &models.ProblemResult{
		BestSubmissionID: sql.NullInt64{Int64: int64(which.ID), Valid: true},
		FailedAttempts:   failedAttempts,
		Penalty:          penalty,
		Score:            maxScore,
		Solved:           which.Verdict == models.VerdictAccepted,
		ProblemID:        s.Problem.ID,
		UserID:           s.Sub.UserID,
	}
}

// MissingTests finds all the tests that are missing a TestResult.
func MissingTests(tests []*models.TestGroupWithTests, results map[int]*models.TestResult) []*models.Test {
	var res []*models.Test
	for _, tg := range tests {
		for _, test := range tg.Tests {
			if _, ok := results[test.ID]; !ok {
				res = append(res, test)
			}
		}
	}
	return res
}

// CompiledSource returns the CompiledSource. Returns false when the submission hasn't been compiled.
// Returns nil if the submission failed to compile.
func (s *ScoreContext) CompiledSource() (bool, []byte) {
	if s.Sub.CompilerOutput == nil {
		return false, nil
	}
	return true, s.Sub.CompiledSource
}
