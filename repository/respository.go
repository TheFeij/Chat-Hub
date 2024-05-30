package repository

// Repository implements the required methods for the business layer to interact with the data layer
type Repository interface {
	// AddMessage adds a message to the data layer
	AddMessage(message *Message) (*Message, error)

	// GetAllMessages retrieves all messages from the database
	GetAllMessages() ([]*Message, error)
}
