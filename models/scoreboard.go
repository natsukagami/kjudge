package models

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"sort"
	"time"

	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/server/httperr"
	"github.com/pkg/errors"
)

// UserResult stores information about user's preformance in the contest
type UserResult struct {
	User *User

	Rank           int
	TotalPenalty   int
	SolvedProblems int
	TotalScore     float64

	ProblemResults map[int]*ProblemResult
}

// JSONScoreboard represents a JSON encoded scoreboard.
type JSONScoreboard struct {
	ContestID           int              `json:"contest_id"`
	ContestType         ContestType      `json:"contest_type"`
	Problems            []JSONProblem    `json:"problems"`
	Users               []JSONUserResult `json:"users"`
	ProblemFirstSolvers map[int]int64    `json:"problem_first_solvers"`
}

// JSONUserResult represents a JSON encoded user in the scoreboard.
type JSONUserResult struct {
	ID             string                    `json:"id"`
	DisplayName    string                    `json:"display_name"`
	Organization   string                    `json:"organization,omitempty"`
	Rank           int                       `json:"rank"`
	TotalPenalty   int                       `json:"total_penalty"`
	SolvedProblems int                       `json:"solved_problems"`
	TotalScore     float64                   `json:"total_score"`
	ProblemResults map[int]JSONProblemResult `json:"problem_results"`
}

func jsonUserResult(u *UserResult, ps []JSONProblem) JSONUserResult {
	problems := make(map[int]JSONProblemResult)
	for _, p := range ps {
		problems[p.ID] = jsonProblemResult(u.ProblemResults[p.ID])
	}
	return JSONUserResult{
		ID:             u.User.ID,
		DisplayName:    u.User.DisplayName,
		Organization:   u.User.Organization,
		Rank:           u.Rank,
		TotalPenalty:   u.TotalPenalty,
		SolvedProblems: u.SolvedProblems,
		TotalScore:     u.TotalScore,
		ProblemResults: problems,
	}
}

// JSONProblemResult represents a JSON encoded user's result of a problem in the scoreboard.
type JSONProblemResult struct {
	Score          float64 `json:"score"`
	Solved         bool    `json:"solved"`
	Penalty        int     `json:"penalty"`
	FailedAttempts int     `json:"failed_attempts"`
	BestSubmission int64   `json:"best_submission"`
}

func jsonProblemResult(p *ProblemResult) JSONProblemResult {
	if p == nil {
		return JSONProblemResult{}
	}

	var bestSubmission int64
	if p.BestSubmissionID.Valid {
		bestSubmission = p.BestSubmissionID.Int64
	} else {
		bestSubmission = -1
	}
	return JSONProblemResult{
		Score:          p.Score,
		Solved:         p.Solved,
		Penalty:        p.Penalty,
		FailedAttempts: p.FailedAttempts,
		BestSubmission: bestSubmission,
	}
}

// JSONProblem represents a JSON encoded problem metadata.
type JSONProblem struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

func jsonProblem(p *Problem) JSONProblem {
	return JSONProblem{
		ID:          p.ID,
		Name:        p.Name,
		DisplayName: p.DisplayName,
	}
}

// Scoreboard is the struct used to render scoreboard
type Scoreboard struct {
	Contest             *Contest
	Problems            []*Problem
	UserResults         []*UserResult
	ProblemFirstSolvers map[int]int64
}

// JSON returns the JSON representation of the scoreboard.
func (s *Scoreboard) JSON() JSONScoreboard {
	sb := JSONScoreboard{
		ContestID:           s.Contest.ID,
		ContestType:         s.Contest.ContestType,
		ProblemFirstSolvers: s.ProblemFirstSolvers,
	}
	for _, p := range s.Problems {
		sb.Problems = append(sb.Problems, jsonProblem(p))
	}
	for _, u := range s.UserResults {
		sb.Users = append(sb.Users, jsonUserResult(u, sb.Problems))
	}
	return sb
}

// compareUserRanking checks if ranking of user[i] is strictly less than the ranking of user[j]
// Returns (comparison, is it just tie-breaking)
func compareUserRanking(userResult []*UserResult, contestType ContestType, i, j int) (bool, bool) {
	a, b := userResult[i], userResult[j]
	switch contestType {
	case ContestTypeWeighted:
		// sort based on totalScore if two users have same totalScore sort based on totalPenalty in an ascending order
		if a.TotalScore != b.TotalScore {
			return a.TotalScore > b.TotalScore, false
		}
		if a.TotalPenalty != b.TotalPenalty {
			return a.TotalPenalty < b.TotalPenalty, false
		}
		return a.User.ID < b.User.ID, true
	case ContestTypeUnweighted:
		// sort based on solvedProblems if two users have same solvedProblems sort based on totalPenalty in an ascending order
		if a.SolvedProblems != b.SolvedProblems {
			return a.SolvedProblems > b.SolvedProblems, false
		}
		if a.TotalPenalty != b.TotalPenalty {
			return a.TotalPenalty < b.TotalPenalty, false
		}
		return a.User.ID < b.User.ID, true
	}
	log.Panicf("unexpected contest type %s", contestType)
	return true, true
}

// Get scoreboard given problems and contest
func GetScoreboard(db db.DBContext, contest *Contest, problems []*Problem) (*Scoreboard, error) {
	// If the contest has not started, throw
	if contest.StartTime.After(time.Now()) {
		return nil, httperr.BadRequestf("Contest has not started")
	}

	// get contestType (weighted and unweighted)
	contestType := contest.ContestType

	users, err := GetAllUsers(db)
	if err != nil {
		return nil, err
	}

	contestProblemResults, err := CollectContestProblemResults(db, problems)
	if err != nil {
		return nil, err
	}

	userResults := []*UserResult{}
	userProblemResults := make(map[string]*UserResult)
	for _, user := range users {
		userProblemResults[user.ID] = &UserResult{
			User:           user,
			ProblemResults: make(map[int]*ProblemResult),
		}
	}
	for _, problemResult := range contestProblemResults {
		userID := problemResult.UserID
		problemID := problemResult.ProblemID

		userProblemResults[userID].TotalScore += problemResult.Score
		userProblemResults[userID].TotalPenalty += problemResult.Penalty
		if problemResult.Solved {
			userProblemResults[userID].SolvedProblems++
		}

		userProblemResults[userID].ProblemResults[problemID] = problemResult
	}

	for _, userProblemResult := range userProblemResults {
		// not display users with no submissions and hidden users
		if len(userProblemResult.ProblemResults) > 0 && !userProblemResult.User.Hidden {
			userResults = append(userResults, userProblemResult)
		}
	}

	// get bestSubmission ID for each problem
	problemFirstSolvers := make(map[int]int64)

	for _, userProblemResult := range userProblemResults {
		// not consider users with no submissions and hidden users
		if len(userProblemResult.ProblemResults) == 0 || userProblemResult.User.Hidden {
			continue
		}

		problemResults := userProblemResult.ProblemResults
		for _, problemResult := range problemResults {
			problemID := problemResult.ProblemID
			// skip the problemResult with verdict != Solved
			if !problemResult.Solved {
				continue
			}
			// skip if there is no submission
			if !problemResult.BestSubmissionID.Valid {
				continue
			}
			submissionID, ok := problemFirstSolvers[problemID]

			if !ok || submissionID > problemResult.BestSubmissionID.Int64 {
				problemFirstSolvers[problemID] = problemResult.BestSubmissionID.Int64
			}
		}
	}

	sort.Slice(userResults, func(i, j int) bool {
		r, _ := compareUserRanking(userResults, contestType, i, j)
		return r
	})

	// after sorting users, we need to calculate users' ranking
	rank := 0
	for i, userCtx := range userResults {
		if i == 0 {
			rank = i + 1
		} else if r, tie := compareUserRanking(userResults, contestType, i-1, i); r && !tie {
			rank = i + 1
		}
		userCtx.Rank = rank
	}

	return &Scoreboard{
		Contest:             contest,
		Problems:            problems,
		UserResults:         userResults,
		ProblemFirstSolvers: problemFirstSolvers,
	}, nil
}

// CSVScoresOnly returns the CSV version of the scoreboard, with only scores.
func (s *Scoreboard) CSVScoresOnly(w io.Writer) error {
	writer := csv.NewWriter(w)
	// First row: Headers
	headers := []string{"Username", "Name", "Organization", "Total Score"}
	for _, p := range s.Problems {
		headers = append(headers, p.Name)
	}
	if err := writer.Write(headers); err != nil {
		return errors.WithStack(err)
	}
	// One for each contestants
	for _, u := range s.UserResults {
		row := []string{u.User.ID, u.User.DisplayName, u.User.Organization, fmt.Sprintf("%.2f", u.TotalScore)}
		for _, p := range s.Problems {
			if score, ok := u.ProblemResults[p.ID]; ok {
				row = append(row, fmt.Sprintf("%.2f", score.Score))
			} else {
				row = append(row, "-")
			}
		}
		if err := writer.Write(row); err != nil {
			return errors.Wrapf(err, "row %s", u.User.ID)
		}
	}
	writer.Flush()
	return errors.WithStack(writer.Error())
}

// CSV returns the CSV version of the scoreboard, with scores and penalties.
func (s *Scoreboard) CSV(w io.Writer) error {
	writer := csv.NewWriter(w)
	// First row: Headers
	headers := []string{"Username", "Name", "Organization", "Total Score", "Total Penalty"}
	for _, p := range s.Problems {
		headers = append(headers, p.Name, p.Name+" (Penalty)")
	}
	if err := writer.Write(headers); err != nil {
		return errors.WithStack(err)
	}
	// One for each contestants
	for _, u := range s.UserResults {
		row := []string{u.User.ID, u.User.DisplayName, u.User.Organization, fmt.Sprintf("%.2f", u.TotalScore), fmt.Sprint(u.TotalPenalty)}
		for _, p := range s.Problems {
			if score, ok := u.ProblemResults[p.ID]; ok {
				row = append(row, fmt.Sprintf("%.2f", score.Score), fmt.Sprint(score.Penalty))
			} else {
				row = append(row, "-", "-")
			}
		}
		if err := writer.Write(row); err != nil {
			return errors.Wrapf(err, "row %s", u.User.ID)
		}
	}
	writer.Flush()
	return errors.WithStack(writer.Error())
}
