package commands

import (
	"fmt"

	utils "github.com/zokypesch/proto-lib/utils"
)

// CommandInterfacing for interfacing command
type CommandInterfacing interface {
	Execute(args map[string]string) error
}

// ListOfCommands for list of commands
var ListOfCommands = map[string]CommandInterfacing{
	"gen-proto-db": NewProtoFromDB(),
	"install":      NewInstall(),
}

// Routing for routing the commands
func Routing(command string, args []string) error {
	v, ok := ListOfCommands[command]

	if !ok {
		return fmt.Errorf("Command not found")
	}

	params := utils.SplitSliceParamsToMap(args, "=")
	return v.Execute(params)
}
