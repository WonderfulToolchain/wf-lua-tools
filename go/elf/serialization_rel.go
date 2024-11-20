// SPDX-License-Identifier: MIT
//
// Copyright (c) 2024 Adrian "asie" Siekierka

package elf

import (
	"encoding/binary"
	"fmt"
	"io"
)

type rel32 struct {
	Offset uint32
	Info   uint32
}

type rel64 struct {
	Offset uint64
	Info   uint64
}

type rela32 struct {
	Offset uint32
	Info   uint32
	Addend int32
}

type rela64 struct {
	Offset uint64
	Info   uint64
	Addend int64
}

func (e *Elf) sizeRelocation(t SectionHeaderType) int {
	if e.Class == ELFCLASS64 {
		if t == SHT_RELA {
			return binary.Size(&rela64{})
		} else {
			return binary.Size(&rel64{})
		}
	} else {
		if t == SHT_RELA {
			return binary.Size(&rela32{})
		} else {
			return binary.Size(&rel32{})
		}
	}
}

func (e *Elf) readRelocation(r io.Reader, s *SectionHeader, t SectionHeaderType) (error, *Relocation) {
	var err error
	var result Relocation
	result.Section = s

	if e.Class == ELFCLASS64 {
		if t == SHT_RELA {
			var rel rela64
			if err = binary.Read(r, e.GetByteOrder(), &rel); err != nil {
				return err, nil
			}
			result.Offset = rel.Offset
			result.symbolIndex = int(rel.Info >> 32)
			result.Type = uint32(rel.Info)
			result.Addend = rel.Addend
		} else if t == SHT_REL {
			var rel rel64
			if err = binary.Read(r, e.GetByteOrder(), &rel); err != nil {
				return err, nil
			}
			result.Offset = rel.Offset
			result.symbolIndex = int(rel.Info >> 32)
			result.Type = uint32(rel.Info)
		} else {
			return fmt.Errorf("unknown type: %d", t), nil
		}
	} else {
		if t == SHT_RELA {
			var rel rela32
			if err = binary.Read(r, e.GetByteOrder(), &rel); err != nil {
				return err, nil
			}
			result.Offset = uint64(rel.Offset)
			result.symbolIndex = int(rel.Info >> 8)
			result.Type = uint32(rel.Info & 0xFF)
			result.Addend = int64(rel.Addend)
		} else if t == SHT_REL {
			var rel rel32
			if err = binary.Read(r, e.GetByteOrder(), &rel); err != nil {
				return err, nil
			}
			result.Offset = uint64(rel.Offset)
			result.symbolIndex = int(rel.Info >> 8)
			result.Type = uint32(rel.Info & 0xFF)
		} else {
			return fmt.Errorf("unknown type: %d", t), nil
		}
	}

	result.Symbol = e.Symbols[result.symbolIndex]
	return nil, &result
}

func (e *Elf) writeRelocation(w io.Writer, s *SectionHeader, t SectionHeaderType, input *Relocation) error {
	if e.Class == ELFCLASS64 {
		if t == SHT_RELA {
			var rel rela64

			rel.Offset = input.Offset
			rel.Info = (uint64(input.symbolIndex) << 32) | uint64(input.Type)
			rel.Addend = input.Addend

			if err := binary.Write(w, e.GetByteOrder(), &rel); err != nil {
				return err
			}
		} else {
			var rel rel64

			rel.Offset = input.Offset
			rel.Info = (uint64(input.symbolIndex) << 32) | uint64(input.Type)

			if err := binary.Write(w, e.GetByteOrder(), &rel); err != nil {
				return err
			}
		}
	} else {
		if t == SHT_RELA {
			var rel rela32

			rel.Offset = uint32(input.Offset)
			rel.Info = (uint32(input.symbolIndex) << 8) | uint32(input.Type)
			rel.Addend = int32(input.Addend)

			if err := binary.Write(w, e.GetByteOrder(), &rel); err != nil {
				return err
			}
		} else {
			var rel rel32

			rel.Offset = uint32(input.Offset)
			rel.Info = (uint32(input.symbolIndex) << 8) | uint32(input.Type)

			if err := binary.Write(w, e.GetByteOrder(), &rel); err != nil {
				return err
			}
		}
	}

	return nil
}
