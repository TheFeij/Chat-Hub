package api

import (
	"Chat-Server/repository"
	"Chat-Server/util"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"net/http"
)

// signup route handler
func (s *server) signup(context *gin.Context) {
	var req SignupRequest

	if err := context.ShouldBindJSON(&req); err != nil {
		context.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		context.JSON(http.StatusInternalServerError, errorResponse(err))
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
				context.JSON(http.StatusForbidden, errorResponse(err))
				return
			}
		}
		context.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	accessToken, accessTokenPayload, err := s.tokenMaker.CreateToken(
		newUser.Username,
		s.configs.AccessTokenDuration(),
	)
	if err != nil {
		context.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, refreshTokenPayload, err := s.tokenMaker.CreateToken(
		newUser.Username,
		s.configs.RefreshTokenDuration())
	if err != nil {
		context.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// setting refresh token and access token in the cookies
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Expires:  refreshTokenPayload.ExpiredAt,
		Path:     "/api/refresh",
		HttpOnly: true,
		//Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		Expires:  accessTokenPayload.ExpiredAt,
		Path:     "/api/chat",
		HttpOnly: true,
		//Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     "username",
		Value:    newUser.Username,
		Expires:  accessTokenPayload.ExpiredAt,
		Path:     "/api/chat",
		HttpOnly: false,
		//Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	context.Status(http.StatusOK)
}

// login route handler
func (s *server) login(context *gin.Context) {
	var req LoginRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		context.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	user, err := s.repository.GetUser(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			context.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		context.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if err := util.CheckPassword(req.Password, user.Password); err != nil {
		context.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	accessToken, accessTokenPayload, err := s.tokenMaker.CreateToken(
		req.Username,
		s.configs.AccessTokenDuration(),
	)
	if err != nil {
		context.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	refreshToken, refreshTokenPayload, err := s.tokenMaker.CreateToken(
		req.Username,
		s.configs.RefreshTokenDuration())
	if err != nil {
		context.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	// setting refresh token and access token in the cookies
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Expires:  refreshTokenPayload.ExpiredAt,
		Path:     "/api/refresh",
		HttpOnly: true,
		//Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		Expires:  accessTokenPayload.ExpiredAt,
		Path:     "/api/chat",
		HttpOnly: true,
		//Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     "username",
		Value:    user.Username,
		Expires:  accessTokenPayload.ExpiredAt,
		Path:     "/chat",
		HttpOnly: false,
		//Secure:   true,
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
		context.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	payload, err := s.tokenMaker.VerifyToken(refreshToken)
	if err != nil {
		context.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	newAccessToken, newAccessTokenPayload, err := s.tokenMaker.CreateToken(
		payload.Username,
		s.configs.AccessTokenDuration(),
	)
	if err != nil {
		context.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	http.SetCookie(context.Writer, &http.Cookie{
		Name:     "accessToken",
		Value:    newAccessToken,
		Expires:  newAccessTokenPayload.ExpiredAt,
		Path:     "/chat",
		HttpOnly: true,
		//Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})

	context.Status(http.StatusOK)
}
