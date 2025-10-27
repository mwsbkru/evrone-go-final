package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/mwsbkru/evrone-go-final/config"
)

func Serve(ctx context.Context, server *Server, cfg *config.Config) {
	router := http.NewServeMux()

	router.HandleFunc("GET /notifications/subscribe", server.SubscribeNotifications(ctx))

	srv := &http.Server{Handler: router, Addr: fmt.Sprintf("%s:%s", cfg.WS.Host, cfg.WS.Port)}

	err := srv.ListenAndServe()
	if err != nil {
		slog.Error("srv.ListenAndServe", "err", err)
	}
}
