package main

import (
	api "github.com/opencorex-org/openrevenue/apps/api"
	app "github.com/opencorex-org/openrevenue/internal/administration/application"
	notification "github.com/opencorex-org/openrevenue/internal/notification/infrastructure"
	"github.com/opencorex-org/openrevenue/pkg/config"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	security, err := config.LoadSecurity()
	if err != nil {
		logger.Error("secure configuration rejected", "error", err)
		os.Exit(1)
	}
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	logger.Info("api starting", "addr", addr, "environment", security.Environment)
	smtpAddress := os.Getenv("SMTP_ADDRESS")
	if smtpAddress == "" {
		smtpAddress = "localhost:1025"
	}
	notifier := notification.SMTPNotifier{Address: smtpAddress, From: "no-reply@openrevenue.local"}
	if err := http.ListenAndServe(addr, api.Router(app.New(notifier))); err != nil {
		logger.Error("api stopped", "error", err)
		os.Exit(1)
	}
}
