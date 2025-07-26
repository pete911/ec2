package main

import (
	"github.com/pete911/ec2/internal/cmd"
	"os"
)

var Version = "dev"

func main() {
	cmd.Version = Version
	if err := cmd.Root.Execute(); err != nil {
		os.Exit(1)
	}
}
