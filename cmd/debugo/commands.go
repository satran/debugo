package main

import (
	"fmt"
	"regexp"
	"strings"	
)

type Command struct {
	Id      string // A unique string to track commands commands from server usually start with an s
	Command string
	Args    []string
	Server  bool

	parsed  []string // Stores parsed command and arguments.
	session *Session // the active session that executed the command
}

var registered map[string]func(*Session, string, ...string)

// Regex for splitting command and arguments.
var cmdRegex = regexp.MustCompile("'.+'|\".+\"|\\S+")


func init() {
	registered = map[string]func(*Session, string, ...string) {
		"run": run,
	}
}

func (c *Command) Exec(s *Session) {
	args := strings.Split(c.Command, " ")
	if len(args) < 1 {
		return
	}
	
	command := args[0]
	args = args[1:]
	
	fn, ok := registered[command]
	if !ok {
		error(s, c.Id, "Unregistered command.")
		return
	}
	fn(s, c.Id, args...)
}

func run(s *Session, id string, args ...string) {
	fmt.Println(id, args)
}

func error(s *Session, id string, args ...string) {
	cmd := Command{
		Command: "error",
		Args: args,
	}
	s.output <- cmd
}