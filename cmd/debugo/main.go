package main

import (
	"fmt"
	"log"
	"os"
	"github.com/satran/debugo"
)

var deb *log.Logger

func init() {
	deb = log.New(os.Stdout, "", log.Lshortfile)
}

func main() {
	var err error
	if len(os.Args) < 2 {
		usage()
		return
	}
	p, err := debugo.New(os.Args[1])
	if err != nil {
		deb.Fatal(err)
	}
	// Wait for the intial call
	pid, w, err := p.Wait()
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
	}

	// Continue the initial pause.
	err = p.Continue(pid)
	if err != nil {
		deb.Fatal(err)
	}

	for {
		pid, w, err := p.Wait()
		if err != nil {
			deb.Fatal(err)
		}
		if w.Exited() {
			deb.Fatal("Process exited")
		}
		err = p.Continue(pid)
		if err != nil {
			deb.Fatal(err)
		}
	}
}

func usage() {
	fmt.Println("Usage: %s <child_process>", os.Args[0])
}
