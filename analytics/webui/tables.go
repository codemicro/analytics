package webui

import (
	"fmt"
	"github.com/flosch/pongo2/v6"
	"github.com/gofiber/fiber/v2"
	"strings"
)

type HTMLTable struct {
	Path                 string
	Headers              []*HTMLTableHeader
	Data                 func(sortKey, sortDirection string) ([][]any, error)
	DefaultSortKey       string
	DefaultSortDirection string
	ShowNumberOfEntries  bool
}

type HTMLTableHeader struct {
	Name     string
	Slug     string
	Sortable bool
	Safe     bool
}

var unsetValue = pongo2.AsSafeValue(`<span class="italic">unset</span>`)

func (wui *WebUI) renderHTMLTable(ctx *fiber.Ctx, ht *HTMLTable) error {
	sortKey := ctx.Query("sortKey")
	sortDirection := strings.ToLower(ctx.Query("sortDir"))

	{
		var validatedSortKey string
		for _, header := range ht.Headers {
			if strings.EqualFold(header.Slug, sortKey) && header.Sortable {
				validatedSortKey = header.Slug
				break
			}
		}
		sortKey = validatedSortKey
	}

	if sortKey == "" {
		sortKey = ht.DefaultSortKey
	}

	if sortDirection == "" || !(sortDirection == "asc" || sortDirection == "desc") {
		fmt.Println("ere", sortDirection, sortKey)
		if sortKey == "" {
			sortDirection = ""
		} else {
			sortDirection = ht.DefaultSortDirection
		}
	}

	data, err := ht.Data(sortKey, sortDirection)
	if err != nil {
		return err
	}

	ctx.Type("html")
	return wui.sendTemplate(ctx, "components/table.html", pongo2.Context{
		"path":                ht.Path,
		"headers":             ht.Headers,
		"rows":                data,
		"sortKey":             sortKey,
		"sortDirection":       sortDirection,
		"showNumberOfEntries": ht.ShowNumberOfEntries,
	})
}

func getValue(v *pongo2.Value, _ *pongo2.Error) *pongo2.Value {
	return v
}
