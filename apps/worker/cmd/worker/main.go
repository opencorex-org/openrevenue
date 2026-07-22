package main

import (
	"log/slog"
	"os"
)

func main() {
	slog.New(slog.NewJSONHandler(os.Stdout, nil)).Info("worker ready", "responsibilities", []string{"outbox", "notifications", "documents", "reconciliation", "reports"})
}
