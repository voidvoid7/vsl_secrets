package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/voidvoid7/vsl_secrets/internal/requesthandler"
	"github.com/voidvoid7/vsl_secrets/internal/secretmap"
)

func main() {
	secretMapHolder := secretmap.NewSecretMapHolder()
	requestHandler, err := requesthandler.NewRequestHandler(secretMapHolder)
	if err != nil {
		slog.Error("Failed to create request handler", "error", err)
		os.Exit(1)
		return
	}
	requestHandler.RegisterHandlers()
	slog.Info("Application starting on port :8080")
	slog.Error("Server crashed", "error", http.ListenAndServe(":8080", nil))
}
