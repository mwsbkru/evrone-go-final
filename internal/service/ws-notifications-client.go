package service

import (
	"fmt"
	"log/slog"

	"github.com/gorilla/websocket"
)

type WsNotificationsClient struct {
	conn  *websocket.Conn
	email string
	send  chan []byte
	close chan struct{}
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
	select {
	case c.send <- notificationBody:
		return nil
	case <-c.close:
		close(c.send)
		return fmt.Errorf("client closed")
	default:
		close(c.send)
		return fmt.Errorf("connection too slow")
	}
}

func (c *WsNotificationsClient) Close() {
	if c.close == nil {
		return
	}

	close(c.close)
	c.close = nil
	c.conn.Close()
}

func (c *WsNotificationsClient) Closed() struct{} {
	return <-c.close
}

func (c *WsNotificationsClient) writePump() {
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.Close()
				return
			}
			c.conn.WriteMessage(websocket.TextMessage, message) //nolint:errcheck
		case <-c.close:
			return
		}
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
