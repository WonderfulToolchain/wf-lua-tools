// SPDX-License-Identifier: MIT
//
// Copyright (c) 2024 Adrian "asie" Siekierka

package elf

import (
	"encoding/binary"
	"io"
)

type sectionHeader32 struct {
	Name      uint32
	Type      uint32
	Flags     uint32
	Address   uint32
	Offset    uint32
	Size      uint32
	Link      uint32
	Info      uint32
	AddrAlign uint32
	EntrySize uint32
}

type sectionHeader64 struct {
	Name      uint32
	Type      uint32
	Flags     uint32
	Address   uint64
	Offset    uint64
	Size      uint32
	Link      uint32
	Info      uint32
	AddrAlign uint32
	EntrySize uint32
}

func (e *Elf) sizeSectionHeader() int {
	if e.Class == ELFCLASS64 {
		return binary.Size(&sectionHeader64{})
	} else {
		return binary.Size(&sectionHeader32{})
	}
}

func (e *Elf) readSectionHeader(r io.ReadSeeker) (error, *SectionHeader) {
	var result SectionHeader

	if e.Class == ELFCLASS64 {
		var sh sectionHeader64
		if err := binary.Read(r, e.GetByteOrder(), &sh); err != nil {
			return err, nil
		}

		result.nameOffset = sh.Name
		result.Type = SectionHeaderType(sh.Type)
		result.Flags = SectionHeaderFlag(sh.Flags)
		result.Address = sh.Address
		result.offset = sh.Offset
		result.Size = sh.Size
		result.Link = sh.Link
		result.Info = sh.Info
		result.AddrAlign = sh.AddrAlign
		result.EntrySize = sh.EntrySize
	} else {
		var sh sectionHeader32
		if err := binary.Read(r, e.GetByteOrder(), &sh); err != nil {
			return err, nil
		}

		result.nameOffset = sh.Name
		result.Type = SectionHeaderType(sh.Type)
		result.Flags = SectionHeaderFlag(sh.Flags)
		result.Address = uint64(sh.Address)
		result.offset = uint64(sh.Offset)
		result.Size = sh.Size
		result.Link = sh.Link
		result.Info = sh.Info
		result.AddrAlign = sh.AddrAlign
		result.EntrySize = sh.EntrySize
	}

	if result.Size > 0 && result.Type.HasDataInFile() {
		pos, _ := r.Seek(0, io.SeekCurrent)

		if _, err := r.Seek(int64(result.offset), io.SeekStart); err != nil {
			return err, nil
		}
		result.Data = make([]byte, result.Size)
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

func (e *Elf) writeSectionHeader(w io.Writer, input *SectionHeader) error {
	if e.Class == ELFCLASS64 {
		var sh sectionHeader64

		sh.Name = input.nameOffset
		sh.Type = uint32(input.Type)
		sh.Flags = uint32(input.Flags)
		sh.Address = input.Address
		sh.Offset = input.offset
		sh.Size = input.Size
		sh.Link = input.Link
		sh.Info = input.Info
		sh.AddrAlign = input.AddrAlign
		sh.EntrySize = input.EntrySize

		if err := binary.Write(w, e.GetByteOrder(), &sh); err != nil {
			return err
		}
	} else {
		var sh sectionHeader32

		sh.Name = input.nameOffset
		sh.Type = uint32(input.Type)
		sh.Flags = uint32(input.Flags)
		sh.Address = uint32(input.Address)
		sh.Offset = uint32(input.offset)
		sh.Size = input.Size
		sh.Link = input.Link
		sh.Info = input.Info
		sh.AddrAlign = input.AddrAlign
		sh.EntrySize = input.EntrySize

		if err := binary.Write(w, e.GetByteOrder(), &sh); err != nil {
			return err
		}
	}

	return nil
}
