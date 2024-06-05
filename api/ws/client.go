package ws

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 1024
)

// Upgrader is a websocket Upgrader instance with the desired configurations
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client is an intermediary between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// username of the client
	username string

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan Message
}

// NewClient creates and returns a new Client object
func NewClient(hub *Hub, conn *websocket.Conn, send chan Message, username string) *Client {
	return &Client{hub: hub, conn: conn, send: send, username: username}
}

// Register the client to the hub
func (c *Client) Register() {
	c.hub.register <- c
}

// Read reads messages from the websocket connection and sends them to the hub.
func (c *Client) Read() {
	// unregister the client and close the connection after method done executing
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	// set websocket connection configurations
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// start reading loop
	for {
		_, text, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// create a Message instance and initialize it with the text and its author
		message := Message{
			Author: c.username,   // author will be client's username
			Text:   string(text), // text is the text read from the client
		}

		// send the message to the hub (hub will save the message and broadcast it to other clients)
		c.hub.broadcast <- message
	}
}

// Write receives messages from the hub and sends them to the client.
func (c *Client) Write() {
	// ticker is used to send ping messages periodically
	ticker := time.NewTicker(pingPeriod)

	// close ticker and close websocket connection after function done executing
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	// start write loop
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// the hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			jsonMessage, err := json.Marshal(message)
			if err != nil {
				return
			}
			w.Write(jsonMessage)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// WriteMessages write input messages to the client
func (c *Client) WriteMessages(messages []*Message) error {
	for _, message := range messages {
		err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
		if err != nil {
			return err
		}

		w, err := c.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}

		jsonMessage, err := json.Marshal(message)
		if err != nil {
			continue
		}
		w.Write(jsonMessage)

		if err := w.Close(); err != nil {
			return err
		}

	}

	return nil
}
