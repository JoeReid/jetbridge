package main

import (
	"fmt"
	"os"

	"github.com/JoeReid/jetbridge/cmd/cli/commands"
)

func main() {
	if err := commands.App.Run(os.Args); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
