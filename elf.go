package main

import (
	"os"
	"fmt"
	"debug/elf"
	"debug/gosym"
)

func getTable(file string) (*gosym.Table, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	textStart, _, symtab, pclntab, err := loadTables(f)
	if err != nil {
		return nil, err
	}

	pcln := gosym.NewLineTable(pclntab, textStart)
	return gosym.NewTable(symtab, pcln)	
}

func loadTables(f *os.File) (textStart uint64, textData, symtab, pclntab []byte, err error) {
	if obj, err := elf.NewFile(f); err == nil {
		if sect := obj.Section(".text"); sect != nil {
			textStart = sect.Addr
			textData, _ = sect.Data()
		}
		if sect := obj.Section(".gosymtab"); sect != nil {
			if symtab, err = sect.Data(); err != nil {
				return 0, nil, nil, nil, err
			}
		}
		if sect := obj.Section(".gopclntab"); sect != nil {
			if pclntab, err = sect.Data(); err != nil {
				return 0, nil, nil, nil, err
			}
		}
		return textStart, textData, symtab, pclntab, nil
	}
	return 0, nil, nil, nil, fmt.Errorf("unrecognized binary format")
}
