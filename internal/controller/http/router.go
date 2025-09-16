package http

import (
	"context"
	"evrone_course_final/config"
	"fmt"
	"log/slog"
	"net/http"
)

func Serve(ctx context.Context, server *Server, cfg *config.Config) {
	router := http.NewServeMux()

	router.HandleFunc("GET /notifications/subscribe", server.SubscribeNotifications(ctx))

	srv := &http.Server{Handler: router, Addr: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)}

	err := srv.ListenAndServe()
	if err != nil {
		slog.Error("srv.ListenAndServe", "err", err)
	}
}
