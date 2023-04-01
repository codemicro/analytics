package webui

import (
	"github.com/codemicro/analytics/analytics/config"
	"github.com/codemicro/analytics/analytics/db"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
	"time"
)

type WebUI struct {
	conf *config.Config
	db   *db.DB

	app *fiber.App
}

func Start(conf *config.Config, db *db.DB) *WebUI {
	wui := &WebUI{
		conf: conf,
		db:   db,
	}
	wui.app = fiber.New()
	wui.registerHandlers()
	go func() {
		if err := wui.app.Listen(conf.HTTP.Address); err != nil {
			log.Error().Err(err).Msg("HTTP server listen failed")
			return
		}
	}()
	return wui
}

func (wui *WebUI) Stop() error {
	return wui.app.ShutdownWithTimeout(time.Second * 5)
}

func (wui *WebUI) registerHandlers() {
	wui.app.Get("/", func(ctx *fiber.Ctx) error {
		return ctx.SendString("Hello! This is the HTTP server.")
	})
}