package contests

import (
	"net/http"
	"sort"
	"time"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models"
	"git.nkagami.me/natsukagami/kjudge/server/httperr"
	"github.com/labstack/echo/v4"
)

// UserResult stores information about user's preformance in the contest
type UserResult struct {
	User *models.User

	Rank           int
	TotalPenalty   int
	SolvedProblems int
	TotalScore     float64

	ProblemResults map[int]*models.ProblemResult
}

// JSONScoreboard represents a JSON encoded scoreboard.
type JSONScoreboard struct {
	ContestID              int                `json:"contest_id"`
	ContestType            models.ContestType `json:"contest_type"`
	Problems               []JSONProblem      `json:"problems"`
	Users                  []JSONUserResult   `json:"users"`
	ProblemBestSubmissions map[int]int64      `json:"problem_best_submissions"`
}

// JSONUserResult represents a JSON encoded user in the scoreboard.
type JSONUserResult struct {
	ID             string                    `json:"id"`
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

func jsonProblemResult(p *models.ProblemResult) JSONProblemResult {
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

func jsonProblem(p *models.Problem) JSONProblem {
	return JSONProblem{
		ID:          p.ID,
		Name:        p.Name,
		DisplayName: p.DisplayName,
	}
}

// ScoreboardCtx is the context required to display the scoreboard page
type ScoreboardCtx struct {
	*ContestCtx

	UserResults            []*UserResult
	ProblemBestSubmissions map[int]int64
}

// Render renders the scoreboard context
func (s *ScoreboardCtx) Render(c echo.Context) error {
	return c.Render(http.StatusOK, "contests/scoreboard", s)
}

// JSON returns the JSON representation of the scoreboard.
func (s *ScoreboardCtx) JSON() JSONScoreboard {
	sb := JSONScoreboard{
		ContestID:              s.Contest.ID,
		ContestType:            s.Contest.ContestType,
		ProblemBestSubmissions: s.ProblemBestSubmissions,
	}
	for _, p := range s.Problems {
		sb.Problems = append(sb.Problems, jsonProblem(p))
	}
	for _, u := range s.UserResults {
		sb.Users = append(sb.Users, jsonUserResult(u, sb.Problems))
	}
	return sb
}

// RenderJSON renders a scoreboard in JSON.
func (s *ScoreboardCtx) RenderJSON(c echo.Context) error {
	return c.JSON(http.StatusOK, s.JSON())
}

// compareUserRanking checks if ranking of user[i] is strictly less than the ranking of user[j]
// Returns (comparison, is it just tie-breaking)
func compareUserRanking(userResult []*UserResult, contestType models.ContestType, i, j int) (bool, bool) {
	a, b := userResult[i], userResult[j]
	if contestType == models.ContestTypeWeighted {
		// sort based on totalScore if two users have same totalScore sort based on totalPenalty in an ascending order
		if a.TotalScore != b.TotalScore {
			return a.TotalScore > b.TotalScore, false
		}
		if a.TotalPenalty != b.TotalPenalty {
			return a.TotalPenalty < b.TotalPenalty, false
		}
		return a.User.ID < b.User.ID, true
	} else {
		// sort based on solvedProblems if two users have same solvedProblems sort based on totalPenalty in an ascending order
		if a.SolvedProblems != b.SolvedProblems {
			return a.SolvedProblems > b.SolvedProblems, false
		}
		if a.TotalPenalty != b.TotalPenalty {
			return a.TotalPenalty < b.TotalPenalty, false
		}
		return a.User.ID < b.User.ID, true
	}
}

// compare two users performance in a problem
func compareProblemResult(r1, r2 *models.ProblemResult) bool {
	if r1.Score != r2.Score {
		return r1.Score > r2.Score
	} else if r1.Penalty != r2.Penalty {
		return r1.Penalty < r2.Penalty
	} else {
		return r1.UserID < r2.UserID
	}
}

// Collect a ScoreboardCtx
func getScoreboardCtx(db db.DBContext, c echo.Context) (*ScoreboardCtx, error) {
	contest, err := getContestCtx(db, c)
	if err != nil {
		return nil, err
	}

	// If the contest has not started, throw
	if contest.Contest.StartTime.After(time.Now()) {
		return nil, httperr.BadRequestf("Contest has not started")
	}

	// get contestType (weighted and unweighted)
	contestType := contest.Contest.ContestType

	// get contest's problems
	problems := contest.Problems

	users, err := models.GetAllUsers(db)
	if err != nil {
		return nil, err
	}

	contestProblemResults, err := models.CollectContestProblemResults(db, problems)
	if err != nil {
		return nil, err
	}

	userResults := []*UserResult{}
	userProblemResults := make(map[string]*UserResult)
	for _, user := range users {
		userProblemResults[user.ID] = &UserResult{
			User:           user,
			ProblemResults: make(map[int]*models.ProblemResult),
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
	problemBestSubmissions := make(map[int]int64)
	problemBestResults := make(map[int]*models.ProblemResult)

	for _, userProblemResult := range userProblemResults {
		// not consider users with no submissions and hidden users
		if len(userProblemResult.ProblemResults) == 0 || userProblemResult.User.Hidden {
			continue
		}

		problemResults := userProblemResult.ProblemResults
		for _, problemResult := range problemResults {
			problemID := problemResult.ProblemID
			bestResult, ok := problemBestResults[problemID]

			if !ok || compareProblemResult(problemResult, bestResult) {
				problemBestResults[problemID] = problemResult
				problemBestSubmissions[problemID] = problemResult.BestSubmissionID.Int64
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

	return &ScoreboardCtx{
		ContestCtx:             contest,
		UserResults:            userResults,
		ProblemBestSubmissions: problemBestSubmissions,
	}, nil
}

// ScoreboardGet implements GET /contest/:id/scoreboard
func (g *Group) ScoreboardGet(c echo.Context) error {
	ctx, err := getScoreboardCtx(g.db, c)
	if err != nil {
		return err
	}
	return ctx.Render(c)
}

// ScoreboardJSONGet implements GET /contest/:id/scoreboard/json
func (g *Group) ScoreboardJSONGet(c echo.Context) error {
	ctx, err := getScoreboardCtx(g.db, c)
	if err != nil {
		return err
	}
	return ctx.RenderJSON(c)
}
