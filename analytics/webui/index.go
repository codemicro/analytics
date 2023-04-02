package webui

import (
	"context"
	"fmt"
	"github.com/flosch/pongo2/v6"
	"github.com/gofiber/fiber/v2"
	"strconv"
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
			var sort string
			if !(sortKey == "" || sortDirection == "") {
				sort = sortKey + " " + sortDirection
			}
			sessions, err := wui.db.GetSessionsWithActivityAfter(30, sort)
			if err != nil {
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

func (wui *WebUI) partial_topURLs(ctx *fiber.Ctx) error {
	nStr := ctx.Query("n")
	var n int
	if nStr != "" {
		n, _ = strconv.Atoi(nStr)
	}

	hoursStr := ctx.Query("hours")
	var hours int
	if hoursStr != "" {
		hours, _ = strconv.Atoi(hoursStr)
	}

	ht := &HTMLTable{
		Headers: []*HTMLTableHeader{
			{"Count", "", false, false},
			{"Host", "", false, false},
			{"Path", "", false, false},
		},
		Data: func(sortKey, sortDirection string) ([][]any, error) {
			var counts []struct {
				Host  string
				URI   string
				Count int
			}
			q := wui.db.DB.NewSelect().
				ColumnExpr(`"host", "uri", COUNT(*) as "count"`).
				Table("requests").
				GroupExpr(`"host", "uri"`).
				OrderExpr(`"count" DESC`)
			if n > 0 {
				q = q.Limit(n)
			}
			if hours > 0 {
				q = q.Where(fmt.Sprintf(`datetime() < datetime(time, '+%d hours')`, hours))
			}
			if err := q.Scan(context.Background(), &counts); err != nil {
				return nil, err
			}

			var res [][]any
			for _, c := range counts {
				res = append(res, []any{c.Count, c.Host, c.URI})
			}

			return res, nil
		},
	}
	return wui.renderHTMLTable(ctx, ht)
}
