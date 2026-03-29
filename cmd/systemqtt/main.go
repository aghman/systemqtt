package main

import (
	"os"

	"github.com/systemqtt/systemqtt/internal/cli"
)

func main() {
	os.Exit(cli.Run())
}
