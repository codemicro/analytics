package httpsrv

import (
	_ "embed"
	"github.com/gofiber/fiber/v2"
)

//go:embed index.html
var indexPage []byte

func (wui *WebUI) index(ctx *fiber.Ctx) error {
	ctx.Set(fiber.HeaderContentType, "text/html")
	return ctx.Send(indexPage)
}
