package webui

import (
	"embed"
	"fmt"
	"github.com/codemicro/analytics/analytics/config"
	"github.com/codemicro/analytics/analytics/db"
	"github.com/codemicro/analytics/analytics/webui/internal/templates"
	"github.com/flosch/pongo2/v6"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

type WebUI struct {
	conf *config.Config
	db   *db.DB

	app       *fiber.App
	templates *pongo2.TemplateSet
}

func init() {
	pongo2.RegisterFilter("shortTimeSince", func(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		tn := time.Now().UTC()
		t := in.Time()
		dur := tn.Sub(t).Round(time.Second)

		var (
			qty        int
			descriptor string
		)

		if int(dur.Minutes()) != 0 {
			qty = int(dur.Minutes())
			descriptor = "minute"
		} else if int(dur.Seconds()) > 30 {
			qty = int(dur.Seconds())
			descriptor = "second"
		} else {
			return pongo2.AsValue("just now"), nil
		}

		if qty != 1 {
			descriptor += "s"
		}

		return pongo2.AsValue(fmt.Sprintf("%d %s ago", qty, descriptor)), nil
	})
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

	wui.templates = pongo2.NewSet("templates", templates.TemplateLoader())

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

//go:embed static/*
var static embed.FS

func (wui *WebUI) registerHandlers() {
	wui.app.Get("/", wui.page_index)
	wui.app.Get("/partial/activeSessions", wui.partial_activeSessionsTable)

	wui.app.Get("/session", wui.page_logsFromSession)
	wui.app.Get("/partial/sessionLogs/:id", wui.partial_logsFromSession)

	wui.app.Get("/partial/topURLs", wui.partial_topURLs)

	wui.app.Use("/", filesystem.New(filesystem.Config{
		Root:       http.FS(static),
		PathPrefix: "static",
	}))
}
