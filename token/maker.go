package token

import "time"

// Maker defines methods used for creating and verifying authorization tokens
type Maker interface {
	CreateToken(username string, duration time.Duration) (string, *Payload, error)
	VerifyToken(token string) (*Payload, error)
}
