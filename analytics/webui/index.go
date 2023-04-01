package webui

import (
	"context"
	"fmt"
	"github.com/codemicro/analytics/analytics/db/models"
	"github.com/flosch/pongo2/v6"
	"github.com/gofiber/fiber/v2"
)

func (wui *WebUI) page_index(ctx *fiber.Ctx) error {
	return wui.sendTemplate(ctx, "index.html", nil)
}

func (wui *WebUI) partial_activeSessionsTable(ctx *fiber.Ctx) error {
	ht := &HTMLTable{
		Path: "/partial/activeSessions",
		Headers: []*HTMLTableHeader{
			{"", "", false, true},
			{"User agent", "", false, false},
			{"IP", "", false, false},
			{"Last seen", "last_seen", true, false},
		},
		Data: func(sortKey, sortDirection string) ([][]any, error) {
			var sessions []*models.Session
			q := wui.db.DB.NewSelect().
				Model((*models.Session)(nil)).
				ColumnExpr("*").
				ColumnExpr(`(select max("time") as "time" from requests where session_id = "session"."id") as "last_seen"`).
				Where(`datetime() < datetime("last_seen", '+30 minutes')`)
			if sortKey != "" {
				q = q.Order(sortKey + " " + sortDirection)
			}
			if err := q.Scan(context.Background(), &sessions); err != nil {
				return nil, err
			}

			var res [][]any
			for _, sess := range sessions {
				ua, _ := pongo2.ApplyFilter("truncatechars", pongo2.AsValue(sess.UserAgent), pongo2.AsValue(40))
				ua, _ = pongo2.ApplyFilter("default", ua, unsetValue)
				res = append(res, []any{
					fmt.Sprintf(`<a href="/session?id=%s">[Link]</a>`, sess.ID),
					ua,
					getValue(pongo2.ApplyFilter("truncatechars", pongo2.AsValue(sess.IPAddr), pongo2.AsValue(30))),
					getValue(pongo2.ApplyFilter("shortTimeSince", pongo2.AsValue(sess.LastSeen), nil)),
				})
			}

			return res, nil
		},
		DefaultSortKey:       "last_seen",
		DefaultSortDirection: "desc",
		ShowNumberOfEntries:  true,
	}
	return wui.renderHTMLTable(ctx, ht)
}
