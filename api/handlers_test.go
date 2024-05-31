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
	"io"
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
				requireBodyMatchLogin(t, recorder.Body, randomUser, testConfigs.TokenSymmetricKey(), refreshToken, accessToken)
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

			accessToken, accessTokenPayload = createToken(t, randomUser.Username, testConfigs.AccessTokenDuration(), testConfigs.TokenSymmetricKey())
			refreshToken, refreshTokenPayload = createToken(t, randomUser.Username, testConfigs.RefreshTokenDuration(), testConfigs.TokenSymmetricKey())

			testCase.buildStubs(services, tokenMaker, testCase.req)

			testServer := NewTestServer(t, services, tokenMaker)

			jsonReq, err := json.Marshal(&testCase.req)
			require.NoError(t, err)

			httpReq, err := http.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(jsonReq))
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
				requireBodyMatchLogin(t, recorder.Body, randomUser, testConfigs.TokenSymmetricKey(), refreshToken, accessToken)
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

			accessToken, accessTokenPayload = createToken(t, randomUser.Username, testConfigs.AccessTokenDuration(), testConfigs.TokenSymmetricKey())
			refreshToken, refreshTokenPayload = createToken(t, randomUser.Username, testConfigs.RefreshTokenDuration(), testConfigs.TokenSymmetricKey())

			testCase.buildStubs(services, tokenMaker, testCase.req)

			testServer := NewTestServer(t, services, tokenMaker)

			jsonReq, err := json.Marshal(&testCase.req)
			require.NoError(t, err)

			httpReq, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonReq))
			require.NoError(t, err)

			recorder := httptest.NewRecorder()
			testServer.router.ServeHTTP(recorder, httpReq)

			testCase.checkResponse(t, recorder, testServer.tokenMaker)
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
func createToken(t *testing.T, username string, duration time.Duration, tokenSymmetricKey string) (string, *token.Payload) {
	tokenMaker, err := token.NewPasetoMaker(tokenSymmetricKey)
	require.NoError(t, err)

	accessToken, payload, err := tokenMaker.CreateToken(username, duration)
	require.NoError(t, err)
	require.NotEmpty(t, accessToken)
	require.NotEmpty(t, payload)

	require.NoError(t, payload.Valid())
	require.Equal(t, username, payload.Username)
	require.WithinDuration(t, time.Now().Add(duration), payload.ExpiredAt, time.Second)
	require.WithinDuration(t, time.Now(), payload.IssuedAt, time.Second)

	return accessToken, payload
}

// requireBodyMatchLogin checks login response body
func requireBodyMatchLogin(
	t *testing.T,
	body *bytes.Buffer,
	user *repository.User,
	tokenSymmetricKey string,
	originalRefreshToken string,
	originalAccessToken string,
) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var loginResponse LoginResponse
	err = json.Unmarshal(data, &loginResponse)

	require.Equal(t, user.Username, loginResponse.Username)

	accessToken := loginResponse.AccessToken
	require.NotEmpty(t, accessToken)
	require.Equal(t, originalAccessToken, accessToken)

	tokenMaker, err := token.NewPasetoMaker(tokenSymmetricKey)
	require.NoError(t, err)

	accessTokenPayload, err := tokenMaker.VerifyToken(accessToken)
	require.NoError(t, err)
	require.NotEmpty(t, accessTokenPayload)

	refreshToken := loginResponse.RefreshToken
	require.NotEmpty(t, refreshToken)
	require.Equal(t, originalRefreshToken, refreshToken)

	refreshTokenPayload, err := tokenMaker.VerifyToken(refreshToken)
	require.NoError(t, err)
	require.NotEmpty(t, refreshTokenPayload)
}
