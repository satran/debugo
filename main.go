package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"syscall"
)

var deb *log.Logger

func init() {
	deb = log.New(os.Stdout, "", log.Lshortfile)

	// All ptrace calls must be made from the same thread that created the process
	runtime.LockOSThread()
}

func main() {
	var err error
	if len(os.Args) < 2 {
		usage()
		return
	}
	p, err := NewProcess(os.Args[1])
	if err != nil {
		deb.Fatal(err)
	}
	// Wait for the intial call
	var w syscall.WaitStatus
	pid, err := syscall.Wait4(p.Pid, &w, syscall.WALL, nil)

	if err != nil {
		deb.Fatal(err)
	}
	if w.Exited() {
		deb.Fatal("Process exited")
	}

	for _, fn := range os.Args[2:] {
		addr, err := p.FuncAddr(fn)
		if err != nil {
			deb.Println(err)
			continue
		}
		err = p.SetBreakpoint(pid, addr)
		if err != nil {
			deb.Println(err)
			continue
		}
		deb.Println("Set breakpoint", fn)
	}

	// Continue the initial pause.
	err = syscall.PtraceCont(pid, 0)
	if err != nil {
		deb.Fatal(err)
	}

	for {
		pid, err = syscall.Wait4(pid, &w, syscall.WALL, nil)
		if err != nil {
			deb.Fatal(err)
		}
		if w.Exited() {
			deb.Fatal("Process exited")
		}
		err = p.LogAndContinue(pid)
		if err != nil {
			deb.Fatal(err)
		}
	}
}

func usage() {
	fmt.Println("Usage: %s <child_process>", os.Args[0])
}
