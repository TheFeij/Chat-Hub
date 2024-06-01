package api

import (
	"Chat-Server/token"
	"Chat-Server/token/mock"
	"Chat-Server/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// addAuthorization adds authorization token to the request header
func addAuthorization(
	t *testing.T,
	tokenMaker token.Maker,
	authorizationType string,
	username string,
	duration time.Duration,
	request *http.Request,
) (string, *token.Payload) {
	accessToken, payload, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, payload)
	require.NotEmpty(t, accessToken)

	authorizationHeader := fmt.Sprintf("%s %s", authorizationType, accessToken)
	request.Header.Set(authorizationHeaderKey, authorizationHeader)

	return accessToken, payload
}

// TestAuthMiddleware tests authMiddleware
func TestAuthMiddleware(t *testing.T) {

	var accessTokenPayload *token.Payload
	var accessToken string

	testCases := []struct {
		name          string
		setupAuth     func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs    func(tokenMaker *mockmaker.MockMaker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				accessToken, accessTokenPayload = addAuthorization(t, tokenMaker, authorizationTypeBearer, util.RandomUsername(), time.Minute, request)
			},
			buildStubs: func(tokenMaker *mockmaker.MockMaker) {
				tokenMaker.EXPECT().VerifyToken(accessToken).Times(1).Return(accessTokenPayload, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "NoAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
			},
			buildStubs: func(tokenMaker *mockmaker.MockMaker) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "UnSupportedAuthorization",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, "unsupported", util.RandomUsername(), time.Minute, request)
			},
			buildStubs: func(tokenMaker *mockmaker.MockMaker) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidAuthorizationFormat",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, tokenMaker, "", util.RandomUsername(), time.Minute, request)
			},
			buildStubs: func(tokenMaker *mockmaker.MockMaker) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ExpiredToken",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				accessToken, accessTokenPayload = addAuthorization(t, tokenMaker, authorizationTypeBearer, util.RandomUsername(), -time.Minute, request)
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

			tokenMaker, err := token.NewPasetoMaker(testConfigs.TokenSymmetricKey())
			require.NoError(t, err)
			testCase.setupAuth(t, request, tokenMaker)

			// in this test buildStubs must be called after the
			// setupAuth method so the accessToken and accessTokenPayload
			// variables are initialized
			testCase.buildStubs(mockTokenMaker)

			server.router.ServeHTTP(recorder, request)
			testCase.checkResponse(t, recorder)
		})
	}

}
