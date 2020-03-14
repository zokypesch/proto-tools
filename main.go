package main

import (
	"flag"
	"log"

	commands "github.com/zokypesch/proto-tools/commands"
)

func main() {
	cmd := flag.String("cmd", "no action", "request command")

	flag.Parse()

	err := commands.Routing(*cmd, flag.Args())
	if err != nil {
		log.Println(err, *cmd, flag.Args())
	}
	log.Println("complete")
}
