package api

import (
	"Chat-Server/token"
	"Chat-Server/token/mock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func addTokenCookie(
	t *testing.T,
	username string,
	req *http.Request,
	cookieName string,
	duration time.Duration,
	cookiePath string,
) (string, *token.Payload) {
	authToken, payload := createToken(t, username, duration)
	req.AddCookie(&http.Cookie{
		Name:     cookieName,
		Value:    authToken,
		Path:     cookiePath,
		Expires:  payload.ExpiredAt,
		Secure:   false,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})

	return authToken, payload
}

// TestAuthMiddleware tests authMiddleware
func TestAuthMiddleware(t *testing.T) {
	randomUser, _ := randomUser(t)

	var accessTokenPayload *token.Payload
	var accessToken string

	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request)
		buildStubs    func(tokenMaker *mockmaker.MockMaker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request) {
				accessToken, accessTokenPayload = addTokenCookie(
					t,
					randomUser.Username,
					request,
					"accessToken",
					testConfigs.AccessTokenDuration(),
					testConfigs.AccessTokenCookiePath(),
				)
			},
			buildStubs: func(tokenMaker *mockmaker.MockMaker) {
				tokenMaker.EXPECT().VerifyToken(accessToken).Times(1).Return(accessTokenPayload, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "AccessTokenNotProvided",
			setupAuth: func(t *testing.T, request *http.Request) {
			},
			buildStubs: func(tokenMaker *mockmaker.MockMaker) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidToken",
			setupAuth: func(t *testing.T, request *http.Request) {
				accessToken = "invalid token"
				accessTokenPayload = nil

				request.AddCookie(&http.Cookie{
					Name:  "accessToken",
					Value: accessToken,
				})
			},
			buildStubs: func(tokenMaker *mockmaker.MockMaker) {
				tokenMaker.EXPECT().VerifyToken(accessToken).Times(1).Return(nil, token.ErrInvalidToken)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request) {
				accessToken, accessTokenPayload = addTokenCookie(
					t,
					randomUser.Username,
					request,
					"accessToken",
					-time.Minute,
					testConfigs.AccessTokenCookiePath(),
				)
			},
			buildStubs: func(tokenMaker *mockmaker.MockMaker) {
				tokenMaker.EXPECT().VerifyToken(accessToken).Times(1).Return(nil, token.ErrExpiredToken)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTokenMaker := mockmaker.NewMockMaker(ctrl)

			server := NewTestServer(t, nil, mockTokenMaker)
			authRoutes := server.router.Group("/").Use(authMiddleware(server.tokenMaker))
			authRoutes.GET("/auth",
				func(context *gin.Context) {
					context.JSON(http.StatusOK, gin.H{})
				},
			)

			recorder := httptest.NewRecorder()
			request, err := http.NewRequest(http.MethodGet, "/auth", nil)
			require.NoError(t, err)

			testCase.setupAuth(t, request)

			// in this test buildStubs must be called after the
			// setupAuth method so the accessToken and accessTokenPayload
			// variables are initialized
			testCase.buildStubs(mockTokenMaker)

			server.router.ServeHTTP(recorder, request)
			testCase.checkResponse(t, recorder)
		})
	}

}
