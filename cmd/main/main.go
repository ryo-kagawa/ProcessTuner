package main

import (
	"fmt"
	"os"

	"github.com/ryo-kagawa/ProcessTuner/cmd/main/command"
	"github.com/ryo-kagawa/ProcessTuner/subcommand"
	"github.com/ryo-kagawa/go-utils/commandline"
)

func main() {
	result, err := commandline.Execute(
		command.Command{},
		os.Args[1:],
		subcommand.Version{},
	)
	if result != "" {
		fmt.Fprintln(os.Stdout, result)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
