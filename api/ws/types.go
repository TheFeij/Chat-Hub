package ws

// Message represents a hub message
// all messages in the hub are transported in this type
type Message struct {
	Author string `json:"author"` // username of the client who wrote the text message
	Text   string `json:"text"`   // text of the message
}
