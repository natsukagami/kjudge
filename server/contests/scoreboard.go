package contests

import (
	"net/http"
	"sort"
	"time"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models"
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

// ScoreboardCtx is the context required to display the scoreboard page
type ScoreboardCtx struct {
	*ContestCtx

	UsersCtx    []*UserResult
	ProblemsCtx []*models.Problem
}

// Render renders the scoreboard context
func (s *ScoreboardCtx) Render(c echo.Context) error {
	return c.Render(http.StatusOK, "contests/scoreboard", s)
}

// compareUserRanking checks if ranking of user[i] is strictly less than the ranking of user[j]
func compareUserRanking(userResult []*UserResult, contestType models.ContestType, i, j int) bool {
	if contestType == "weighted" {
		// sort based on totalScore if two users have same totalScore sort based on totalPenalty in an ascending order
		return (userResult[i].TotalScore > userResult[j].TotalScore ||
			(userResult[i].TotalScore == userResult[j].TotalScore && userResult[i].TotalPenalty < userResult[j].TotalPenalty))
	} else {
		// sort based on solvedProblems if two users have same solvedProblems sort based on totalPenalty in an ascending order
		return (userResult[i].SolvedProblems > userResult[j].SolvedProblems ||
			(userResult[i].SolvedProblems == userResult[j].SolvedProblems && userResult[i].TotalPenalty < userResult[j].TotalPenalty))
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
		return nil, echo.NewHTTPError(http.StatusBadRequest, "Contest has not started")
	}
	// get contestType (weighted and unweighted)
	contestType := contest.Contest.ContestType

	problems, err := models.GetContestProblems(db, contest.Contest.ID)
	if err != nil {
		return nil, err
	}

	users, err := models.GetAllUsers(db)
	if err != nil {
		return nil, err
	}

	userResults := []*UserResult{}
	contestProblemResults, err := models.CollectContestProblemResults(db, problems)
	if err != nil {
		return nil, err
	}

	userProblemResults := make(map[string]*UserResult)
	for _, user := range users {
		userProblemResults[user.ID] = &UserResult{
			User:           user,
			ProblemResults: make(map[int]*models.ProblemResult),
		}
		for _, problem := range problems {
			userProblemResults[user.ID].ProblemResults[problem.ID] = &models.ProblemResult{}
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

	sort.SliceStable(userResults, func(i, j int) bool {
		return compareUserRanking(userResults, contestType, i, j)
	})

	// after sorting users, we need to calculate users' ranking
	rank := 0
	for i, userCtx := range userResults {
		if i == 0 || compareUserRanking(userResults, contestType, i-1, i) {
			rank = rank + 1
		}
		userCtx.Rank = rank
	}

	return &ScoreboardCtx{
		ContestCtx:  contest,
		UsersCtx:    userResults,
		ProblemsCtx: problems,
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
