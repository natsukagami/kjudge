package contests

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/natsukagami/kjudge/db"
	"github.com/natsukagami/kjudge/models"
	"github.com/natsukagami/kjudge/server/httperr"
	"github.com/pkg/errors"
)

// ProblemCtx is the context required to display a problem page.
type ProblemCtx struct {
	*ContestCtx

	Problem     *models.Problem
	Files       map[string]*models.File
	Submissions []*models.Submission
}

// Render renders the context.
func (p *ProblemCtx) Render(c echo.Context) error {
	return c.Render(http.StatusOK, "contests/problem", p)
}

// Collect a problemCtx.
func getProblemCtx(db db.DBContext, c echo.Context) (*ProblemCtx, error) {
	contest, err := getContestCtx(db, c)
	if err != nil {
		return nil, err
	}

	// If the contest has not started, throw
	if contest.Contest.StartTime.After(time.Now()) {
		return nil, httperr.BadRequestf("Contest has not started")
	}

	name := c.Param("problem")
	problem, err := models.GetProblemByName(db, contest.Contest.ID, name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, httperr.NotFoundf("Problem not found: %s", name)
	} else if err != nil {
		return nil, err
	}
	files, err := models.GetProblemFilesMeta(db, problem.ID)
	if err != nil {
		return nil, err
	}
	fm := make(map[string]*models.File)
	for _, f := range files {
		if f.Public {
			fm[f.Filename] = f
		}
	}
	subs, err := models.GetUserProblemSubmissions(db, contest.Me.ID, problem.ID)
	if err != nil {
		return nil, err
	}
	return &ProblemCtx{
		ContestCtx:  contest,
		Problem:     problem,
		Files:       fm,
		Submissions: subs,
	}, nil
}

// ProblemGet implements GET /contest/:id/problems/:problem
func (g *Group) ProblemGet(c echo.Context) error {
	ctx, err := getProblemCtx(g.db, c)
	if err != nil {
		return err
	}
	return ctx.Render(c)
}

// FileGet implements GET /contest/:id/problems/:problem/files/:file
func (g *Group) FileGet(c echo.Context) error {
	ctx, err := getProblemCtx(g.db, c)
	if err != nil {
		return err
	}
	fileIDStr := c.Param("file")
	fileID, err := strconv.Atoi(fileIDStr)
	if err != nil {
		return httperr.NotFoundf("File not found: %v", fileID)
	}
	file, err := models.GetFile(g.db, fileID)
	if errors.Is(err, sql.ErrNoRows) {
		return httperr.NotFoundf("File not found: %v", fileID)
	} else if err != nil {
		return err
	}

	if file.ProblemID != ctx.Problem.ID || !file.Public {
		return httperr.NotFoundf("File not found: %v", fileID)
	}
	http.ServeContent(c.Response(), c.Request(), file.Filename, time.Now(), bytes.NewReader(file.Content))
	return nil
}

// SubmitPost implements POST /contest/:id/problems/:problem/submit.
func (g *Group) SubmitPost(c echo.Context) error {
	tx, err := g.db.Beginx()
	if err != nil {
		return errors.WithStack(err)
	}
	defer db.Rollback(tx)

	now := time.Now()

	ctx, err := getProblemCtx(tx, c)
	if err != nil {
		return err
	}

	if ctx.Contest.EndTime.Before(time.Now()) {
		return httperr.BadRequestf("Contest has already ended")
	}

	// Check limits
	if ctx.Problem.MaxSubmissionsCount > 0 && len(ctx.Submissions) >= ctx.Problem.MaxSubmissionsCount {
		return httperr.Newf(http.StatusTooManyRequests, "You can only submit %d submissions for this problem", ctx.Problem.MaxSubmissionsCount)
	}
	if len(ctx.Submissions) > 0 {
		lastSub := ctx.Submissions[0]
		if now.Sub(lastSub.SubmittedAt) < time.Duration(ctx.Problem.SecondsBetweenSubmissions)*time.Second {
			return httperr.Newf(http.StatusTooManyRequests, "Please wait %d seconds between submissions for this problem", ctx.Problem.SecondsBetweenSubmissions)
		}
	}

	form, err := c.MultipartForm()
	if err != nil {
		return err
	}
	files, ok := form.File["file"]
	if !ok || len(files) != 1 {
		return httperr.BadRequestf("One file must be attached")
	}
	file := files[0]
	lang, err := models.LanguageByExt(filepath.Ext(file.Filename))
	if err != nil {
		return httperr.BadRequestf("Cannot resolve language: %v", err)
	}
	fileContent, err := file.Open()
	if err != nil {
		return errors.WithStack(err)
	}
	defer fileContent.Close()
	source, err := io.ReadAll(fileContent)
	if err != nil {
		return errors.WithStack(err)
	}

	// Submitted files can be executable
	if models.IsTextFile(file.Filename) {
		source, err = models.NormalizeEndingsUnix(source)
		if err != nil {
			return err
		}
	}
	sub := models.Submission{
		ProblemID:   ctx.Problem.ID,
		UserID:      ctx.Me.ID,
		Source:      source,
		Language:    lang,
		SubmittedAt: now,
		Verdict:     models.VerdictIsInQueue,
	}

	if err := sub.Write(tx); err != nil {
		return err
	}

	job := models.NewJobScore(sub.ID)
	if err := job.Write(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errors.WithStack(err)
	}
	return c.Redirect(http.StatusSeeOther, fmt.Sprintf("/contests/%d/problems/%s#submissions", ctx.Contest.ID, ctx.Problem.Name))
}
