package kratos

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	resty "github.com/go-resty/resty/v2"
)

type KratosIdentity struct {
	ID                  string                   `json:"id"`
	SchemaID            string                   `json:"schema_id"`
	State               string                   `json:"state"`
	Traits              KratosTraits             `json:"traits"`
	VerifiableAddresses []KratosVerifiableAddress `json:"verifiable_addresses"`
}

type KratosTraits struct {
	Name struct {
		First string `json:"first"`
		Last  string `json:"last"`
	} `json:"name"`
	Email string `json:"email"`
}

type KratosVerifiableAddress struct {
	ID        string `json:"id"`
	Value     string `json:"value"`
	Verified  bool   `json:"verified"`
	Via       string `json:"via"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type KratosAdminClient interface {
	FindIdentitiesByEmail(ctx context.Context, email string) ([]KratosIdentity, error)
	GetIdentity(ctx context.Context, identityID string) (*KratosIdentity, error)
	VerifyEmailAddress(ctx context.Context, identityID string, email string) (*KratosIdentity, error)
}

type kratosAdminClient struct {
	adminURL string
	client   *resty.Client
}

func NewKratosAdminClient(adminURL string) KratosAdminClient {
	return &kratosAdminClient{
		adminURL: strings.TrimRight(adminURL, "/"),
		client:   resty.New().SetHeader("Accept", "application/json"),
	}
}

func (c *kratosAdminClient) FindIdentitiesByEmail(ctx context.Context, email string) ([]KratosIdentity, error) {
	var identities []KratosIdentity
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	resp, err := c.client.R().
		SetContext(ctx).
		SetQueryParam("credentials_identifier", normalizedEmail).
		SetResult(&identities).
		Get(c.adminURL + "/admin/identities")
	if err != nil {
		return nil, fmt.Errorf("Kratos API request failed: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("Kratos API returned HTTP %d: %s", resp.StatusCode(), resp.String())
	}

	var matched []KratosIdentity
	for _, ident := range identities {
		if strings.EqualFold(strings.TrimSpace(ident.Traits.Email), normalizedEmail) {
			matched = append(matched, ident)
		}
	}

	return matched, nil
}

func (c *kratosAdminClient) GetIdentity(ctx context.Context, identityID string) (*KratosIdentity, error) {
	var identity KratosIdentity
	resp, err := c.client.R().
		SetContext(ctx).
		SetResult(&identity).
		Get(c.adminURL + "/admin/identities/" + identityID)
	if err != nil {
		return nil, fmt.Errorf("Kratos API request failed: %w", err)
	}

	if resp.StatusCode() == http.StatusNotFound {
		return nil, fmt.Errorf("identity %s not found: HTTP 404", identityID)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("Kratos API returned HTTP %d: %s", resp.StatusCode(), resp.String())
	}

	return &identity, nil
}

func (c *kratosAdminClient) VerifyEmailAddress(ctx context.Context, identityID string, email string) (*KratosIdentity, error) {
	identity, err := c.GetIdentity(ctx, identityID)
	if err != nil {
		return nil, err
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(email))

	type patchOp struct {
		Op    string      `json:"op"`
		Path  string      `json:"path"`
		Value interface{} `json:"value"`
	}

	var patchPayload []patchOp
	found := false
	for idx, addr := range identity.VerifiableAddresses {
		if strings.EqualFold(strings.TrimSpace(addr.Value), normalizedEmail) {
			patchPayload = append(patchPayload, patchOp{
				Op:    "replace",
				Path:  fmt.Sprintf("/verifiable_addresses/%d/verified", idx),
				Value: true,
			}, patchOp{
				Op:    "replace",
				Path:  fmt.Sprintf("/verifiable_addresses/%d/status", idx),
				Value: "completed",
			})
			found = true
		}
	}

	if !found {
		// If the address was not found in verifiable_addresses list, add it
		patchPayload = append(patchPayload, patchOp{
			Op:   "add",
			Path: "/verifiable_addresses",
			Value: []interface{}{
				map[string]interface{}{
					"value":    normalizedEmail,
					"verified": true,
					"via":      "email",
					"status":   "completed",
				},
			},
		})
	}

	var updatedIdentity KratosIdentity
	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json-patch+json").
		SetBody(patchPayload).
		SetResult(&updatedIdentity).
		Patch(c.adminURL + "/admin/identities/" + identityID)
	if err != nil {
		return nil, fmt.Errorf("Kratos PATCH API request failed: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("Kratos PATCH API returned HTTP %d: %s", resp.StatusCode(), resp.String())
	}

	return &updatedIdentity, nil
}
