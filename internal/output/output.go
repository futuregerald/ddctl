// internal/output/output.go
package output

import (
	"encoding/json"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

func JSON(w io.Writer, data any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func YAML(w io.Writer, data any) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	defer enc.Close()
	return enc.Encode(data)
}

func Table(w io.Writer, headers []string, rows [][]string) {
	if len(rows) == 0 {
		fmt.Fprintln(w, "  No results.")
		return
	}

	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	for i, h := range headers {
		fmt.Fprintf(w, "  %-*s", widths[i]+2, h)
	}
	fmt.Fprintln(w)

	// Print separator
	for i := range headers {
		fmt.Fprintf(w, "  ")
		for j := 0; j < widths[i]; j++ {
			fmt.Fprintf(w, "─")
		}
	}
	fmt.Fprintln(w)

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				fmt.Fprintf(w, "  %-*s", widths[i]+2, cell)
			}
		}
		fmt.Fprintln(w)
	}
}

type ErrorResponse struct {
	Error struct {
		Code    int    `json:"code" yaml:"code"`
		Message string `json:"message" yaml:"message"`
	} `json:"error" yaml:"error"`
}

func Error(w io.Writer, format string, code int, message string) {
	switch format {
	case "json":
		resp := ErrorResponse{}
		resp.Error.Code = code
		resp.Error.Message = message
		JSON(w, resp)
	case "yaml":
		resp := ErrorResponse{}
		resp.Error.Code = code
		resp.Error.Message = message
		YAML(w, resp)
	default:
		fmt.Fprintf(w, "Error: %s\n", message)
	}
}
