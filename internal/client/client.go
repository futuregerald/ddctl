// internal/client/client.go
package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	datadog "github.com/DataDog/datadog-api-client-go/v2/api/datadog"
	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
	"github.com/futuregerald/ddctl/internal/keyring"
)

type Client struct {
	apiClient *datadog.APIClient
	ctx       context.Context
	mu        sync.Mutex
	lastCall  time.Time
	rateLimit time.Duration
}

func New(site string, creds *keyring.Credentials) (*Client, error) {
	ctx := datadog.NewDefaultContext(context.Background())
	ctx = context.WithValue(ctx, datadog.ContextAPIKeys, map[string]datadog.APIKey{
		"apiKeyAuth": {Key: creds.APIKey},
		"appKeyAuth": {Key: creds.AppKey},
	})
	ctx = context.WithValue(ctx, datadog.ContextServerVariables, map[string]string{
		"site": site,
	})

	config := datadog.NewConfiguration()
	apiClient := datadog.NewAPIClient(config)

	return &Client{
		apiClient: apiClient,
		ctx:       ctx,
		rateLimit: 2 * time.Second, // 30 req/min default
	}, nil
}

func (c *Client) throttle() {
	c.mu.Lock()
	defer c.mu.Unlock()

	since := time.Since(c.lastCall)
	if since < c.rateLimit {
		time.Sleep(c.rateLimit - since)
	}
	c.lastCall = time.Now()
}

func (c *Client) APIClient() *datadog.APIClient {
	return c.apiClient
}

func (c *Client) Context() context.Context {
	return c.ctx
}

func (c *Client) Validate() error {
	c.throttle()
	api := datadogV1.NewAuthenticationApi(c.apiClient)
	_, _, err := api.Validate(c.ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}
	return nil
}
