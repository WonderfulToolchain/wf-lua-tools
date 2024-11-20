// SPDX-License-Identifier: MIT
//
// Copyright (c) 2024 Adrian "asie" Siekierka

package elf

type Elf struct {
	ElfHeader
	ProgramHeaders []*ProgramHeader
	Sections       []*SectionHeader
	Symbols        []*Symbol
	Relocations    map[*SectionHeader][]*Relocation `json:"-"`
	symtabIdx      int
	symtabShndxIdx int
}

type ElfHeader struct {
	// Identification
	Class         FileClass
	Endian        FileEndian
	HeaderVersion uint8
	ABI           FileABI
	ABIVersion    uint8

	// Header
	Type             FileType
	Machine          MachineType
	Version          uint32
	Entry            uint64
	progHdrOffset    uint64
	secHdrOffset     uint64
	Flags            uint32
	headerSize       uint16
	progHdrEntrySize uint16
	progHdrCount     uint16
	secHdrEntrySize  uint16
	secHdrCount      uint16
	secHdrStrIdx     uint16
}

type ProgramHeader struct {
	Type     ProgramHeaderType
	Flags    ProgramHeaderFlag
	offset   uint64
	VAddr    uint64
	PAddr    uint64
	fileSize uint64
	MemSize  uint64
	Align    uint64
	Data     []byte
}

type SectionHeader struct {
	Name        string
	nameOffset  uint32
	Type        SectionHeaderType
	Flags       SectionHeaderFlag
	Address     uint64
	offset      uint64
	Size        uint32
	Link        uint32
	LinkSection *SectionHeader
	Info        uint32
	InfoSection *SectionHeader
	AddrAlign   uint32
	EntrySize   uint32
	Data        []byte
}

type Symbol struct {
	Name         string
	nameOffset   uint32
	Type         SymbolType
	Binding      SymbolBinding
	Other        uint8
	Section      *SectionHeader
	SectionIndex uint16
	Value        uint64
	Size         uint64
}

type Relocation struct {
	Section     *SectionHeader
	Symbol      *Symbol
	symbolIndex int
	Offset      uint64
	Type        uint32
	Addend      int64
}
