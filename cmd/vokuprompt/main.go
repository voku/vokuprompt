package main

import (
	"fmt"
	"io"
	"os"

	"github.com/voku/vokuprompt/internal/cli"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "usage: vokuprompt <categories|optimize>")
		return 1
	}

	var err error
	switch args[0] {
	case "categories":
		err = cli.ExecuteCategories(args[1:], stdout)
	case "optimize":
		err = cli.ExecuteOptimize(args[1:], stdout)
	default:
		err = fmt.Errorf("unknown command: %s", args[0])
	}

	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return 0
}
