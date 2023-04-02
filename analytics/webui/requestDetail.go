package webui

import (
	"context"
	"database/sql"
	"errors"
	"github.com/codemicro/analytics/analytics/db/models"
	"github.com/flosch/pongo2/v6"
	"github.com/gofiber/fiber/v2"
)

func (wui *WebUI) page_requestDetail(ctx *fiber.Ctx) error {
	id := ctx.Query("id")
	if id == "" {
		return fiber.ErrBadRequest
	}

	request := new(models.Request)
	if err := wui.db.DB.NewSelect().Model(request).Where("id = ?", id).Scan(context.Background(), request); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.ErrNotFound
		}
		return err
	}
	pctx := pongo2.Context{
		"request": request,
	}

	return wui.sendTemplate(ctx, "request-detail.html", pctx)
}
