package trace

import (
	"os"
	"runtime"
	"syscall"
)

// Listen starts the process and waits for messages to be send to the in chan. Ideally this
// should be started as a separate goroutine.
func Listen(in chan *Command, resp chan int, path string, args ...string) {
	// ptrace commands must be made fro mthe same thread that created the process.
	runtime.LockOSThread()

	attr := &syscall.ProcAttr{
		Sys:   &syscall.SysProcAttr{Ptrace: true},
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
	}

	newargs := make([]string, 0, len(args)+1)
	newargs = append(newargs, path)
	for _, arg := range args {
		newargs = append(newargs, arg)
	}

	pid, err := syscall.ForkExec(path, newargs, attr)
	if err != nil {
		resp <- -1
		return
	}

	resp <- pid

	for comm := range in {
		exec(comm)
	}

	runtime.UnlockOSThread()
}

// exec executes the appropriate command and pushes the response data into the out chan.
func exec(c *Command) {
	switch c.method {
	case C_WAIT:
		c.wait()
	case C_BREAKPOINT:
		c.setbreakpoint()
	case C_CONTINUE:
		c.kontinue()
	case C_PC:
		c.pc()
	}
}
