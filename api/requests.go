package api

// SignupRequest represents a signup request body
type SignupRequest struct {
	Username string `json:"username" binding:"required,validUsername"`
	Password string `json:"password" binding:"required,validPassword"`
}

// LoginRequest represents a login request body
type LoginRequest struct {
	Username string `json:"username" binding:"required,validUsername"`
	Password string `json:"password" binding:"required,validPassword"`
}
