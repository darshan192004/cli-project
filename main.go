package main

import (
	"os"

	"dataset-cli/cmd"
)

func main() {
	if len(os.Args) == 1 {
		cmd.StartInteractive()
		return
	}
	cmd.Execute()
}
