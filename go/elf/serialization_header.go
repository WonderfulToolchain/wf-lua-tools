// SPDX-License-Identifier: MIT
//
// Copyright (c) 2024 Adrian "asie" Siekierka

package elf

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type elfHeader32 struct {
	Type             uint16
	Machine          uint16
	Version          uint32
	Entry            uint32
	ProgHdrOff       uint32
	SecHdrOff        uint32
	Flags            uint32
	HeaderSize       uint16
	ProgHdrEntrySize uint16
	ProgHdrCount     uint16
	SecHdrEntrySize  uint16
	SecHdrCount      uint16
	SecHdrStrIndex   uint16
}

type elfHeader64 struct {
	Type             uint16
	Machine          uint16
	Version          uint32
	Entry            uint64
	ProgHdrOff       uint64
	SecHdrOff        uint64
	Flags            uint32
	HeaderSize       uint16
	ProgHdrEntrySize uint16
	ProgHdrCount     uint16
	SecHdrEntrySize  uint16
	SecHdrCount      uint16
	SecHdrStrIndex   uint16
}

func (e *Elf) sizeElfHeader() int {
	// Add 16 bytes of ELF identification section
	if e.Class == ELFCLASS64 {
		return binary.Size(&elfHeader64{}) + 16
	} else {
		return binary.Size(&elfHeader32{}) + 16
	}
}

func (e *Elf) readElfHeader(r io.Reader) error {
	ident := make([]byte, 16)

	if _, err := r.Read(ident); err != nil {
		return err
	}

	if ident[0] != 0x7F || ident[1] != 0x45 || ident[2] != 0x4C || ident[3] != 0x46 {
		return errors.New("invalid magic")
	}

	e.Class = FileClass(ident[4])
	e.Endian = FileEndian(ident[5])
	e.HeaderVersion = ident[6]

	if e.Class == ELFCLASS64 {
		e.ABI = FileABI(ident[7])
		e.ABIVersion = ident[8]

		var fh elfHeader64
		if err := binary.Read(r, e.GetByteOrder(), &fh); err != nil {
			return err
		}

		e.Type = FileType(fh.Type)
		e.Machine = MachineType(fh.Machine)
		e.Version = fh.Version
		e.Entry = fh.Entry
		e.progHdrOffset = fh.ProgHdrOff
		e.secHdrOffset = fh.SecHdrOff
		e.Flags = fh.Flags
		e.headerSize = fh.HeaderSize
		e.progHdrEntrySize = fh.ProgHdrEntrySize
		e.progHdrCount = fh.ProgHdrCount
		e.secHdrEntrySize = fh.SecHdrEntrySize
		e.secHdrCount = fh.SecHdrCount
		e.secHdrStrIdx = fh.SecHdrStrIndex
	} else if e.Class == ELFCLASS32 {
		e.ABI = 0
		e.ABIVersion = 0

		var fh elfHeader32
		if err := binary.Read(r, e.GetByteOrder(), &fh); err != nil {
			return err
		}

		e.Type = FileType(fh.Type)
		e.Machine = MachineType(fh.Machine)
		e.Version = fh.Version
		e.Entry = uint64(fh.Entry)
		e.progHdrOffset = uint64(fh.ProgHdrOff)
		e.secHdrOffset = uint64(fh.SecHdrOff)
		e.Flags = fh.Flags
		e.headerSize = fh.HeaderSize
		e.progHdrEntrySize = fh.ProgHdrEntrySize
		e.progHdrCount = fh.ProgHdrCount
		e.secHdrEntrySize = fh.SecHdrEntrySize
		e.secHdrCount = fh.SecHdrCount
		e.secHdrStrIdx = fh.SecHdrStrIndex
	} else {
		return errors.New(fmt.Sprint("invalid class: ", e.Class))
	}

	if e.secHdrStrIdx == SHN_XINDEX {
		panic("TODO: SHN_XINDEX support")
	}

	return nil
}

func (e *Elf) writeElfHeader(w io.Writer) error {
	ident := make([]byte, 16)

	ident[0] = 0x7F
	ident[1] = 0x45
	ident[2] = 0x4C
	ident[3] = 0x46

	ident[4] = uint8(e.Class)
	ident[5] = uint8(e.Endian)
	ident[6] = uint8(e.HeaderVersion)
	ident[7] = uint8(e.ABI)
	ident[8] = uint8(e.ABIVersion)

	if _, err := w.Write(ident); err != nil {
		return err
	}

	if e.Class == ELFCLASS64 {
		var fh elfHeader64

		fh.Type = uint16(e.Type)
		fh.Machine = uint16(e.Machine)
		fh.Version = e.Version
		fh.Entry = e.Entry
		fh.ProgHdrOff = e.progHdrOffset
		fh.SecHdrOff = e.secHdrOffset
		fh.Flags = e.Flags
		fh.HeaderSize = e.headerSize
		fh.ProgHdrEntrySize = e.progHdrEntrySize
		fh.ProgHdrCount = e.progHdrCount
		fh.SecHdrEntrySize = e.secHdrEntrySize
		fh.SecHdrCount = e.secHdrCount
		fh.SecHdrStrIndex = e.secHdrStrIdx

		if err := binary.Write(w, e.GetByteOrder(), &fh); err != nil {
			return err
		}
	} else if e.Class == ELFCLASS32 {
		var fh elfHeader32

		fh.Type = uint16(e.Type)
		fh.Machine = uint16(e.Machine)
		fh.Version = e.Version
		fh.Entry = uint32(e.Entry)
		fh.ProgHdrOff = uint32(e.progHdrOffset)
		fh.SecHdrOff = uint32(e.secHdrOffset)
		fh.Flags = e.Flags
		fh.HeaderSize = e.headerSize
		fh.ProgHdrEntrySize = e.progHdrEntrySize
		fh.ProgHdrCount = e.progHdrCount
		fh.SecHdrEntrySize = e.secHdrEntrySize
		fh.SecHdrCount = e.secHdrCount
		fh.SecHdrStrIndex = e.secHdrStrIdx

		if err := binary.Write(w, e.GetByteOrder(), &fh); err != nil {
			return err
		}
	}

	return nil
}
