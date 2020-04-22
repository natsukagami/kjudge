package admin

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/server/httperr"
	"github.com/pkg/errors"
)

// ClarificationsCtx is the context for rendering clarifications.
type ClarificationsCtx struct {
	Clarifications []*models.Clarification
	Problems       map[int]*models.Problem
	Contests       map[int]*models.Contest
}

// Render renders the context.
func (ctx *ClarificationsCtx) Render(c echo.Context) error {
	return c.Render(http.StatusOK, "admin/clarifications", ctx)
}

func getClarificationsCtx(db db.DBContext, c echo.Context) (*ClarificationsCtx, error) {
	clars, err := models.GetAllClarifications(db)
	if err != nil {
		return nil, err
	}
	var problemIDs, contestIDs []int
	for _, c := range clars {
		if c.ProblemID.Valid {
			problemIDs = append(problemIDs, int(c.ProblemID.Int64))
		}
		contestIDs = append(contestIDs, c.ContestID)
	}
	problems, err := models.CollectProblemsByID(db, problemIDs...)
	if err != nil {
		return nil, err
	}
	contests, err := models.CollectContestsByID(db, contestIDs...)
	if err != nil {
		return nil, err
	}
	return &ClarificationsCtx{
		Clarifications: clars,
		Problems:       problems,
		Contests:       contests,
	}, nil
}

// ClarificationsGet implements GET /admin/clarifications
func (g *Group) ClarificationsGet(c echo.Context) error {
	ctx, err := getClarificationsCtx(g.db, c)
	if err != nil {
		return err
	}
	if c.QueryParam("unanswered") == "true" {
		return c.JSON(http.StatusOK, ctx.getUnansweredClarificationsCount())
	}
	return ctx.Render(c)
}

// ClarificationReplyForm is a form for replying a clarification.
type ClarificationReplyForm struct {
	Response string
}

// ClarificationReplyPost implements POST /admin/clarifications/:id
func (g *Group) ClarificationReplyPost(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return httperr.NotFoundf("Clarification not found: %s", c.Param("id"))
	}
	tx, err := g.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer db.Rollback(tx)

	clar, err := models.GetClarification(tx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return httperr.NotFoundf("Clarification not found: %s", c.Param("id"))
	} else if err != nil {
		return err
	}

	if clar.Responded() {
		return httperr.BadRequestf("Clarification has already been responded.")
	}
	var form ClarificationReplyForm
	if err := c.Bind(&form); err != nil {
		return httperr.BindFail(err)
	}
	clar.Response = []byte(form.Response)
	clar.UpdatedAt = time.Now()

	if err := clar.Write(tx); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return errors.WithStack(err)
	}

	return c.Redirect(http.StatusSeeOther, "/admin/clarifications")
}

func (ctx *ClarificationsCtx) getUnansweredClarificationsCount() int {
	count := 0
	for _, c := range ctx.Clarifications {
		if !c.Responded() {
			count++
		}
	}
	return count
}
