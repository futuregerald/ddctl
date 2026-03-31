// internal/client/dashboards.go
package client

import (
	"encoding/json"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

func (c *Client) ListDashboards() ([]datadogV1.DashboardSummaryDefinition, error) {
	c.throttle()
	api := datadogV1.NewDashboardsApi(c.apiClient)
	resp, _, err := api.ListDashboards(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing dashboards: %w", err)
	}
	return resp.GetDashboards(), nil
}

func (c *Client) GetDashboard(id string) ([]byte, error) {
	c.throttle()
	api := datadogV1.NewDashboardsApi(c.apiClient)
	resp, _, err := api.GetDashboard(c.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting dashboard %s: %w", id, err)
	}
	return json.MarshalIndent(resp, "", "  ")
}

func (c *Client) CreateDashboard(body []byte) (string, error) {
	c.throttle()
	api := datadogV1.NewDashboardsApi(c.apiClient)

	var dashboard datadogV1.Dashboard
	if err := json.Unmarshal(body, &dashboard); err != nil {
		return "", fmt.Errorf("parsing dashboard JSON: %w", err)
	}

	resp, _, err := api.CreateDashboard(c.ctx, dashboard)
	if err != nil {
		return "", fmt.Errorf("creating dashboard: %w", err)
	}
	return resp.GetId(), nil
}

func (c *Client) UpdateDashboard(id string, body []byte) error {
	c.throttle()
	api := datadogV1.NewDashboardsApi(c.apiClient)

	var dashboard datadogV1.Dashboard
	if err := json.Unmarshal(body, &dashboard); err != nil {
		return fmt.Errorf("parsing dashboard JSON: %w", err)
	}

	_, _, err := api.UpdateDashboard(c.ctx, id, dashboard)
	if err != nil {
		return fmt.Errorf("updating dashboard %s: %w", id, err)
	}
	return nil
}

func (c *Client) DeleteDashboard(id string) error {
	c.throttle()
	api := datadogV1.NewDashboardsApi(c.apiClient)
	_, _, err := api.DeleteDashboard(c.ctx, id)
	if err != nil {
		return fmt.Errorf("deleting dashboard %s: %w", id, err)
	}
	return nil
}
