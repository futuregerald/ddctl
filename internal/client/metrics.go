// internal/client/metrics.go
package client

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV1"
)

type MetricSearchResult struct {
	Metrics []string `json:"metrics"`
}

func (c *Client) SearchMetrics(query string) (*MetricSearchResult, error) {
	c.throttle()
	api := datadogV1.NewMetricsApi(c.apiClient)
	resp, _, err := api.ListMetrics(c.ctx, query)
	if err != nil {
		return nil, fmt.Errorf("searching metrics: %w", err)
	}
	results := resp.GetResults()
	return &MetricSearchResult{Metrics: results.GetMetrics()}, nil
}

type MetricQueryResult struct {
	Series []json.RawMessage `json:"series"`
}

func (c *Client) QueryMetrics(query string, from, to time.Time) ([]byte, error) {
	c.throttle()
	api := datadogV1.NewMetricsApi(c.apiClient)
	resp, _, err := api.QueryMetrics(c.ctx, from.Unix(), to.Unix(), query)
	if err != nil {
		return nil, fmt.Errorf("querying metrics: %w", err)
	}
	return json.MarshalIndent(resp, "", "  ")
}
