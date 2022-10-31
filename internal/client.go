package internal

import (
	"fmt"

	"github.com/deepmap/oapi-codegen/pkg/securityprovider"
	"github.com/nkvoll/innosonix-maxx/pkg/api"
)

// NewClient returns a new client for the Innosonix Maxx REST API.
func NewClient(address, token string) (*api.ClientWithResponses, error) {
	apiKeyProvider, err := securityprovider.NewSecurityProviderApiKey("header", "token", token)
	if err != nil {
		return nil, err
	}

	return api.NewClientWithResponses(fmt.Sprintf("http://%s/rest-api", address), api.WithRequestEditorFn(apiKeyProvider.Intercept))
}
