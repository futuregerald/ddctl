// main.go
package main

import (
	"os"

	"github.com/futuregerald/ddctl/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
