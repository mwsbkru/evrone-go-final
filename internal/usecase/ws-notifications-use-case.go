package usecase

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
)

type WsNotificationsUseCase struct {
	connections map[string]*websocket.Conn
}

func NewWsNotificationsUseCase() *WsNotificationsUseCase {
	return &WsNotificationsUseCase{connections: make(map[string]*websocket.Conn)}
}

func (u *WsNotificationsUseCase) HandleConnection(userEmail string, connection *websocket.Conn) {
	u.connections[userEmail] = connection
	go u.handleConnection(userEmail, connection)
}

func (u *WsNotificationsUseCase) handleConnection(userEmail string, connection *websocket.Conn) {
	defer connection.Close()

	for {
		messageType, receivedMessage, err := connection.ReadMessage()
		if err != nil {
			log.Println(err)
			break
		}

		for _, userConnection := range u.connections {
			messageText := fmt.Sprintf("Message from %s: %s; type: %v", userEmail, string(receivedMessage), messageType)
			userConnection.WriteMessage(websocket.TextMessage, []byte(messageText))
		}
	}
}
