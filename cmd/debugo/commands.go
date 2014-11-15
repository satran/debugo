package main

import (
	"regexp"	
)

type Command struct {
	Id      string // A unique string to track commands commands from server usually start with an s
	Command string
	Args    []string
	Server  bool

	parsed  []string // Stores parsed command and arguments.
	session *Session // the active session that executed the command
}

// Regex for splitting command and arguments.
var cmdRegex = regexp.MustCompile("'.+'|\".+\"|\\S+")

func (c *Command) Exec(s *Session) {

}
