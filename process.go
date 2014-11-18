package debugo

import (
	"debug/gosym"
	"errors"
	"syscall"
	"runtime"
	"strings"
	"fmt"

	"github.com/satran/debugo/trace"
)

var (
	UndefinedError  = errors.New("method send undefined response")
	TableEmptyError = errors.New("symbol table was empty")
)

type Process struct {
	Path  string
	Table *gosym.Table
	Pid int                 // the main process id after the fork and exec

	in  chan *trace.Command // the channel used to communicate with the goroutine that runs the debugee.
	err error               // sets any exception that was raised initially during creation of process.
}

// New starts the debugee.
func New(path string, args ...string) (*Process, error) {
	p := Process{Path: path}
	var err error
	p.Table, err = getTable(path)
	if err != nil {
		return nil, err
	}

	resp := make(chan int)
	p.in = make(chan *trace.Command)

	go trace.Listen(p.in, resp, path, args...)

	pid := <-resp
	if pid == -1 {
		return nil, errors.New("Process did not start.")
	}

	p.Pid = pid

	return &p, nil
}

// Wait is a nice wrapper over the syscall Wait4.
func (p *Process) Wait() (int, syscall.WaitStatus, error) {
	out := make(chan interface{})
	comm := trace.New(
		trace.C_WAIT,
		trace.Args{
			"pid": p.Pid,
		},
		out,
	)
	p.in <- comm

	resp := <-out

	switch resp.(type) {
	case error:
		return 0, 0, resp.(error)
	case trace.Args:
		args := resp.(trace.Args)
		pid := args["pid"].(int)
		status := args["status"].(syscall.WaitStatus)
		return pid, status, nil
	}

	return 0, 0, UndefinedError
}

// FuncAddr returns the virtual address of the function
func (p *Process) FuncAddr(function string) (uint64, error) {
	if p.Table == nil {
		return 0, TableEmptyError
	}

	fn := p.Table.LookupFunc(function)
	if fn == nil {
		return 0, nil
	}
	return fn.Entry, nil
}

// LineAddr returns the virtual address of a line number in a file
func (p *Process) LineAddr(file string, line int) (uint64, error) {
	if p.Table == nil {
		return 0, TableEmptyError
	}

	pc, _, err := p.Table.LineToPC(file, line)
	if err != nil {
		return 0, err
	}
	return pc, nil
}

// SetBreakpoint sets the break point on the given address.
func (p *Process) SetBreakpoint(pid int, addr uint64) error {
	out := make(chan interface{})
	comm := trace.New(
		trace.C_BREAKPOINT,
		trace.Args{
			"pid":  pid,
			"addr": addr,
		},
		out,
	)
	p.in <- comm

	resp := <-out

	switch resp.(type) {
	case error:
		return resp.(error)
	}

	return nil
}

// Continue resumes execution after a break point has been reached.
func (p *Process) Continue(pid int) error {
	out := make(chan interface{})
	comm := trace.New(
		trace.C_CONTINUE,
		trace.Args{
			"pid": pid,
		},
		out,
	)
	p.in <- comm

	resp := <-out

	switch resp.(type) {
	case error:
		return resp.(error)
	}

	return nil
}

// PC returns the current address of execution.
func (p *Process) PC(pid int) (uint64, error) {
	out := make(chan interface{})
	comm := trace.New(
		trace.C_PC,
		trace.Args{
			"pid": pid,
		},
		out,
	)
	p.in <- comm

	resp := <-out

	switch resp.(type) {
	case error:
		return 0, resp.(error)
	case uint64:
		return resp.(uint64), nil
	}

	return 0, UndefinedError
}

// Files returns the source code of files, if any, that the process was created with.
func (p *Process) Files() []string {
	root := runtime.GOROOT()
	files := make([]string, 0, len(p.Table.Files))
	for key := range p.Table.Files {
		if strings.HasPrefix(key, root) {
			continue
		}
		files = append(files, key)			
	}
	fmt.Println(files)
	return files
}

// CurrentPosition returns the file and line number the process is currently paused on.
func (p *Process) CurrentPosition(pid int) (string, int, error){
	pc, err := p.PC(pid)
	if err != nil {
		return "", 0, nil
	}
	file, line, _ := p.Table.PCToLine(pc)
	return file, line, nil
}