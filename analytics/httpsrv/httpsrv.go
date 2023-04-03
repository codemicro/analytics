package httpsrv

import (
	"github.com/codemicro/analytics/analytics/config"
	"github.com/codemicro/analytics/analytics/db"
	"github.com/flosch/pongo2/v6"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

type WebUI struct {
	conf *config.Config
	db   *db.DB

	app       *fiber.App
	templates *pongo2.TemplateSet
}

func Start(conf *config.Config, db *db.DB) *WebUI {
	wui := &WebUI{
		conf: conf,
		db:   db,
	}

	wui.app = fiber.New(fiber.Config{
		DisableStartupMessage: !config.Debug,
	})
	wui.registerHandlers()

	//wui.templates = pongo2.NewSet("templates", templates.TemplateLoader())

	go func() {
		if err := wui.app.Listen(conf.HTTP.Address); err != nil {
			log.Error().Err(err).Msg("HTTP server listen failed")
			return
		}
	}()

	log.Info().Msgf("HTTP server alive on %s", conf.HTTP.Address)
	return wui
}

func (wui *WebUI) sendTemplate(ctx *fiber.Ctx, fname string, renderCtx pongo2.Context) error {
	tpl, err := wui.templates.FromFile(fname)
	if err != nil {
		return err
	}
	res, err := tpl.ExecuteBytes(renderCtx)
	if err != nil {
		return err
	}
	ctx.Type("html")
	return ctx.Send(res)
}

func (wui *WebUI) Stop() error {
	return wui.app.ShutdownWithTimeout(time.Second * 5)
}

func (wui *WebUI) registerHandlers() {
	wui.app.Get("/", wui.index)
	wui.app.Use("/ds", func(ctx *fiber.Ctx) error {
		path := ctx.Path()
		path = strings.TrimPrefix(path, ctx.Route().Path)
		return proxy.Do(ctx, wui.conf.Datasette.Address+path)
	})
}
