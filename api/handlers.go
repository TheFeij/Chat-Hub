package api

import (
	"Chat-Server/api/ws"
	"Chat-Server/repository"
	"Chat-Server/token"
	"Chat-Server/util"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"net/http"
)

var InternalServerError = fmt.Errorf("something went wrong please try later")

// signup route handler
func (s *server) signup(context *gin.Context) {
	var req SignupRequest

	if err := context.ShouldBindJSON(&req); err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			for _, valErr := range valErrs {
				err = fmt.Errorf("invalid %s", valErr.Field())
				context.JSON(http.StatusBadRequest, errorResponse(err))
				return
			}
		}
		err = fmt.Errorf("invalid credentials")
		context.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		context.JSON(http.StatusInternalServerError, errorResponse(InternalServerError))
	}

	newUser, err := s.repository.AddUser(&repository.User{
		Username: req.Username,
		Password: hashedPassword,
	})
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			switch pgError.ConstraintName {
			case "users_pkey":
				err = fmt.Errorf("username already exists")
				context.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		context.JSON(http.StatusInternalServerError, errorResponse(InternalServerError))
		return
	}

	accessToken, accessTokenPayload, err := s.tokenMaker.CreateToken(
		newUser.Username,
		s.configs.AccessTokenDuration(),
	)
	if err != nil {
		context.JSON(http.StatusInternalServerError, errorResponse(InternalServerError))
		return
	}

	refreshToken, refreshTokenPayload, err := s.tokenMaker.CreateToken(
		newUser.Username,
		s.configs.RefreshTokenDuration())
	if err != nil {
		context.JSON(http.StatusInternalServerError, errorResponse(InternalServerError))
		return
	}

	// setting refresh token and access token in the cookies
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Expires:  refreshTokenPayload.ExpiredAt,
		Path:     s.configs.RefreshTokenCookiePath(),
		HttpOnly: true,
		Secure:   s.configs.IsProductionEnv(),
		SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		Expires:  accessTokenPayload.ExpiredAt,
		Path:     s.configs.AccessTokenCookiePath(),
		HttpOnly: true,
		Secure:   s.configs.IsProductionEnv(),
		SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     "username",
		Value:    newUser.Username,
		Expires:  accessTokenPayload.ExpiredAt,
		Path:     s.configs.UsernameCookiePath(),
		HttpOnly: false,
		Secure:   s.configs.IsProductionEnv(),
		SameSite: http.SameSiteStrictMode,
	})

	context.Status(http.StatusOK)
}

// login route handler
func (s *server) login(context *gin.Context) {
	var req LoginRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		var valErrs validator.ValidationErrors
		if errors.As(err, &valErrs) {
			for _, valErr := range valErrs {
				err = fmt.Errorf("invalid %s", valErr.Field())
				context.JSON(http.StatusBadRequest, errorResponse(err))
				return
			}
		}
		err = fmt.Errorf("invalid credentials")
		context.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := s.repository.GetUser(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = fmt.Errorf("this username doesn't have an account")
			context.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		context.JSON(http.StatusInternalServerError, errorResponse(InternalServerError))
		return
	}

	if err := util.CheckPassword(req.Password, user.Password); err != nil {
		err = fmt.Errorf("wrong passowrd")
		context.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, accessTokenPayload, err := s.tokenMaker.CreateToken(
		req.Username,
		s.configs.AccessTokenDuration(),
	)
	if err != nil {
		context.JSON(http.StatusInternalServerError, errorResponse(InternalServerError))
		return
	}

	refreshToken, refreshTokenPayload, err := s.tokenMaker.CreateToken(
		req.Username,
		s.configs.RefreshTokenDuration())
	if err != nil {
		context.JSON(http.StatusInternalServerError, errorResponse(InternalServerError))
		return
	}

	// setting refresh token and access token in the cookies
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Expires:  refreshTokenPayload.ExpiredAt,
		Path:     s.configs.RefreshTokenCookiePath(),
		HttpOnly: true,
		Secure:   s.configs.IsProductionEnv(),
		SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		Expires:  accessTokenPayload.ExpiredAt,
		Path:     s.configs.AccessTokenCookiePath(),
		HttpOnly: true,
		Secure:   s.configs.IsProductionEnv(),
		SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     "username",
		Value:    user.Username,
		Expires:  accessTokenPayload.ExpiredAt,
		Path:     s.configs.UsernameCookiePath(),
		HttpOnly: false,
		Secure:   s.configs.IsProductionEnv(),
		SameSite: http.SameSiteStrictMode,
	})

	context.Status(http.StatusOK)
}

// errorResponse puts the error into a gin.H instance
func errorResponse(err error) gin.H {
	return gin.H{
		"error": err.Error(),
	}
}

// refreshToken reads refresh token from the cookies, and if valid creates another access token for the client
func (s *server) refreshToken(context *gin.Context) {
	refreshToken, err := context.Cookie("refreshToken")
	if err != nil {
		err := fmt.Errorf("no refresh token provided")
		context.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	payload, err := s.tokenMaker.VerifyToken(refreshToken)
	if err != nil {
		if errors.Is(err, token.ErrExpiredToken) {
			err = fmt.Errorf("expired refresh token")
		} else {
			err = fmt.Errorf("invalid refresh token")
		}

		context.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	newAccessToken, newAccessTokenPayload, err := s.tokenMaker.CreateToken(
		payload.Username,
		s.configs.AccessTokenDuration(),
	)
	if err != nil {
		context.JSON(http.StatusInternalServerError, errorResponse(InternalServerError))
		return
	}

	http.SetCookie(context.Writer, &http.Cookie{
		Name:     "accessToken",
		Value:    newAccessToken,
		Expires:  newAccessTokenPayload.ExpiredAt,
		Path:     s.configs.AccessTokenCookiePath(),
		HttpOnly: true,
		Secure:   s.configs.IsProductionEnv(),
		SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     "username",
		Value:    payload.Username,
		Expires:  newAccessTokenPayload.ExpiredAt,
		Path:     s.configs.UsernameCookiePath(),
		HttpOnly: false,
		Secure:   s.configs.IsProductionEnv(),
		SameSite: http.SameSiteStrictMode,
	})

	context.Status(http.StatusOK)
}

// chat is the handler for the "/api/chat" route, starts a websocket connection to enter the chat server
func (s *server) chat(context *gin.Context) {
	// get access token payload to get username of the client from it
	accessTokenPayload := context.MustGet(authorizationPayloadKey).(*token.Payload)

	// upgrade the connection to a websocket connection
	conn, err := ws.Upgrader.Upgrade(context.Writer, context.Request, nil)
	if err != nil {
		context.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	defer conn.Close()

	// create a new client instance
	client := ws.NewClient(s.chatHub, conn, make(chan ws.Message, 10), accessTokenPayload.Username)
	client.Register()

	// start writing and reading logics
	go client.Write()
	client.Read()
}
