package webui

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/codemicro/analytics/analytics/db/models"
	"github.com/flosch/pongo2/v6"
	"github.com/gofiber/fiber/v2"
	"html"
)

func (wui *WebUI) page_logsFromSession(ctx *fiber.Ctx) error {
	id := ctx.Query("id")

	if id == "" {
		return fiber.ErrBadRequest
	}

	session := new(models.Session)
	if err := wui.db.DB.NewSelect().Model(session).Where("id = ?", id).Scan(context.Background(), session); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.ErrNotFound
		}
		return err
	}
	pctx := pongo2.Context{
		"session": session,
	}

	return wui.sendTemplate(ctx, "logs-from-session.html", pctx)
}

func (wui *WebUI) partial_logsFromSession(ctx *fiber.Ctx) error {
	id := ctx.Params("id")
	pa := fmt.Sprintf("/partial/sessionLogs/%s", html.EscapeString(id))

	ht := &HTMLTable{
		Path: pa,
		Headers: []*HTMLTableHeader{
			{"Datetime", "time", true, false},
			{"Host", "", false, false},
			{"Raw path", "raw_uri", true, false},
			{"Status", "status_code", true, false},
			{"Referer", "", false, false},
			{"", "", false, true},
		},
		Data: func(sortKey, sortDirection string) ([][]any, error) {
			var reqs []*models.Request
			q := wui.db.DB.NewSelect().Model(&reqs).Where("session_id = ?", id)
			if sortKey != "" {
				q = q.Order(sortKey + " " + sortDirection)
			}
			if err := q.Scan(context.Background(), &reqs); err != nil {
				return nil, err
			}

			var res [][]any
			for _, request := range reqs {
				res = append(res, []any{
					getValue(pongo2.ApplyFilter("date", pongo2.AsValue(request.Time), pongo2.AsValue("2006-01-02 15:04:05"))),
					request.Host,
					request.RawURI,
					request.StatusCode,
					getValue(pongo2.ApplyFilter("default", pongo2.AsValue(request.Referer), unsetValue)),
					fmt.Sprintf(`<a href="/request?id=%s">[Link]</a>`, request.ID),
				})
			}

			return res, nil
		},
		DefaultSortKey:       "time",
		DefaultSortDirection: "desc",
		ShowNumberOfEntries:  true,
	}
	return wui.renderHTMLTable(ctx, ht)
}
