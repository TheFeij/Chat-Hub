package repository

// The Message represents a repository message
type Message struct {
	// Text of the message
	Text string
	// Author of the message (username of the person who sent the message)
	Author string
}

// User represents a repository user
type User struct {
	// Username of the user
	Username string
	// Password of the user
	Password string
}
