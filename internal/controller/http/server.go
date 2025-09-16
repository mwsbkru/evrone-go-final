package http

import (
	"context"
	"encoding/json"
	"evrone_course_final/config"
	"evrone_course_final/internal/entity/dto"
	"evrone_course_final/internal/usecase"
	"log/slog"
	"net/http"

	websocket "github.com/gorilla/websocket"
)

type Server struct {
	cfg                    *config.Config
	wsNotificationsUseCase *usecase.WsNotificationsUseCase
	upgrader               *websocket.Upgrader
}

func NewServer(cfg *config.Config, wsNotificationsUseCase *usecase.WsNotificationsUseCase) *Server {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			return !cfg.WsCheckOrigin || origin == cfg.WsAllowedOrigin
		},
	}
	return &Server{cfg: cfg, wsNotificationsUseCase: wsNotificationsUseCase, upgrader: &upgrader}
}

func (s *Server) SubscribeNotifications(ctx context.Context) func(http.ResponseWriter, *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		userEmail := request.URL.Query().Get("userEmail")
		if userEmail == "" {
			s.respondWithError(writer, http.StatusBadRequest, "get param userEmail must be present")
			return
		}

		ws, err := s.upgrader.Upgrade(writer, request, nil)
		if err != nil {
			slog.Error("can`t prepare WS connection", slog.String("error", err.Error()))
			return
		}

		s.wsNotificationsUseCase.HandleConnection(ctx, userEmail, ws)
	}
}

func (s *Server) respondWithError(writer http.ResponseWriter, code int, message string) {
	errorObject := dto.ErrorResponse{
		Code:    code,
		Message: message,
	}

	responseBody, err := json.Marshal(&errorObject)
	if err != nil {
		slog.Error("can`t prepare HTTP response error", slog.String("error", err.Error()))
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.WriteHeader(code)
	writer.Write(responseBody)
}
