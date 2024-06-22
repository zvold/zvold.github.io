package main

import (
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
)

// WsClient is a mutex-protected pointer to `websocket.Conn`.
type WsClient struct {
	sync.Mutex
	conn *websocket.Conn
}

// Writes the message to underlying `websocket.Conn` in a thread-safe way.
func (client *WsClient) send(msg string) error {
	client.Lock()
	defer client.Unlock()
	return client.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

// WsClients is a mutex-protected set of all connected websocket clients.
type WsClients struct {
	sync.Mutex
	clients map[*WsClient]int
}

// Adds a new websocket client to the set.
func (m *WsClients) add(client *WsClient) {
	m.Lock()
	defer m.Unlock()
	m.clients[client] = len(m.clients) + 1
	slog.Debug("client connected.", "client", m.clients[client])
}

// Removes existing websocket client from the set (e.g. on disconnect).
func (m *WsClients) remove(client *WsClient) {
	m.Lock()
	defer m.Unlock()
	slog.Debug("client disconnected.", "client", m.clients[client])
	delete(m.clients, client)
}

// Sends the message to all currenly connected websocket clients.
func (m *WsClients) broadcast(msg string) {
	m.Lock()
	defer m.Unlock()
	for c, v := range m.clients {
		slog.Debug("sending a message.", "client", v)
		err := c.send(msg)
		if err != nil {
			slog.Error("send failed, ignoring.", "client", v, "error", err)
		}
	}
}
