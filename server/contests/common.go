package contests

import (
	"database/sql"
	"strconv"

	"git.nkagami.me/natsukagami/kjudge/db"
	"git.nkagami.me/natsukagami/kjudge/models"
	"git.nkagami.me/natsukagami/kjudge/server/user"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
)

// ContestCtx implements a context that is common to all Contest UI page.
type ContestCtx struct {
	*user.AuthCtx

	Contest  *models.Contest
	Problems []*models.Problem
}

// Collect a contestctx from the echo Context.
func getContestCtx(db db.DBContext, c echo.Context) (*ContestCtx, error) {
	me, err := user.Me(db, c)
	if err != nil {
		return nil, err
	}
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return nil, echo.ErrNotFound
	}
	contest, err := models.GetContest(db, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, echo.ErrNotFound
	} else if err != nil {
		return nil, errors.WithStack(err)
	}
	problems, err := models.GetContestProblems(db, contest.ID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &ContestCtx{
		AuthCtx:  me,
		Contest:  contest,
		Problems: problems,
	}, nil
}
