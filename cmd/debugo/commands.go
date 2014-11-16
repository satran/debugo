package main

import (
	"fmt"
	"regexp"
	"io/ioutil"
	
	"github.com/satran/debugo"
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
		"getfile": getfile,
	}
}

func (c *Command) Exec(s *Session) {	
	fn, ok := registered[c.Command]
	if !ok {
		sendError(s, c.Id, "Unregistered command.")
		return
	}
	fn(s, c.Id, c.Args...)
}

func run(s *Session, id string, args ...string) {
	fmt.Println("run", args)
	var err error
	if s.process != nil {
		return
	}
	s.process, err = debugo.New(args[0], args[1:]...)
	if err != nil {
		sendError(s, "", err.Error())
		return
	}
	files := s.process.Files()
	cmd := Command {
		Command: "popFiles",
		Args: files,
	}
	s.output <- cmd	
}

func getfile(s *Session, id string, args ...string) {
	fmt.Println("getfile", args)
	content, err := ioutil.ReadFile(args[0])
	if err != nil {
		fmt.Println(err)
		sendError(s, id, err.Error())
		return
	}
	
	cmd := Command {
		Command: "getfile",
		Id: id,
		Args: []string{string(content)},
	}
	s.output <- cmd	
}
func sendError(s *Session, id string, args ...string) {
	cmd := Command{
		Id: id,
		Command: "error",
		Args: args,
	}
	s.output <- cmd
}