package http

import (
	"evrone_course_final/config"
	"fmt"
	"log/slog"
	"net/http"
)

func Serve(server *Server, cfg *config.Config) {
	router := http.NewServeMux()

	router.HandleFunc("GET /notifications/subscribe", server.SubscribeNotifications)

	srv := &http.Server{Handler: router, Addr: fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)}

	err := srv.ListenAndServe()
	if err != nil {
		slog.Error("srv.ListenAndServe", "err", err)
	}
}
