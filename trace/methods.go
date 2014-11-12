package trace

import (
	"syscall"
)



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
