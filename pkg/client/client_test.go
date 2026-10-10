package client_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/netologist/go-service-template/pkg/client"
)

func TestClient_FindItems(t *testing.T) {
	t.Parallel()

	itemID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/items", r.URL.Path)
		assert.Equal(t, "Widget", r.URL.Query().Get("name"))

		items := []client.Item{
			{
				Id:          itemID,
				Name:        "Widget",
				Description: "A great widget",
				Price:       4999,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(items)
	}))
	defer server.Close()

	c, err := client.NewClientWithResponses(server.URL)
	require.NoError(t, err)

	nameQuery := "Widget"
	resp, err := c.FindItemsWithResponse(context.Background(), &client.FindItemsParams{
		Name: &nameQuery,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.NotNil(t, resp.JSON200)
	require.Len(t, *resp.JSON200, 1)
	assert.Equal(t, "Widget", (*resp.JSON200)[0].Name)
	assert.Equal(t, int64(4999), (*resp.JSON200)[0].Price)
}

func TestClient_CreateItem(t *testing.T) {
	t.Parallel()

	itemID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/items", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)

		var req client.CreateItemJSONRequestBody
		if err := json.NewDecoder(r.Body).Decode(&req); !assert.NoError(t, err) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		desc := ""
		if req.Description != nil {
			desc = *req.Description
		}

		created := client.Item{
			Id:          itemID,
			Name:        req.Name,
			Description: desc,
			Price:       req.Price,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(created)
	}))
	defer server.Close()

	c, err := client.NewClientWithResponses(server.URL)
	require.NoError(t, err)

	desc := "Ergonomic keyboard"
	resp, err := c.CreateItemWithResponse(context.Background(), client.CreateItemJSONRequestBody{
		Name:        "Mechanical Keyboard",
		Price:       9900,
		Description: &desc,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode())
	require.NotNil(t, resp.JSON201)
	assert.Equal(t, "Mechanical Keyboard", resp.JSON201.Name)
	assert.Equal(t, int64(9900), resp.JSON201.Price)
}
