package service

import (
	"fmt"
	"log/slog"

	"github.com/gorilla/websocket"
)

type WsNotificationsClient struct {
	conn     *websocket.Conn
	email    string
	send     chan []byte
	close    chan struct{}
	isClosed bool
}

func NewWsNotificationsClient(conn *websocket.Conn, email string) *WsNotificationsClient {
	c := &WsNotificationsClient{
		conn:  conn,
		email: email,
		send:  make(chan []byte, 256),
		close: make(chan struct{}),
	}
	go c.writePump()
	go c.readPump()
	return c
}

func (c *WsNotificationsClient) SendNotification(notificationBody []byte) error {
	if c.isClosed {
		return fmt.Errorf("client closed")
	}

	select {
	case c.send <- notificationBody:
		return nil
	default:
		c.Close()
		return fmt.Errorf("connection too slow")
	}
}

func (c *WsNotificationsClient) Close() {
	if c.isClosed {
		return
	}

	close(c.close)
	close(c.send)
	c.isClosed = true
}

func (c *WsNotificationsClient) Closed() chan struct{} {
	return c.close
}

func (c *WsNotificationsClient) writePump() {
	defer c.conn.Close() //nolint:errcheck

	for message := range c.send {
		c.conn.WriteMessage(websocket.TextMessage, message) //nolint:errcheck
	}
}

func (c *WsNotificationsClient) readPump() {
	for {
		select {
		case <-c.close:
			return
		default:
			messageType, _, err := c.conn.ReadMessage()
			if err != nil {
				slog.Error("Error in handleConnectionClosedByUser", slog.String("user_email", c.email), slog.String("error", err.Error()))
				c.Close()
				return
			}

			if messageType == websocket.CloseMessage {
				slog.Info("WS connection closed by user", slog.String("user_email", c.email))
				c.Close()
				return
			}
		}
	}
}
