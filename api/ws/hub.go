package ws

import (
	"Chat-Server/repository"
	"log"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan Message

	// register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Hub messages
	messages []*Message
}

// NewHub creates and returns a new hub
func NewHub() *Hub {
	return &Hub{
		broadcast:  make(chan Message),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

// RunChatHub runs chat hub
func (h *Hub) RunChatHub(r repository.Repository) {
	// get all previous messages in the hub from the repository
	messages, _ := r.GetAllMessages()

	// initialize hub's messages with the size of retrieved messages from repository
	h.messages = make([]*Message, len(messages))

	// insert messages into hub messages slice
	for i := 0; i < len(messages); i++ {
		h.messages[i] = &Message{Author: messages[i].Author, Text: messages[i].Text}
	}

	for {
		select {
		case client := <-h.register:
			// initialize clients chat page with all previous messages
			err := client.WriteMessages(h.messages)
			if err != nil {
				continue
			}

			// register the client to the hub
			h.clients[client] = true

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}

		case message := <-h.broadcast:
			// save the message into the repository in a separate go routine
			go saveMessageToRepository(r, message)
			h.messages = append(h.messages, &message)

			// broadcast the new message to all the clients
			broadCastMessage(message, h.clients)
		}
	}
}

// saveMessageToRepository saves the input message into the repository
func saveMessageToRepository(r repository.Repository, message Message) {
	_, err := r.AddMessage(&repository.Message{
		Author: message.Author,
		Text:   message.Text,
	})
	if err != nil {
		log.Println(err)
	}
}

// broadCastMessage broadcast the input message to all clients
func broadCastMessage(message Message, clients map[*Client]bool) {
	// broadcast
	for client := range clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(clients, client)
		}
	}
}
