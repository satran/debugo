package main

import (
	"debug/gosym"
	"errors"
	"log"
	"os"
	"syscall"
)

var ErrorTableEmpty = errors.New("Process table should not be empty.")
var BPDoesNotExist = errors.New("Break point does not exist.")

type Process struct {
	Path     string
	Table    *gosym.Table
	Pid      int
	Original map[uint64]byte
}

// NewProcess starts a process and loads its sym table.
func NewProcess(path string) (*Process, error) {
	p := Process{Path: path}
	var err error
	p.Table, err = getTable(path)
	if err != nil {
		return nil, err
	}

	attr := &syscall.ProcAttr{
		Sys:   &syscall.SysProcAttr{Ptrace: true},
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
	}
	p.Pid, err = syscall.ForkExec(path, []string{path}, attr)
	if err != nil {
		return nil, err
	}

	p.Original = make(map[uint64]byte)
	return &p, nil
}

func (p *Process) FuncAddr(function string) (uint64, error) {
	if p.Table == nil {
		return 0, ErrorTableEmpty
	}

	fn := p.Table.LookupFunc(function)
	if fn == nil {
		return 0, nil
	}
	return fn.Entry, nil
}

func (p *Process) SetBreakpoint(pid int, addr uint64) error {
	var text = []byte{0}

	// Read the original data
	_, err := syscall.PtracePeekText(pid, uintptr(addr), text)
	if err != nil {
		deb.Println(err)
		return err
	}

	// Store the original data in the cache, useful when we want to
	// clear the breakpoint
	p.Original[addr] = text[0]

	// Write the breakpoint, very specific to x86 the int 3 instruction.
	text = []byte{0xCC}

	_, err = syscall.PtracePokeText(pid, uintptr(addr), text)
	if err != nil {
		deb.Println(err)
		return err
	}

	return nil
}

func (p *Process) LogAndContinue(pid int) error {
	pc, err := p.PC(pid)
	if err != nil {
		log.Println("Error fetching program counter", err)
		return err
	}

	_, ok := p.Original[pc-1]
	if !ok {
		log.Println("Looks like this was not a break point")
		return err
	}

	file, line, _ := p.Table.PCToLine(pc)
	log.Printf("%s:%d\n", file, line)

	err = p.ContinueBreakpoint(pid, pc-1)
	if err != nil {
		log.Println("Continue break point", err)
		return nil
	}
	return nil
}

func (p *Process) ContinueBreakpoint(pid int, addr uint64) error {
	err := p.ClearBreakpoint(pid, addr)
	if err != nil {
		deb.Println(err)
		return err
	}

	regs := syscall.PtraceRegs{}
	err = syscall.PtraceGetRegs(pid, &regs)
	if err != nil {
		deb.Println(err)
		return err
	}
	regs.SetPC(uint64(addr))
	err = syscall.PtraceSetRegs(pid, &regs)
	if err != nil {
		deb.Println(err)
		return err
	}

	err = syscall.PtraceCont(pid, 0)
	if err != nil {
		deb.Fatal(err)
	}
	return nil
}

func (p *Process) ClearBreakpoint(pid int, addr uint64) error {
	original, ok := p.Original[addr]
	if !ok {
		return BPDoesNotExist
	}

	var text = []byte{original}
	_, err := syscall.PtracePokeText(pid, uintptr(addr), text)
	if err != nil {
		return err
	}

	// Remove the original breakpoint data
	delete(p.Original, addr)

	return nil
}

func (p *Process) PC(pid int) (uint64, error) {
	regs := syscall.PtraceRegs{}
	err := syscall.PtraceGetRegs(p.Pid, &regs)
	if err != nil {
		return 0, err
	}
	return regs.PC(), nil
}
