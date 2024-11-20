// SPDX-License-Identifier: MIT
//
// Copyright (c) 2024 Adrian "asie" Siekierka

package elf

import (
	"encoding/binary"
	"io"
)

type programHeader32 struct {
	Type     uint32
	Offset   uint32
	VAddr    uint32
	PAddr    uint32
	FileSize uint32
	MemSize  uint32
	Flags    uint32
	Align    uint32
}

type programHeader64 struct {
	Type     uint32
	Flags    uint32
	Offset   uint64
	VAddr    uint64
	PAddr    uint64
	FileSize uint64
	MemSize  uint64
	Align    uint64
}

func (e *Elf) sizeProgramHeader() int {
	if e.Class == ELFCLASS64 {
		return binary.Size(&programHeader64{})
	} else {
		return binary.Size(&programHeader32{})
	}
}

func (e *Elf) readProgramHeader(r io.ReadSeeker) (error, *ProgramHeader) {
	var result ProgramHeader

	if e.Class == ELFCLASS64 {
		var ph programHeader64
		if err := binary.Read(r, e.GetByteOrder(), &ph); err != nil {
			return err, nil
		}

		result.Type = ProgramHeaderType(ph.Type)
		result.Flags = ProgramHeaderFlag(ph.Flags)
		result.offset = ph.Offset
		result.VAddr = ph.VAddr
		result.PAddr = ph.PAddr
		result.fileSize = ph.FileSize
		result.MemSize = ph.MemSize
		result.Align = ph.Align
	} else {
		var ph programHeader32
		if err := binary.Read(r, e.GetByteOrder(), &ph); err != nil {
			return err, nil
		}

		result.Type = ProgramHeaderType(ph.Type)
		result.Flags = ProgramHeaderFlag(ph.Flags)
		result.offset = uint64(ph.Offset)
		result.VAddr = uint64(ph.VAddr)
		result.PAddr = uint64(ph.PAddr)
		result.fileSize = uint64(ph.FileSize)
		result.MemSize = uint64(ph.MemSize)
		result.Align = uint64(ph.Align)
	}

	if result.fileSize > 0 {
		pos, _ := r.Seek(0, io.SeekCurrent)

		if _, err := r.Seek(int64(result.offset), io.SeekStart); err != nil {
			return err, nil
		}
		result.Data = make([]byte, result.fileSize)
		if _, err := r.Read(result.Data); err != nil {
			return err, nil
		}

		if _, err := r.Seek(pos, io.SeekStart); err != nil {
			return err, nil
		}
	} else {
		result.Data = nil
	}

	return nil, &result
}

func (e *Elf) writeProgramHeader(w io.Writer, input *ProgramHeader) error {
	if e.Class == ELFCLASS64 {
		var ph programHeader64

		ph.Type = uint32(input.Type)
		ph.Flags = uint32(input.Flags)
		ph.Offset = input.offset
		ph.VAddr = input.VAddr
		ph.PAddr = input.PAddr
		ph.FileSize = input.fileSize
		ph.MemSize = input.MemSize
		ph.Align = input.Align

		if err := binary.Write(w, e.GetByteOrder(), &ph); err != nil {
			return err
		}
	} else {
		var ph programHeader32

		ph.Type = uint32(input.Type)
		ph.Flags = uint32(input.Flags)
		ph.Offset = uint32(input.offset)
		ph.VAddr = uint32(input.VAddr)
		ph.PAddr = uint32(input.PAddr)
		ph.FileSize = uint32(input.fileSize)
		ph.MemSize = uint32(input.MemSize)
		ph.Align = uint32(input.Align)

		if err := binary.Write(w, e.GetByteOrder(), &ph); err != nil {
			return err
		}
	}

	return nil
}
