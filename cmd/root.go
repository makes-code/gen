package cmd

import (
	"log"
	"os"

	"github.com/makes-code/gen/pkg/command"

	"github.com/mitchellh/cli"
)

const (
	typeModel    = "type model"
	typeDocument = "type document"
	typePayload  = "type payload"
)

func Run() {
	c := cli.NewCLI("makes-code", "0.0.0")
	c.Args = os.Args[1:]
	c.Commands = map[string]cli.CommandFactory{
		typeModel:    command.TypeModel,
		typeDocument: command.TypeDocument,
		typePayload:  command.TypePayload,
	}
	c.HelpWriter = os.Stdout
	c.ErrorWriter = os.Stderr

	exitCode, err := c.Run()
	if err != nil {
		log.Println(err)
	}
	os.Exit(exitCode)
}
