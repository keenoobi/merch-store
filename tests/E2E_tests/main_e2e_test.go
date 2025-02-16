package e2e

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestE2E(t *testing.T) {
	server, cleanup := setupTestServer(t)
	t.Cleanup(func() {
		cleanup()
	})

	makeRequest := func(method, url, body string, token string, result interface{}) *http.Response {
		req, err := http.NewRequest(method, url, strings.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		err = json.NewDecoder(resp.Body).Decode(result)
		require.NoError(t, err)

		return resp
	}

	t.Run("Auth_SuccessNewUser", func(t *testing.T) {
		reqBody := `{"username": "newuser", "password": "password123"}`

		var authResponse AuthResponse
		resp := makeRequest(http.MethodPost, server.URL+"/api/auth", reqBody, "", &authResponse)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.NotEmpty(t, authResponse.Token)
	})

	t.Run("Auth_SuccessExistingUser", func(t *testing.T) {
		reqBody := `{"username": "existinguser", "password": "password123"}`
		var authResponse AuthResponse
		resp := makeRequest(http.MethodPost, server.URL+"/api/auth", reqBody, "", &authResponse)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.NotEmpty(t, authResponse.Token)

		resp = makeRequest(http.MethodPost, server.URL+"/api/auth", reqBody, "", &authResponse)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.NotEmpty(t, authResponse.Token)
	})

	t.Run("Auth_MissingUsernameOrPassword", func(t *testing.T) {
		reqBody := `{"password": "password123"}`
		var errorResponse ErrorResponse
		resp := makeRequest(http.MethodPost, server.URL+"/api/auth", reqBody, "", &errorResponse)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		require.Contains(t, errorResponse.Errors, "Username and password are required")

		reqBody = `{"username": "testuser"}`
		resp = makeRequest(http.MethodPost, server.URL+"/api/auth", reqBody, "", &errorResponse)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		require.Contains(t, errorResponse.Errors, "Username and password are required")
	})

	t.Run("Auth_InvalidJSON", func(t *testing.T) {
		reqBody := `{"username": "testuser", "password": 123}`
		var errorResponse ErrorResponse
		resp := makeRequest(http.MethodPost, server.URL+"/api/auth", reqBody, "", &errorResponse)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		require.Contains(t, errorResponse.Errors, "Invalid request")
	})

	t.Run("Auth_UnauthorizedInvalidToken", func(t *testing.T) {
		reqBody := `{"username": "testuser", "password": "password123"}`
		var authResponse AuthResponse
		resp := makeRequest(http.MethodPost, server.URL+"/api/auth", reqBody, "", &authResponse)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.NotEmpty(t, authResponse.Token)

		req, err := http.NewRequest(http.MethodGet, server.URL+"/api/info", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer invalidtoken")

		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		require.Equal(t, "application/json", resp.Header.Get("Content-Type"))

		var errorResponse ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		require.NoError(t, err)

		require.Contains(t, errorResponse.Errors, "Invalid token")
	})

	t.Run("GetUserInfo_Success", func(t *testing.T) {
		reqBody := `{"username": "testuser", "password": "password123"}`
		var authResponse AuthResponse
		resp := makeRequest(http.MethodPost, server.URL+"/api/auth", reqBody, "", &authResponse)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		token := authResponse.Token
		require.NotEmpty(t, token)

		var infoResponse InfoResponse
		resp = makeRequest(http.MethodGet, server.URL+"/api/info", "", token, &infoResponse)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Equal(t, 1000, infoResponse.Coins)
		require.Empty(t, infoResponse.Inventory)
		require.Empty(t, infoResponse.CoinHistory.Received)
		require.Empty(t, infoResponse.CoinHistory.Sent)
	})

	t.Run("GetUserInfo_Unauthorized", func(t *testing.T) {
		var errorResponse ErrorResponse
		resp := makeRequest(http.MethodGet, server.URL+"/api/info", "", "", &errorResponse)

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		require.Equal(t, "Unauthorized", errorResponse.Errors)
	})

	t.Run("GetUserInfo_InvalidToken", func(t *testing.T) {
		var errorResponse ErrorResponse
		resp := makeRequest(http.MethodGet, server.URL+"/api/info", "", "invalidtoken", &errorResponse)

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		require.Equal(t, "Invalid token", errorResponse.Errors)
	})

	t.Run("SendCoin_Success", func(t *testing.T) {
		requset := `{"username": "user1", "password": "password123"}`
		var authResponse1 AuthResponse
		resp := makeRequest(http.MethodPost, server.URL+"/api/auth", requset, "", &authResponse1)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		token1 := authResponse1.Token
		require.NotEmpty(t, token1)

		reqBody2 := `{"username": "user2", "password": "password123"}`
		var authResponse2 AuthResponse
		resp = makeRequest(http.MethodPost, server.URL+"/api/auth", reqBody2, "", &authResponse2)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		token2 := authResponse2.Token
		require.NotEmpty(t, token2)

		sendCoinReq := SendCoinRequest{
			ToUser: "user2",
			Amount: 100,
		}
		reqBody, err := json.Marshal(sendCoinReq)
		require.NoError(t, err)

		var sendCoinResponse SendCoinResponse
		resp = makeRequest(http.MethodPost, server.URL+"/api/sendCoin", string(reqBody), token1, &sendCoinResponse)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Equal(t, "Coins transferred successfully", sendCoinResponse.Message)
	})

	t.Run("SendCoin_InsufficientFunds", func(t *testing.T) {
		requset := `{"username": "user1", "password": "password123"}`
		var authResponse1 AuthResponse
		resp := makeRequest(http.MethodPost, server.URL+"/api/auth", requset, "", &authResponse1)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		token1 := authResponse1.Token
		require.NotEmpty(t, token1)

		reqBody2 := `{"username": "user2", "password": "password123"}`
		var authResponse2 AuthResponse
		resp = makeRequest(http.MethodPost, server.URL+"/api/auth", reqBody2, "", &authResponse2)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		token2 := authResponse2.Token
		require.NotEmpty(t, token2)

		sendCoinReq := SendCoinRequest{
			ToUser: "user2",
			Amount: 1000000000,
		}
		reqBody, err := json.Marshal(sendCoinReq)
		require.NoError(t, err)

		var sendCoinResponse SendCoinResponse
		resp = makeRequest(http.MethodPost, server.URL+"/api/sendCoin", string(reqBody), token1, &sendCoinResponse)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		require.Contains(t, sendCoinResponse.Errors, "insufficient coins")
	})

	t.Run("SendCoin_SelfTransfer", func(t *testing.T) {
		requset := `{"username": "user1", "password": "password123"}`
		var authResponse AuthResponse
		resp := makeRequest(http.MethodPost, server.URL+"/api/auth", requset, "", &authResponse)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		token := authResponse.Token
		require.NotEmpty(t, token)

		sendCoinReq := SendCoinRequest{
			ToUser: "user1",
			Amount: 100,
		}
		reqBody, err := json.Marshal(sendCoinReq)
		require.NoError(t, err)

		var sendCoinResponse SendCoinResponse
		resp = makeRequest(http.MethodPost, server.URL+"/api/sendCoin", string(reqBody), token, &sendCoinResponse)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		require.Contains(t, sendCoinResponse.Errors, "cannot send coins to yourself")
	})

	t.Run("SendCoin_MissingToUserOrAmount", func(t *testing.T) {
		requset := `{"username": "user1", "password": "password123"}`
		var authResponse AuthResponse
		resp := makeRequest(http.MethodPost, server.URL+"/api/auth", requset, "", &authResponse)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		token := authResponse.Token
		require.NotEmpty(t, token)

		sendCoinReq := SendCoinRequest{
			Amount: 100,
		}
		reqBody, err := json.Marshal(sendCoinReq)
		require.NoError(t, err)

		var sendCoinResponse SendCoinResponse
		resp = makeRequest(http.MethodPost, server.URL+"/api/sendCoin", string(reqBody), token, &sendCoinResponse)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		require.Contains(t, sendCoinResponse.Errors, "toUser is required")

		sendCoinReq = SendCoinRequest{
			ToUser: "user2",
		}
		reqBody, err = json.Marshal(sendCoinReq)
		require.NoError(t, err)

		resp = makeRequest(http.MethodPost, server.URL+"/api/sendCoin", string(reqBody), token, &sendCoinResponse)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		require.Contains(t, sendCoinResponse.Errors, "Amount is required")
	})

	t.Run("SendCoin_InvalidJSON", func(t *testing.T) {
		reqBody := `{"username": "user1", "password": "password123"}`
		var authResponse AuthResponse
		resp := makeRequest(http.MethodPost, server.URL+"/api/auth", reqBody, "", &authResponse)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		token := authResponse.Token
		require.NotEmpty(t, token)

		invalidReqBody := `{"toUser": "user2", "amount": "100"}`
		var sendCoinResponse SendCoinResponse
		resp = makeRequest(http.MethodPost, server.URL+"/api/sendCoin", invalidReqBody, token, &sendCoinResponse)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		require.Contains(t, sendCoinResponse.Errors, "invalid request body")
	})

	t.Run("SendCoin_NonExistentUser", func(t *testing.T) {
		requset := `{"username": "user1", "password": "password123"}`
		var authResponse AuthResponse
		resp := makeRequest(http.MethodPost, server.URL+"/api/auth", requset, "", &authResponse)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		token := authResponse.Token
		require.NotEmpty(t, token)

		sendCoinReq := SendCoinRequest{
			ToUser: "nonexistentuser",
			Amount: 100,
		}
		reqBody, err := json.Marshal(sendCoinReq)
		require.NoError(t, err)

		var sendCoinResponse SendCoinResponse
		resp = makeRequest(http.MethodPost, server.URL+"/api/sendCoin", string(reqBody), token, &sendCoinResponse)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		require.Contains(t, sendCoinResponse.Errors, "recipient does not exist")
	})

	t.Run("BuyItem_Success", func(t *testing.T) {
		reqBody := `{"username": "user1", "password": "password123"}`
		var authResponse AuthResponse
		resp := makeRequest(http.MethodPost, server.URL+"/api/auth", reqBody, "", &authResponse)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		token := authResponse.Token
		require.NotEmpty(t, token)

		itemName := "t-shirt"
		var buyItemResponse BuyItemResponse
		resp = makeRequest(http.MethodGet, server.URL+"/api/buy/"+itemName, "", token, &buyItemResponse)

		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Equal(t, "Item purchased successfully", buyItemResponse.Message)
	})

	t.Run("BuyItem_InsufficientFunds", func(t *testing.T) {
		reqBody := `{"username": "user1", "password": "password123"}`
		var authResponse AuthResponse
		resp := makeRequest(http.MethodPost, server.URL+"/api/auth", reqBody, "", &authResponse)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		token := authResponse.Token
		require.NotEmpty(t, token)

		itemName := "pink-hoody"
		var buyItemResponse BuyItemResponse
		resp = makeRequest(http.MethodGet, server.URL+"/api/buy/"+itemName, "", token, &buyItemResponse)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		resp = makeRequest(http.MethodGet, server.URL+"/api/buy/"+itemName, "", token, &buyItemResponse)
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		require.Contains(t, buyItemResponse.Errors, "insufficient coins")
	})

	t.Run("BuyItem_NonExistentItem", func(t *testing.T) {
		reqBody := `{"username": "user1", "password": "password123"}`
		var authResponse AuthResponse
		resp := makeRequest(http.MethodPost, server.URL+"/api/auth", reqBody, "", &authResponse)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		token := authResponse.Token
		require.NotEmpty(t, token)

		itemName := "nonexistent-item"
		var buyItemResponse BuyItemResponse
		resp = makeRequest(http.MethodGet, server.URL+"/api/buy/"+itemName, "", token, &buyItemResponse)

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)

		require.Contains(t, buyItemResponse.Errors, "item not found")
	})

	t.Run("BuyItem_Unauthorized", func(t *testing.T) {
		itemName := "t-shirt"
		var buyItemResponse BuyItemResponse
		resp := makeRequest(http.MethodGet, server.URL+"/api/buy/"+itemName, "", "", &buyItemResponse)

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		require.Equal(t, "Unauthorized", buyItemResponse.Errors)
	})

	t.Run("BuyItem_InvalidToken", func(t *testing.T) {
		reqBody := `{"username": "user1", "password": "password123"}`
		var authResponse AuthResponse
		resp := makeRequest(http.MethodPost, server.URL+"/api/auth", reqBody, "", &authResponse)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		token := authResponse.Token
		require.NotEmpty(t, token)

		itemName := "t-shirt"
		var buyItemResponse BuyItemResponse
		resp = makeRequest(http.MethodGet, server.URL+"/api/buy/"+itemName, "", "invalidtoken", &buyItemResponse)

		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		require.Equal(t, "Invalid token", buyItemResponse.Errors)
	})

}
