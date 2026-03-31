// internal/client/monitors.go
package client

import (
	"encoding/json"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

func (c *Client) ListMonitors() ([]datadogV1.Monitor, error) {
	c.throttle()
	api := datadogV1.NewMonitorsApi(c.apiClient)
	resp, _, err := api.ListMonitors(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing monitors: %w", err)
	}
	return resp, nil
}

func (c *Client) GetMonitor(id int64) ([]byte, error) {
	c.throttle()
	api := datadogV1.NewMonitorsApi(c.apiClient)
	resp, _, err := api.GetMonitor(c.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting monitor %d: %w", id, err)
	}
	return json.MarshalIndent(resp, "", "  ")
}

func (c *Client) CreateMonitor(body []byte) (int64, error) {
	c.throttle()
	api := datadogV1.NewMonitorsApi(c.apiClient)

	var monitor datadogV1.Monitor
	if err := json.Unmarshal(body, &monitor); err != nil {
		return 0, fmt.Errorf("parsing monitor JSON: %w", err)
	}

	resp, _, err := api.CreateMonitor(c.ctx, monitor)
	if err != nil {
		return 0, fmt.Errorf("creating monitor: %w", err)
	}
	return resp.GetId(), nil
}

func (c *Client) UpdateMonitor(id int64, body []byte) error {
	c.throttle()
	api := datadogV1.NewMonitorsApi(c.apiClient)

	var update datadogV1.MonitorUpdateRequest
	if err := json.Unmarshal(body, &update); err != nil {
		return fmt.Errorf("parsing monitor update JSON: %w", err)
	}

	_, _, err := api.UpdateMonitor(c.ctx, id, update)
	if err != nil {
		return fmt.Errorf("updating monitor %d: %w", id, err)
	}
	return nil
}

func (c *Client) DeleteMonitor(id int64) error {
	c.throttle()
	api := datadogV1.NewMonitorsApi(c.apiClient)
	_, _, err := api.DeleteMonitor(c.ctx, id)
	if err != nil {
		return fmt.Errorf("deleting monitor %d: %w", id, err)
	}
	return nil
}
