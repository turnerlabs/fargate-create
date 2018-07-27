package main

import (
	"github.com/turnerlabs/fargate-create/cmd"
)

var version string

func main() {
	cmd.Execute(version)
}
