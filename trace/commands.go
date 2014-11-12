package trace

import (
	"syscall"
	"errors"
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
		args: args,
		out: out,
	}
}

func (c *Command) wait() {
	pid := c.args["pid"].(int)
	var w syscall.WaitStatus
	pid, err := syscall.Wait4(pid, &w, syscall.WALL, nil)
	if err != nil {
		c.out <- err
	} else {
		args := Args{
			"pid": pid,
			"status":  w,
		}
		c.out <- args
	}
}

func (c *Command) setbreakpoint() {
	pid := c.args["pid"].(int)
	addr := c.args["addr"].(uint64)

	var text = []byte{0}

	// Read the original data
	_, err := syscall.PtracePeekText(pid, uintptr(addr), text)
	if err != nil {
		c.out <- err
		return
	}

	// Store the original data in the cache, useful when we want to
	// clear the breakpoint
	breakpoints[addr] = text[0]

	// Write the breakpoint, very specific to x86 the int 3 instruction.
	text = []byte{0xCC}

	_, err = syscall.PtracePokeText(pid, uintptr(addr), text)
	if err != nil {
		c.out <- err
		return
	}
	
	c.out <- nil
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