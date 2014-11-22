package main

import (
	"fmt"
	"regexp"
	"io/ioutil"
	"strings"
	"strconv"
	
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
		"b": breakpoint,
		"c": kontinue,
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

func breakpoint(s *Session, id string, args ...string) {
	for _, arg := range args {
		// We make an assumption that the argument will be file_name:line_number
		params := strings.Split(arg, ":")
		if len(params) != 2 {
			continue
		}
		file := params[0]
		line, err := strconv.Atoi(params[1])
		if err != nil {
			sendError(s, id, err.Error())
			continue
		}
		
		pc, err := s.process.LineAddr(file, line)
		if err != nil {
			fmt.Println(err)
			sendError(s, id, err.Error())
			continue
		}
		
		err = s.process.SetBreakpoint(s.process.Pid, pc)
		if err != nil {
			fmt.Println(err)
			sendError(s, id, err.Error())
			continue
		}
		cmd := Command {
			Id: id,
			Command: "b",
			Args: []string{
				file,
				fmt.Sprintf("%d", line),
			},
		}
		s.output <- cmd
	}
}

func kontinue(s *Session, id string, args ...string) {
	err := s.process.Continue(s.process.Pid)
	if err != nil {
		fmt.Println(err)
		sendError(s, id, err.Error())
		return
	}
	
	// Lets wait for signals from the process.
	pid, w, err := s.process.Wait()
	if err != nil {
		fmt.Println(err)
		sendError(s, id, err.Error())
		return
	}
	
	if w.Exited() {
		sendError(s, id, "Process exited.")
		return
	}
	
	file, line, err := s.process.CurrentPosition(pid)
	if err != nil {
		fmt.Println(err)
		sendError(s, id, err.Error())
		return
	}
	
	cmd := Command {
		Command: "paused",
		Args: []string{
			file,
			fmt.Sprintf("%d", line),
		},
	}
	s.output <- cmd
}