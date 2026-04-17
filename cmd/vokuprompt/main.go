package main

import (
	"fmt"
	"os"

	"github.com/voku/vokuprompt/internal/cli"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: vokuprompt <categories|optimize>")
		os.Exit(1)
	}

	var err error
	switch os.Args[1] {
	case "categories":
		err = cli.ExecuteCategories(os.Args[2:], os.Stdout)
	case "optimize":
		err = cli.ExecuteOptimize(os.Args[2:], os.Stdout)
	default:
		err = fmt.Errorf("unknown command: %s", os.Args[1])
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
