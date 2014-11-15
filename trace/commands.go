package trace

import (
	"errors"
	"syscall"
)

// Declarations of all commands that the exec accepts
const (
	C_WAIT = iota
	C_BREAKPOINT
	C_CONTINUE
	C_PC
)

var (
	NotBreakPointError = errors.New("cannot perform continue when break point was not set")
	ProcessExitedError = errors.New("process exited before we could do anything")
)

var (
	breakpoints map[uint64]byte // stores the original data of address when over writing with int3
)

type Args map[string]interface{}

type Command struct {
	method int
	args   Args
	out    chan interface{}
}

func init() {
	breakpoints = make(map[uint64]byte)
}

func New(method int, args Args, out chan interface{}) *Command {
	return &Command{
		method: method,
		args:   args,
		out:    out,
	}
}

func (c *Command) wait() {
	pid := c.args["pid"].(int)

	pid, w, err := wait(pid)
	if err != nil {
		c.out <- err
	} else {
		args := Args{
			"pid":    pid,
			"status": w,
		}
		c.out <- args
	}
}

func (c *Command) setbreakpoint() {
	pid := c.args["pid"].(int)
	addr := c.args["addr"].(uint64)

	c.out <- setbreakpoint(pid, addr)
}

func (c *Command) kontinue() {
	pid := c.args["pid"].(int)
	pc, err := pc(pid)
	if err != nil {
		c.out <- err
		return
	}

	_, ok := breakpoints[pc-1]
	if !ok {
		// The break was not from one of our registered breakpoints. Lets just continue
		err = syscall.PtraceCont(pid, 0)
		c.out <- err
		return

	}

	err = continueBreakpoint(pid, pc-1)
	if err != nil {
		c.out <- err
		return
	}

	c.out <- nil
}

func (c *Command) pc() {
	pid := c.args["pid"].(int)
	pc, err := pc(pid)
	if err != nil {
		c.out <- err
		return
	}
	c.out <- pc
}
