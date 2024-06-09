package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"sync"
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
	fmt.Printf("client %d connected\n", m.clients[client])
}

// Removes existing websocket client from the set (e.g. on disconnect).
func (m *WsClients) remove(client *WsClient) {
	m.Lock()
	defer m.Unlock()
	fmt.Printf("client %d disconnected\n", m.clients[client])
	delete(m.clients, client)
}

// Sends the message to all currenly connected websocket clients.
func (m *WsClients) broadcast(msg string) {
	m.Lock()
	defer m.Unlock()
	for c, v := range m.clients {
		fmt.Printf("sending message to client %d\n", v)
		err := c.send(msg)
		if err != nil {
			fmt.Printf("send to client %d failed, ignoring: %s\n", v, err)
		}
	}
}
