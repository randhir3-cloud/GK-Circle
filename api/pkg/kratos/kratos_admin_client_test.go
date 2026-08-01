package kratos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKratosAdminClient_FindIdentitiesByEmail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/admin/identities", req.URL.Path)
		assert.Equal(t, "randhirsandhu81@gmail.com", req.URL.Query().Get("credentials_identifier"))

		identities := []KratosIdentity{
			{
				ID:       "id-123",
				SchemaID: "default",
				State:    "active",
			},
		}
		identities[0].Traits.Email = "randhirsandhu81@gmail.com"
		identities[0].Traits.Name.First = "Randhir"
		identities[0].Traits.Name.Last = "Sandhu"

		data, _ := json.Marshal(identities)
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write(data)
	}))
	defer server.Close()

	client := NewKratosAdminClient(server.URL)
	results, err := client.FindIdentitiesByEmail(context.Background(), "randhirsandhu81@gmail.com")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "id-123", results[0].ID)
	assert.Equal(t, "randhirsandhu81@gmail.com", results[0].Traits.Email)
}

func TestKratosAdminClient_GetIdentity(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		assert.Equal(t, "/admin/identities/id-123", req.URL.Path)

		identity := KratosIdentity{
			ID:       "id-123",
			SchemaID: "default",
			State:    "active",
		}
		identity.Traits.Email = "randhirsandhu81@gmail.com"

		data, _ := json.Marshal(identity)
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		rw.Write(data)
	}))
	defer server.Close()

	client := NewKratosAdminClient(server.URL)
	result, err := client.GetIdentity(context.Background(), "id-123")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "id-123", result.ID)
	assert.Equal(t, "randhirsandhu81@gmail.com", result.Traits.Email)
}

func TestKratosAdminClient_VerifyEmailAddress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodGet {
			assert.Equal(t, "/admin/identities/id-123", req.URL.Path)
			identity := KratosIdentity{
				ID:       "id-123",
				SchemaID: "default",
				State:    "active",
			}
			identity.Traits.Email = "randhirsandhu81@gmail.com"
			data, _ := json.Marshal(identity)
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			rw.Write(data)
			return
		}

		if req.Method == http.MethodPut {
			assert.Equal(t, "/admin/identities/id-123", req.URL.Path)

			var payload struct {
				SchemaID            string                   `json:"schema_id"`
				State               string                   `json:"state"`
				Traits              KratosTraits             `json:"traits"`
				VerifiableAddresses []KratosVerifiableAddress `json:"verifiable_addresses"`
			}
			err := json.NewDecoder(req.Body).Decode(&payload)
			assert.NoError(t, err)
			assert.Equal(t, "default", payload.SchemaID)
			assert.Equal(t, "active", payload.State)
			assert.Equal(t, "randhirsandhu81@gmail.com", payload.Traits.Email)
			assert.Len(t, payload.VerifiableAddresses, 1)
			assert.True(t, payload.VerifiableAddresses[0].Verified)
			assert.Equal(t, "completed", payload.VerifiableAddresses[0].Status)

			response := KratosIdentity{
				ID:                  "id-123",
				SchemaID:            payload.SchemaID,
				State:               payload.State,
				Traits:              payload.Traits,
				VerifiableAddresses: payload.VerifiableAddresses,
			}
			data, _ := json.Marshal(response)
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusOK)
			rw.Write(data)
			return
		}

		t.Fatalf("unexpected request method %s", req.Method)
	}))
	defer server.Close()

	client := NewKratosAdminClient(server.URL)
	result, err := client.VerifyEmailAddress(context.Background(), "id-123", "randhirsandhu81@gmail.com")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.VerifiableAddresses, 1)
	assert.True(t, result.VerifiableAddresses[0].Verified)
	assert.Equal(t, "completed", result.VerifiableAddresses[0].Status)
}
