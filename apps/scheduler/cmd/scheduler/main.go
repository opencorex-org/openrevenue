package main

import (
	"log/slog"
	"os"
)

func main() {
	slog.New(slog.NewJSONHandler(os.Stdout, nil)).Info("scheduler ready", "responsibilities", []string{"deadlines", "penalties", "interest", "compliance", "retention", "reports"})
}
