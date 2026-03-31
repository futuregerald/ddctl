// internal/client/logs.go
package client

import (
	"encoding/json"
	"fmt"

	"github.com/DataDog/datadog-api-client-go/v2/api/datadogV2"
)

func (c *Client) SearchLogs(query string, from, to string, limit int) ([]byte, error) {
	c.throttle()

	api := datadogV2.NewLogsApi(c.apiClient)

	body := datadogV2.LogsListRequest{
		Filter: &datadogV2.LogsQueryFilter{
			Query: &query,
			From:  &from,
			To:    &to,
		},
	}

	if limit > 0 {
		page := datadogV2.LogsListRequestPage{
			Limit: datadog_int32(int32(limit)),
		}
		body.Page = &page
	}

	resp, _, err := api.ListLogs(c.ctx, *datadogV2.NewListLogsOptionalParameters().WithBody(body))
	if err != nil {
		return nil, fmt.Errorf("searching logs: %w", err)
	}
	return json.MarshalIndent(resp, "", "  ")
}

func datadog_int32(i int32) *int32 {
	return &i
}
