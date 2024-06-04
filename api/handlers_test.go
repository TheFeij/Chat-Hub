package api

import (
	"Chat-Server/repository"
	mockdb "Chat-Server/repository/mock"
	"Chat-Server/token"
	mockmaker "Chat-Server/token/mock"
	"Chat-Server/util"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// userMatcher custom gomock matcher for Repository.AddUser mock
type userMatcher struct {
	username string
	password string
}

func (s userMatcher) Matches(x any) bool {
	inputUser, ok := x.(*repository.User)
	if !ok {
		return false
	}

	err := util.CheckPassword(s.password, inputUser.Password)
	return err == nil && s.username == inputUser.Username
}

func (s userMatcher) String() string {
	return fmt.Sprintf("is equal to User")
}

func newUserMatcher(username, password string) gomock.Matcher {
	return userMatcher{
		username: username,
		password: password,
	}
}

// TestSignup tests signup route handler
func TestSignup(t *testing.T) {
	randomUser, password := randomUser(t)

	var accessTokenPayload *token.Payload
	var accessToken string
	var refreshTokenPayload *token.Payload
	var refreshToken string

	testCases := []struct {
		name       string
		req        SignupRequest
		buildStubs func(
			repository *mockdb.MockRepository,
			tokenMaker *mockmaker.MockMaker,
			req SignupRequest,
		)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder, tokenMaker token.Maker)
	}{
		{
			name: "OK",
			req: SignupRequest{
				Username: randomUser.Username,
				Password: password,
			},
			buildStubs: func(
				repository *mockdb.MockRepository,
				tokenMaker *mockmaker.MockMaker,
				req SignupRequest,
			) {
				repository.EXPECT().
					AddUser(newUserMatcher(randomUser.Username, password)).
					Times(1).
					Return(randomUser, nil)
				tokenMaker.EXPECT().
					CreateToken(req.Username, testConfigs.AccessTokenDuration()).
					Times(1).
					Return(accessToken, accessTokenPayload, nil)
				tokenMaker.EXPECT().
					CreateToken(req.Username, testConfigs.RefreshTokenDuration()).
					Times(1).
					Return(refreshToken, refreshTokenPayload, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, tokenMaker token.Maker) {
				require.Equal(t, http.StatusOK, recorder.Code)

				checkLoginResponse(t, randomUser.Username, accessToken, refreshToken, recorder)
			},
		},
		{
			name: "UsernameAlreadyExists",
			req: SignupRequest{
				Username: randomUser.Username,
				Password: password,
			},
			buildStubs: func(
				repository *mockdb.MockRepository,
				tokenMaker *mockmaker.MockMaker,
				req SignupRequest,
			) {
				repository.EXPECT().
					AddUser(newUserMatcher(randomUser.Username, password)).
					Times(1).
					Return(nil, &pgconn.PgError{ConstraintName: "users_pkey"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, tokenMaker token.Maker) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
		{
			name: "BadRequest",
			req: SignupRequest{
				Username: "invalid username", // has space
				Password: password,
			},
			buildStubs: func(
				repository *mockdb.MockRepository,
				tokenMaker *mockmaker.MockMaker,
				req SignupRequest,
			) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, tokenMaker token.Maker) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "DBInternalServerError",
			req: SignupRequest{
				Username: randomUser.Username,
				Password: password,
			},
			buildStubs: func(
				repository *mockdb.MockRepository,
				tokenMaker *mockmaker.MockMaker,
				req SignupRequest,
			) {
				repository.EXPECT().
					AddUser(newUserMatcher(randomUser.Username, password)).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, tokenMaker token.Maker) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "CreateAccessTokenInternalServerError",
			req: SignupRequest{
				Username: randomUser.Username,
				Password: password,
			},
			buildStubs: func(
				repository *mockdb.MockRepository,
				tokenMaker *mockmaker.MockMaker,
				req SignupRequest,
			) {
				repository.EXPECT().
					AddUser(newUserMatcher(randomUser.Username, password)).
					Times(1).
					Return(randomUser, nil)
				tokenMaker.EXPECT().
					CreateToken(req.Username, testConfigs.AccessTokenDuration()).
					Times(1).
					Return("", &token.Payload{}, errors.New("failed to encode payload to []byte"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, tokenMaker token.Maker) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "CreateRefreshTokenInternalServerError",
			req: SignupRequest{
				Username: randomUser.Username,
				Password: password,
			},
			buildStubs: func(
				repository *mockdb.MockRepository,
				tokenMaker *mockmaker.MockMaker,
				req SignupRequest,
			) {
				repository.EXPECT().
					AddUser(newUserMatcher(randomUser.Username, password)).
					Times(1).
					Return(randomUser, nil)
				tokenMaker.EXPECT().
					CreateToken(req.Username, testConfigs.AccessTokenDuration()).
					Times(1).
					Return(accessToken, accessTokenPayload, nil)
				tokenMaker.EXPECT().
					CreateToken(req.Username, testConfigs.RefreshTokenDuration()).
					Times(1).
					Return("", &token.Payload{}, errors.New("failed to encode payload to []byte"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, tokenMaker token.Maker) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			services := mockdb.NewMockRepository(ctrl)
			tokenMaker := mockmaker.NewMockMaker(ctrl)

			accessToken, accessTokenPayload = createToken(t, randomUser.Username, testConfigs.AccessTokenDuration())
			refreshToken, refreshTokenPayload = createToken(t, randomUser.Username, testConfigs.RefreshTokenDuration())

			testCase.buildStubs(services, tokenMaker, testCase.req)

			testServer := NewTestServer(t, services, tokenMaker)

			jsonReq, err := json.Marshal(&testCase.req)
			require.NoError(t, err)

			httpReq, err := http.NewRequest(http.MethodPost, "/api/signup", bytes.NewBuffer(jsonReq))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			testServer.router.ServeHTTP(recorder, httpReq)

			testCase.checkResponse(t, recorder, testServer.tokenMaker)
		})

	}

}

// TestLogin tests login route handler
func TestLogin(t *testing.T) {
	randomUser, password := randomUser(t)

	var accessTokenPayload *token.Payload
	var accessToken string
	var refreshTokenPayload *token.Payload
	var refreshToken string

	testCases := []struct {
		name       string
		req        LoginRequest
		buildStubs func(
			repository *mockdb.MockRepository,
			tokenMaker *mockmaker.MockMaker,
			req LoginRequest,
		)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder, tokenMaker token.Maker)
	}{
		{
			name: "OK",
			req: LoginRequest{
				Username: randomUser.Username,
				Password: password,
			},
			buildStubs: func(
				repository *mockdb.MockRepository,
				tokenMaker *mockmaker.MockMaker,
				req LoginRequest,
			) {
				repository.EXPECT().
					GetUser(gomock.Eq(req.Username)).
					Times(1).
					Return(randomUser, nil)
				tokenMaker.EXPECT().
					CreateToken(req.Username, testConfigs.AccessTokenDuration()).
					Times(1).
					Return(accessToken, accessTokenPayload, nil)
				tokenMaker.EXPECT().
					CreateToken(req.Username, testConfigs.RefreshTokenDuration()).
					Times(1).
					Return(refreshToken, refreshTokenPayload, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, tokenMaker token.Maker) {
				require.Equal(t, http.StatusOK, recorder.Code)

				checkLoginResponse(t, randomUser.Username, accessToken, refreshToken, recorder)
			},
		},
		{
			name: "NotFound",
			req: LoginRequest{
				Username: randomUser.Username,
				Password: password,
			},
			buildStubs: func(
				repository *mockdb.MockRepository,
				tokenMaker *mockmaker.MockMaker,
				req LoginRequest,
			) {
				repository.EXPECT().
					GetUser(gomock.Eq(req.Username)).
					Times(1).
					Return(nil, gorm.ErrRecordNotFound)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, tokenMaker token.Maker) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "BadRequest",
			req: LoginRequest{
				Username: "invalid username", // has space
				Password: password,
			},
			buildStubs: func(
				repository *mockdb.MockRepository,
				tokenMaker *mockmaker.MockMaker,
				req LoginRequest,
			) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, tokenMaker token.Maker) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "DBInternalServerError",
			req: LoginRequest{
				Username: randomUser.Username,
				Password: password,
			},
			buildStubs: func(
				repository *mockdb.MockRepository,
				tokenMaker *mockmaker.MockMaker,
				req LoginRequest,
			) {
				repository.EXPECT().
					GetUser(gomock.Eq(req.Username)).
					Times(1).
					Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, tokenMaker token.Maker) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "WrongPasswordUnAuthorized",
			req: LoginRequest{
				Username: randomUser.Username,
				Password: "wrongpassword",
			},
			buildStubs: func(
				repository *mockdb.MockRepository,
				tokenMaker *mockmaker.MockMaker,
				req LoginRequest,
			) {
				repository.EXPECT().
					GetUser(gomock.Eq(req.Username)).
					Times(1).
					Return(randomUser, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, tokenMaker token.Maker) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "CreateAccessTokenInternalServerError",
			req: LoginRequest{
				Username: randomUser.Username,
				Password: password,
			},
			buildStubs: func(
				repository *mockdb.MockRepository,
				tokenMaker *mockmaker.MockMaker,
				req LoginRequest,
			) {
				repository.EXPECT().
					GetUser(gomock.Eq(req.Username)).
					Times(1).
					Return(randomUser, nil)
				tokenMaker.EXPECT().
					CreateToken(req.Username, testConfigs.AccessTokenDuration()).
					Times(1).
					Return("", &token.Payload{}, errors.New("failed to encode payload to []byte"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, tokenMaker token.Maker) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "CreateRefreshTokenInternalServerError",
			req: LoginRequest{
				Username: randomUser.Username,
				Password: password,
			},
			buildStubs: func(
				repository *mockdb.MockRepository,
				tokenMaker *mockmaker.MockMaker,
				req LoginRequest,
			) {
				repository.EXPECT().
					GetUser(gomock.Eq(req.Username)).
					Times(1).
					Return(randomUser, nil)
				tokenMaker.EXPECT().
					CreateToken(req.Username, testConfigs.AccessTokenDuration()).
					Times(1).
					Return(accessToken, accessTokenPayload, nil)
				tokenMaker.EXPECT().
					CreateToken(req.Username, testConfigs.RefreshTokenDuration()).
					Times(1).
					Return("", &token.Payload{}, errors.New("failed to encode payload to []byte"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder, tokenMaker token.Maker) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			services := mockdb.NewMockRepository(ctrl)
			tokenMaker := mockmaker.NewMockMaker(ctrl)

			accessToken, accessTokenPayload = createToken(t, randomUser.Username, testConfigs.AccessTokenDuration())
			refreshToken, refreshTokenPayload = createToken(t, randomUser.Username, testConfigs.RefreshTokenDuration())

			testCase.buildStubs(services, tokenMaker, testCase.req)

			testServer := NewTestServer(t, services, tokenMaker)

			jsonReq, err := json.Marshal(&testCase.req)
			require.NoError(t, err)

			httpReq, err := http.NewRequest(http.MethodPost, "/api/login", bytes.NewBuffer(jsonReq))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			testServer.router.ServeHTTP(recorder, httpReq)

			testCase.checkResponse(t, recorder, testServer.tokenMaker)
		})

	}

}

// TestRefresh tests refreshToken route handler
func TestRefresh(t *testing.T) {
	randomUser, _ := randomUser(t)

	var refreshToken string
	var refreshTokenPayload *token.Payload

	var accessToken string
	var accessTokenPayload *token.Payload

	testCases := []struct {
		name          string
		req           *http.Request
		setupAuth     func(t *testing.T, request *http.Request)
		buildStubs    func(tokenMaker *mockmaker.MockMaker)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			setupAuth: func(t *testing.T, request *http.Request) {
				refreshToken, refreshTokenPayload = addTokenCookie(
					t,
					randomUser.Username,
					request,
					"refreshToken",
					testConfigs.RefreshTokenDuration(),
					testConfigs.RefreshTokenCookiePath(),
				)
			},
			buildStubs: func(tokenMaker *mockmaker.MockMaker) {
				tokenMaker.
					EXPECT().
					VerifyToken(refreshToken).
					Times(1).
					Return(refreshTokenPayload, nil)

				tokenMaker.
					EXPECT().
					CreateToken(refreshTokenPayload.Username, testConfigs.AccessTokenDuration()).
					Times(1).
					Return(accessToken, accessTokenPayload, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				checkLoginResponse(t, randomUser.Username, accessToken, refreshToken, recorder)
			},
		},
		{
			name: "RefreshTokenNotProvided",
			setupAuth: func(t *testing.T, request *http.Request) {
			},
			buildStubs: func(tokenMaker *mockmaker.MockMaker) {
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InvalidRefreshToken",
			setupAuth: func(t *testing.T, request *http.Request) {
				refreshToken = "invalid token"
				refreshTokenPayload = nil

				request.AddCookie(&http.Cookie{
					Name:  "refreshToken",
					Value: refreshToken,
				})
			},
			buildStubs: func(tokenMaker *mockmaker.MockMaker) {
				tokenMaker.
					EXPECT().
					VerifyToken(refreshToken).
					Times(1).
					Return(nil, token.ErrInvalidToken)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "ExpiredRefreshToken",
			setupAuth: func(t *testing.T, request *http.Request) {
				refreshToken, refreshTokenPayload = addTokenCookie(
					t,
					randomUser.Username,
					request,
					"refreshToken",
					-time.Minute,
					testConfigs.RefreshTokenCookiePath(),
				)
			},
			buildStubs: func(tokenMaker *mockmaker.MockMaker) {
				tokenMaker.
					EXPECT().
					VerifyToken(refreshToken).
					Times(1).
					Return(nil, token.ErrExpiredToken)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalServerError",
			setupAuth: func(t *testing.T, request *http.Request) {
				refreshToken, refreshTokenPayload = addTokenCookie(
					t,
					randomUser.Username,
					request,
					"refreshToken",
					testConfigs.RefreshTokenDuration(),
					testConfigs.RefreshTokenCookiePath(),
				)
			},
			buildStubs: func(tokenMaker *mockmaker.MockMaker) {
				tokenMaker.
					EXPECT().
					VerifyToken(refreshToken).
					Times(1).
					Return(refreshTokenPayload, nil)

				tokenMaker.
					EXPECT().
					CreateToken(refreshTokenPayload.Username, testConfigs.AccessTokenDuration()).
					Times(1).
					Return("", &token.Payload{}, errors.New("failed to encode payload to []byte"))
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				checkLoginResponse(t, randomUser.Username, accessToken, refreshToken, recorder)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			tokenMaker := mockmaker.NewMockMaker(controller)

			accessToken, accessTokenPayload = createToken(
				t,
				randomUser.Username,
				testConfigs.AccessTokenDuration(),
			)

			req, err := http.NewRequest(http.MethodPost, "/api/refresh", nil)
			require.NoError(t, err)

			testCase.setupAuth(t, req)

			testCase.buildStubs(tokenMaker)

			server := NewTestServer(t, nil, tokenMaker)

			recorder := httptest.NewRecorder()
			server.router.ServeHTTP(recorder, req)

			// check response
			testCase.checkResponse(t, recorder)
		})
	}
}

// randomUser creates a random user
func randomUser(t *testing.T) (*repository.User, string) {
	password := util.RandomPassword()
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	return &repository.User{
		Username: util.RandomUsername(),
		Password: hashedPassword,
	}, password
}

// createToken creates a token and returns it with its payload
func createToken(t *testing.T, username string, duration time.Duration) (string, *token.Payload) {
	tokenMaker, err := token.NewPasetoMaker(testConfigs.TokenSymmetricKey())
	require.NoError(t, err)

	accessToken, payload, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, accessToken)
	require.NotEmpty(t, payload)

	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, time.Now().Add(duration), payload.ExpiredAt, time.Second)
	require.WithinDuration(t, time.Now(), payload.IssuedAt, time.Second)

	return accessToken, payload
}

// checkLoginResponse checks login response
func checkLoginResponse(t *testing.T, username, accessToken, refreshToken string, recorder *httptest.ResponseRecorder) {
	cookies := recorder.Result().Cookies()

	for _, cookie := range cookies {
		if cookie.Name == "accessToken" {
			require.Equal(t, accessToken, cookie.Value)
			require.WithinDuration(t, time.Now().Add(15*time.Minute), cookie.Expires, time.Second)
			require.True(t, cookie.HttpOnly)
			require.Equal(t, testConfigs.AccessTokenCookiePath(), cookie.Path)
		} else if cookie.Name == "refreshToken" {
			require.Equal(t, refreshToken, cookie.Value)
			require.WithinDuration(t, time.Now().Add(24*time.Hour), cookie.Expires, time.Second)
			require.True(t, cookie.HttpOnly)
			require.Equal(t, testConfigs.RefreshTokenCookiePath(), cookie.Path)
		} else if cookie.Name == "username" {
			require.Equal(t, username, cookie.Value)
			require.WithinDuration(t, time.Now().Add(15*time.Minute), cookie.Expires, time.Second)
			require.False(t, cookie.HttpOnly)
			require.Equal(t, testConfigs.UsernameCookiePath(), cookie.Path)
		}
		require.Equal(t, http.SameSiteStrictMode, cookie.SameSite)
	}
}
