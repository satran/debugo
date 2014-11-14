package trace

import (
	"syscall"
)

func wait(pid int) (int, syscall.WaitStatus, error) {
	var w syscall.WaitStatus
	pid, err := syscall.Wait4(pid, &w, syscall.WALL, nil)
	return pid, w, err
}

func setbreakpoint(pid int, addr uint64) error {
	var text = []byte{0}

	// Read the original data
	_, err := syscall.PtracePeekText(pid, uintptr(addr), text)
	if err != nil {
		return err
	}

	// If there exists a breakpoint there is no point replacing the content.
	// We shall only lose the correct data.
	if text[0] == 0xCC {
		return nil
	}
	
	// Store the original data in the cache, useful when we want to
	// clear the breakpoint
	breakpoints[addr] = text[0]

	// Write the breakpoint, very specific to x86 the int 3 instruction.
	text = []byte{0xCC}

	_, err = syscall.PtracePokeText(pid, uintptr(addr), text)
	if err != nil {
		return err
	}
	return nil
}

func continueBreakpoint(pid int, addr uint64) error {
	err := clearBreakpoint(pid, addr)
	if err != nil {
		return err
	}

	regs := syscall.PtraceRegs{}
	err = syscall.PtraceGetRegs(pid, &regs)
	if err != nil {
		return err
	}
	regs.SetPC(uint64(addr))
	err = syscall.PtraceSetRegs(pid, &regs)
	if err != nil {
		return err
	}

	// Set the breakpoint again so that it stops correctly. To do so lets move
	// one instruction ahead, set the break point in the old addr and continue
	err = syscall.PtraceSingleStep(pid)
	if err != nil {
		return err
	}
	pid, w, err := wait(pid)
	if w.Exited() {
		return ProcessExitedError
	}
	err =  setbreakpoint(pid, addr)	
	if err != nil {
		return err
	}

	return syscall.PtraceCont(pid, 0)
}

func clearBreakpoint(pid int, addr uint64) error {
	original, ok := breakpoints[addr]
	if !ok {
		return NotBreakPointError
	}

	var text = []byte{original}
	_, err := syscall.PtracePokeText(pid, uintptr(addr), text)
	if err != nil {
		return err
	}

	// Remove the original breakpoint data
	delete(breakpoints, addr)

	return nil
}


func pc(pid int) (uint64, error) {
	regs := syscall.PtraceRegs{}
	err := syscall.PtraceGetRegs(pid, &regs)
	if err != nil {
		return 0, err
	}
	return regs.PC(), nil
}
