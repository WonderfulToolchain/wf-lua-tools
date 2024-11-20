// SPDX-License-Identifier: MIT
//
// Copyright (c) 2024 Adrian "asie" Siekierka

package elf

import (
	"encoding/binary"
	"io"
)

func (e *Elf) readString(r io.ReadSeeker, idx int, offset uint64) (error, string) {
	if _, err := r.Seek(int64(e.Sections[idx].offset+offset), io.SeekStart); err != nil {
		return err, ""
	}
	return readString(r)
}

func (e *Elf) GetByteOrder() binary.ByteOrder {
	if e.Endian == ELFDATA2MSB {
		return binary.BigEndian
	} else {
		return binary.LittleEndian
	}
}

func ReadELF(r io.ReadSeeker) (error, *Elf) {
	e := &Elf{}
	e.Relocations = make(map[*SectionHeader][]*Relocation)

	// Read main header
	if err := e.readElfHeader(r); err != nil {
		return err, nil
	}

	// Read program headers
	r.Seek(int64(e.progHdrOffset), io.SeekStart)
	for i := 0; i < int(e.progHdrCount); i++ {
		err, hdr := e.readProgramHeader(r)
		if err != nil {
			return err, nil
		}
		e.ProgramHeaders = append(e.ProgramHeaders, hdr)
	}
	for i := 0; i < int(e.progHdrCount); i++ {
		hdr := e.ProgramHeaders[i]
		if hdr.fileSize > 0 {
			r.Seek(int64(hdr.offset), io.SeekStart)
			hdr.Data = make([]byte, hdr.fileSize)
			if _, err := r.Read(hdr.Data); err != nil {
				return err, nil
			}
		}
	}

	// Read section headers
	r.Seek(int64(e.secHdrOffset), io.SeekStart)
	for i := 0; i < int(e.secHdrCount); i++ {
		err, hdr := e.readSectionHeader(r)
		if err != nil {
			return err, nil
		}
		e.Sections = append(e.Sections, hdr)
		if hdr.Type == SHT_SYMTAB {
			e.symtabIdx = i
		} else if hdr.Type == SHT_SYMTAB_SHNDX {
			e.symtabShndxIdx = i
		}
	}

	for i := 0; i < int(e.secHdrCount); i++ {
		hdr := e.Sections[i]
		if hdr.Link < SHN_LORESERVE {
			hdr.LinkSection = e.Sections[hdr.Link]
		}
		if hdr.Info < SHN_LORESERVE && hdr.Type.HasSectionInInfo() {
			hdr.InfoSection = e.Sections[hdr.Info]
		}
	}

	// Read shstrtab
	if e.secHdrStrIdx != SHN_UNDEF {
		for i := 0; i < int(e.secHdrCount); i++ {
			hdr := e.Sections[i]
			err, s := e.readString(r, int(e.secHdrStrIdx), uint64(hdr.nameOffset))
			if err != nil {
				return err, nil
			}
			hdr.Name = s
		}
	}

	// Read symbols
	if e.symtabIdx > 0 {
		symtab := e.Sections[e.symtabIdx]
		symbolCount := symtab.Size / symtab.EntrySize
		r.Seek(int64(symtab.offset), io.SeekStart)
		for i := 0; i < int(symbolCount); i++ {
			err, sym := e.readSymbol(r, symtab)
			if err != nil {
				return err, nil
			}
			e.Symbols = append(e.Symbols, sym)
		}
	}

	// Read relocations
	for i := 0; i < int(e.secHdrCount); i++ {
		hdr := e.Sections[i]
		if hdr.Type == SHT_REL || hdr.Type == SHT_RELA {
			sec := e.Sections[int(hdr.Info)]
			relCount := hdr.Size / hdr.EntrySize
			r.Seek(int64(hdr.offset), io.SeekStart)
			for i := 0; i < int(relCount); i++ {
				err, rel := e.readRelocation(r, sec, hdr.Type)
				if err != nil {
					return err, nil
				}
				e.Relocations[sec] = append(e.Relocations[sec], rel)
			}
		}
	}

	// Drop already parsed sections. Do this last!
	sections := make([]*SectionHeader, 0)
	for _, sh := range e.Sections {
		if sh.Type == SHT_REL || sh.Type == SHT_RELA {
			continue
		}

		if sh.Type == SHT_SYMTAB {
			continue
		}

		if sh.Type == SHT_STRTAB || sh.Type == SHT_SYMTAB_SHNDX {
			continue
		}

		sections = append(sections, sh)
	}
	e.Sections = sections

	return nil, e
}
