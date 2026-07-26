package main

import (
	"log/slog"
	"os"

	"github.com/opencorex-org/openrevenue/pkg/config"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	security, err := config.LoadSecurity()
	if err != nil {
		logger.Error("secure configuration rejected", "error", err)
		os.Exit(1)
	}
	logger.Info(
		"worker ready",
		"environment", security.Environment,
		"responsibilities", []string{"outbox", "notifications", "documents", "reconciliation", "reports"},
	)
}
