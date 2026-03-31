// internal/client/slos.go
package client

import (
	"encoding/json"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

func (c *Client) ListSLOs() ([]datadogV1.ServiceLevelObjective, error) {
	c.throttle()
	api := datadogV1.NewServiceLevelObjectivesApi(c.apiClient)
	resp, _, err := api.ListSLOs(c.ctx)
	if err != nil {
		return nil, fmt.Errorf("listing SLOs: %w", err)
	}
	return resp.GetData(), nil
}

func (c *Client) GetSLO(id string) ([]byte, error) {
	c.throttle()
	api := datadogV1.NewServiceLevelObjectivesApi(c.apiClient)
	resp, _, err := api.GetSLO(c.ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting SLO %s: %w", id, err)
	}
	return json.MarshalIndent(resp.GetData(), "", "  ")
}

func (c *Client) CreateSLO(body []byte) (string, error) {
	c.throttle()
	api := datadogV1.NewServiceLevelObjectivesApi(c.apiClient)

	var slo datadogV1.ServiceLevelObjectiveRequest
	if err := json.Unmarshal(body, &slo); err != nil {
		return "", fmt.Errorf("parsing SLO JSON: %w", err)
	}

	resp, _, err := api.CreateSLO(c.ctx, slo)
	if err != nil {
		return "", fmt.Errorf("creating SLO: %w", err)
	}
	data := resp.GetData()
	if len(data) == 0 {
		return "", fmt.Errorf("no SLO returned from create")
	}
	return data[0].GetId(), nil
}

func (c *Client) UpdateSLO(id string, body []byte) error {
	c.throttle()
	api := datadogV1.NewServiceLevelObjectivesApi(c.apiClient)

	var slo datadogV1.ServiceLevelObjective
	if err := json.Unmarshal(body, &slo); err != nil {
		return fmt.Errorf("parsing SLO JSON: %w", err)
	}

	_, _, err := api.UpdateSLO(c.ctx, id, slo)
	if err != nil {
		return fmt.Errorf("updating SLO %s: %w", id, err)
	}
	return nil
}

func (c *Client) DeleteSLO(id string) error {
	c.throttle()
	api := datadogV1.NewServiceLevelObjectivesApi(c.apiClient)
	_, _, err := api.DeleteSLO(c.ctx, id)
	if err != nil {
		return fmt.Errorf("deleting SLO %s: %w", id, err)
	}
	return nil
}
