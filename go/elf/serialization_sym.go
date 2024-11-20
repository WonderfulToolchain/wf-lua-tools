// SPDX-License-Identifier: MIT
//
// Copyright (c) 2024 Adrian "asie" Siekierka

package elf

import (
	"encoding/binary"
	"io"
)

type symbol32 struct {
	Name         uint32
	Value        uint32
	Size         uint32
	Info         uint8
	Other        uint8
	SectionIndex uint16
}

type symbol64 struct {
	Name         uint32
	Info         uint8
	Other        uint8
	SectionIndex uint16
	Value        uint64
	Size         uint64
}

func (e *Elf) sizeSymbol() int {
	if e.Class == ELFCLASS64 {
		return binary.Size(&symbol64{})
	} else {
		return binary.Size(&symbol32{})
	}
}

func (e *Elf) readSymbol(r io.ReadSeeker, parent *SectionHeader) (error, *Symbol) {
	var err error
	var ofs int64
	var result Symbol

	if e.Class == ELFCLASS64 {
		var sh symbol64
		if err = binary.Read(r, e.GetByteOrder(), &sh); err != nil {
			return err, nil
		}
		ofs, err = r.Seek(0, io.SeekCurrent)
		if err != nil {
			return err, nil
		}

		result.nameOffset = sh.Name
		result.Type = SymbolType(sh.Info & 0xF)
		result.Binding = SymbolBinding(sh.Info >> 4)
		result.Other = sh.Other
		result.SectionIndex = sh.SectionIndex
		result.Value = sh.Value
		result.Size = sh.Size
	} else {
		var sh symbol32
		if err = binary.Read(r, e.GetByteOrder(), &sh); err != nil {
			return err, nil
		}
		ofs, err = r.Seek(0, io.SeekCurrent)
		if err != nil {
			return err, nil
		}

		result.nameOffset = sh.Name
		result.Type = SymbolType(sh.Info & 0xF)
		result.Binding = SymbolBinding(sh.Info >> 4)
		result.Other = sh.Other
		result.SectionIndex = sh.SectionIndex
		result.Value = uint64(sh.Value)
		result.Size = uint64(sh.Size)
	}

	err, s := e.readString(r, int(parent.Link), uint64(result.nameOffset))
	if err != nil {
		return err, nil
	}
	result.Name = s

	if result.SectionIndex == SHN_XINDEX {
		panic("TODO: SHN_XINDEX support")
	} else if result.SectionIndex < SHN_LORESERVE && result.SectionIndex > 0 {
		result.Section = e.Sections[int(result.SectionIndex)]
		result.SectionIndex = 0
	}

	if _, err := r.Seek(ofs, io.SeekStart); err != nil {
		return err, nil
	}

	return nil, &result
}

func (e *Elf) writeSymbol(w io.Writer, input *Symbol) error {
	if e.Class == ELFCLASS64 {
		var sh symbol64

		sh.Name = input.nameOffset
		sh.Info = uint8(input.Type) | (uint8(input.Binding) << 4)
		sh.Other = input.Other
		sh.SectionIndex = input.SectionIndex
		sh.Value = input.Value
		sh.Size = input.Size

		if err := binary.Write(w, e.GetByteOrder(), &sh); err != nil {
			return err
		}
	} else {
		var sh symbol32

		sh.Name = input.nameOffset
		sh.Info = uint8(input.Type) | (uint8(input.Binding) << 4)
		sh.Other = input.Other
		sh.SectionIndex = input.SectionIndex
		sh.Value = uint32(input.Value)
		sh.Size = uint32(input.Size)

		if err := binary.Write(w, e.GetByteOrder(), &sh); err != nil {
			return err
		}
	}

	return nil
}
