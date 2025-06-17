package http_client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	config := Config{
		BaseURL: "https://api.example.com",
		Timeout: 10 * time.Second,
		DefaultHeaders: map[string]string{
			"Authorization": "Bearer token",
		},
	}

	client := New(config)

	assert.NotNil(t, client)
	assert.Equal(t, "https://api.example.com", client.baseURL)
	assert.Equal(t, map[string]string{"Authorization": "Bearer token"}, client.headers)
	assert.Equal(t, 10*time.Second, client.httpClient.Timeout)
}

func TestNew_DefaultTimeout(t *testing.T) {
	config := Config{
		BaseURL: "https://api.example.com",
	}

	client := New(config)

	assert.Equal(t, 30*time.Second, client.httpClient.Timeout)
}

func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/users", r.URL.Path)
		assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "success"})
	}))
	defer server.Close()

	client := New(Config{BaseURL: server.URL})
	headers := map[string]string{"Authorization": "Bearer token"}

	resp, err := client.Get(context.Background(), "/users", headers)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.True(t, resp.IsSuccess())

	var result map[string]string
	err = resp.JSON(&result)
	require.NoError(t, err)
	assert.Equal(t, "success", result["message"])
}

func TestClient_Post(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "/users", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "John", body["name"])

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": "123"})
	}))
	defer server.Close()

	client := New(Config{BaseURL: server.URL})
	requestBody := map[string]string{"name": "John"}

	resp, err := client.Post(context.Background(), "/users", requestBody, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.True(t, resp.IsSuccess())
}

func TestClient_Put(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		assert.Equal(t, "/users/123", r.URL.Path)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(Config{BaseURL: server.URL})
	requestBody := map[string]string{"name": "John Updated"}

	resp, err := client.Put(context.Background(), "/users/123", requestBody, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "/users/123", r.URL.Path)

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := New(Config{BaseURL: server.URL})

	resp, err := client.Delete(context.Background(), "/users/123", nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestClient_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPatch, r.Method)
		assert.Equal(t, "/users/123", r.URL.Path)

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := New(Config{BaseURL: server.URL})
	requestBody := map[string]string{"email": "john@example.com"}

	resp, err := client.Patch(context.Background(), "/users/123", requestBody, nil)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestResponse_String(t *testing.T) {
	resp := &Response{
		Body: []byte("test response"),
	}

	assert.Equal(t, "test response", resp.String())
}

func TestResponse_IsSuccess(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{200, true},
		{201, true},
		{299, true},
		{300, false},
		{400, false},
		{500, false},
	}

	for _, tt := range tests {
		resp := &Response{StatusCode: tt.statusCode}
		assert.Equal(t, tt.expected, resp.IsSuccess())
	}
}

func TestClient_BuildURL(t *testing.T) {
	client := New(Config{BaseURL: "https://api.example.com"})
	
	url := client.buildURL("/users")
	assert.Equal(t, "https://api.example.com/users", url)

	clientNoBase := New(Config{})
	url = clientNoBase.buildURL("/users")
	assert.Equal(t, "/users", url)
}