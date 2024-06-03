package api

import (
	"Chat-Server/token"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	authorizationCookieName string = "accessToken"
	authorizationPayloadKey string = "authorization_payload"
)

// authMiddleware checks for access token in the cookies and if valid, extracts the token payload
// and saves it as authorizationPayloadKey in the context
func authMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(context *gin.Context) {
		accessToken, err := context.Cookie(authorizationCookieName)

		if err != nil {
			err := fmt.Errorf("access token not provided")
			context.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			context.AbortWithStatusJSON(http.StatusUnauthorized, errorResponse(err))
			return
		}

		context.Set(authorizationPayloadKey, payload)
		context.Next()
	}
}
