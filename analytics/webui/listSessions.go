package webui

import (
	"fmt"
	"github.com/flosch/pongo2/v6"
	"github.com/gofiber/fiber/v2"
)

func (wui *WebUI) page_listSessions(ctx *fiber.Ctx) error {
	return wui.sendTemplate(ctx, "list-sessions.html", nil)
}

func (wui *WebUI) partial_listSessions(ctx *fiber.Ctx) error {
	ht := &HTMLTable{
		Path: "/partial/listSessions",
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
			sessions, err := wui.db.GetSessionsWithActivityAfter(60*24, sort)
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
