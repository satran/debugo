package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	
	"github.com/satran/debugo"	
)

var deb *log.Logger
var lines map[string]int

func init() {
	deb = log.New(os.Stdout, "", log.Lshortfile)
	lines = make(map[string]int)
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

	addr, err := p.FuncAddr("main.main")
	if err != nil {
		deb.Fatal(err)
	}
	err = p.SetBreakpoint(pid, addr)
	if err != nil {
		deb.Fatal(err)
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

		pc, err := p.PC(pid)
		if err != nil {
			deb.Fatal(err)
		}
		file, line, _ := p.Table.PCToLine(pc)
		fmt.Printf("%s:%d\n", file, line)
		markNextLine(pid, p, file, line)
		err = p.Continue(pid)
		if err != nil {
			deb.Fatal(err)
		}
	}
}

func usage() {
	fmt.Println("Usage: %s <child_process>", os.Args[0])
}

func markNextLine(pid int, p *debugo.Process, file string, line int) {
	next := line
	for {
		next++
		maximum, err := getLines(file)
		if err != nil {
			deb.Println(err)
			return
		}
		if next > maximum {
			return
		}
		pc, _, err := p.Table.LineToPC(file, next)
		if err != nil {
			continue
		}
		err = p.SetBreakpoint(pid, pc)
		if err != nil {
			deb.Fatal(err)
		}
		return
	}
}

func getLines(file string) (int, error) {
	n, ok := lines[file]
	if ok {
		return n, nil
	}
	r, err := os.Open(file)
	if err != nil {
		return 0, err
	}
	buf := make([]byte, 8196)
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := r.Read(buf)
		if err != nil && err != io.EOF {
			return count, err
		}

		count += bytes.Count(buf[:c], lineSep)

		if err == io.EOF {
			break
		}
	}
	lines[file] = count
	return count, nil
}
