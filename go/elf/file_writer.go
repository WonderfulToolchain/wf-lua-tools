// SPDX-License-Identifier: MIT
//
// Copyright (c) 2024 Adrian "asie" Siekierka

package elf

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"slices"
	"strings"
)

type stringTable struct {
	strings map[string]uint32
	pos     uint32
}

func newStringTable() stringTable {
	return stringTable{
		strings: make(map[string]uint32),
		pos:     0,
	}
}

func (e *stringTable) Add(s string) uint32 {
	// TODO: Support substrings
	if val, ok := e.strings[s]; ok {
		return val
	}
	sPos := e.pos
	e.pos += uint32(len(s)) + 1
	e.strings[s] = sPos
	return sPos
}

func (e *stringTable) ToData() []byte {
	data := make([]byte, e.pos)
	for s, i := range e.strings {
		data = slices.Replace(data, int(i), int(i)+len(s), []byte(s)...)
	}
	return data
}

func (e *Elf) Write(w io.Writer) error {
	writeOffset := int64(0)

	// Layout file contents:
	// - file header
	// - program headers
	// - program data
	// - section headers
	// - section data
	// - symbol table

	sections := e.Sections

	if len(sections) > 65500 {
		panic(fmt.Sprint("TODO: Unsupported section count: ", len(sections)))
	}

	// Sort symbol table to put STB_LOCAL symbols first
	slices.SortFunc(e.Symbols, func(a *Symbol, b *Symbol) int {
		if a.Binding != b.Binding {
			return int(a.Binding) - int(b.Binding)
		} else {
			return strings.Compare(a.Name, b.Name)
		}
	})

	// Create new symbol table, string table sections
	sectionStringTable := newStringTable()
	stringTable := newStringTable()

	sectionStringTableSection := &SectionHeader{
		Name: ".shstrtab",
		Type: SHT_STRTAB,
	}
	stringTableSection := &SectionHeader{
		Name: ".strtab",
		Type: SHT_STRTAB,
	}
	symbolTableSection := &SectionHeader{
		Name:      ".symtab",
		Type:      SHT_SYMTAB,
		EntrySize: uint32(e.sizeSymbol()),
	}
	e.secHdrStrIdx = uint16(len(sections))
	strTabIdx := len(sections) + 1
	symTabIdx := len(sections) + 2
	sections = append(sections, sectionStringTableSection, stringTableSection, symbolTableSection)
	symbolTableSection.Link = uint32(strTabIdx)

	// Create relocation table sections
	for parentSection, relocations := range e.Relocations {
		relType := SHT_REL
		relName := ".rel"
		for _, rel := range relocations {
			if rel.Addend != 0 {
				relType = SHT_RELA
				relName = ".rela"
				break
			}
		}
		relSection := &SectionHeader{
			Name:      relName + parentSection.Name,
			Type:      relType,
			EntrySize: uint32(e.sizeRelocation(relType)),
			Info:      uint32(slices.Index(sections, parentSection)),
			Link:      uint32(symTabIdx),
		}

		var relBuffer bytes.Buffer
		relWriter := bufio.NewWriter(&relBuffer)

		for _, rel := range relocations {
			// TODO: Reduce O(n^2)
			if rel.Symbol != nil {
				rel.symbolIndex = slices.Index(e.Symbols, rel.Symbol)
			}
			if err := e.writeRelocation(relWriter, parentSection, relType, rel); err != nil {
				return err
			}
		}

		if err := relWriter.Flush(); err != nil {
			return err
		}
		relSection.Data = relBuffer.Bytes()
		sections = append(sections, relSection)
	}

	// Populate section string table
	for _, sh := range sections {
		sh.nameOffset = sectionStringTable.Add(sh.Name)
	}

	// Populate symbol table
	globalBindingSet := false
	var symtabBuffer bytes.Buffer
	symtabWriter := bufio.NewWriter(&symtabBuffer)
	for i, sym := range e.Symbols {
		sym.nameOffset = stringTable.Add(sym.Name)
		// TODO: Reduce O(n^2)
		if sym.Section != nil {
			sym.SectionIndex = uint16(slices.Index(sections, sym.Section))
		}
		if !globalBindingSet && sym.Binding != STB_LOCAL {
			symbolTableSection.Info = uint32(i)
			globalBindingSet = true
		}
		if err := e.writeSymbol(symtabWriter, sym); err != nil {
			return err
		}
	}
	if err := symtabWriter.Flush(); err != nil {
		return err
	}
	symbolTableSection.Data = symtabBuffer.Bytes()

	// Write string tables
	sectionStringTableSection.Data = sectionStringTable.ToData()
	stringTableSection.Data = stringTable.ToData()

	// Layout file header
	e.headerSize = uint16(e.sizeElfHeader())
	writeOffset += int64(e.headerSize)

	// Layout program headers
	e.progHdrEntrySize = uint16(e.sizeProgramHeader())
	e.progHdrCount = uint16(len(e.ProgramHeaders))
	if e.progHdrCount > 0 {
		e.progHdrOffset = uint64(writeOffset)
		writeOffset += int64(e.progHdrCount) * int64(e.progHdrEntrySize)
	} else {
		e.progHdrOffset = 0
	}

	// Layout program data
	for _, ph := range e.ProgramHeaders {
		ph.fileSize = uint64(len(ph.Data))
		ph.offset = uint64(writeOffset)
		writeOffset += int64(ph.fileSize)
	}

	// Layout section headers
	e.secHdrEntrySize = uint16(e.sizeSectionHeader())
	e.secHdrCount = uint16(len(sections))
	if e.secHdrCount > 0 {
		e.secHdrOffset = uint64(writeOffset)
		writeOffset += int64(e.secHdrCount) * int64(e.secHdrEntrySize)
	} else {
		e.secHdrOffset = 0
	}

	// Layout section data
	for _, sh := range sections {
		if sh.Type.HasDataInFile() {
			sh.Size = uint32(len(sh.Data))
			sh.offset = uint64(writeOffset)
			writeOffset += int64(sh.Size)
		}
	}

	// Write file header
	if err := e.writeElfHeader(w); err != nil {
		return err
	}

	// Write program headers
	for _, ph := range e.ProgramHeaders {
		if err := e.writeProgramHeader(w, ph); err != nil {
			return err
		}
	}

	// Write program data
	for _, ph := range e.ProgramHeaders {
		if _, err := w.Write(ph.Data); err != nil {
			return err
		}
	}

	// Write section headers
	for _, sh := range sections {
		if err := e.writeSectionHeader(w, sh); err != nil {
			return err
		}
	}

	// Write section data
	for _, sh := range sections {
		if sh.Type.HasDataInFile() {
			if _, err := w.Write(sh.Data); err != nil {
				return err
			}
		}
	}

	return nil
}
