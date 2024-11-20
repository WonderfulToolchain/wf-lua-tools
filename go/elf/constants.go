// SPDX-License-Identifier: MIT
//
// Copyright (c) 2024 Adrian "asie" Siekierka

package elf

type FileClass uint8

const (
	ELFCLASS32 FileClass = 1
	ELFCLASS64 FileClass = 2
)

type FileEndian uint8

const (
	ELFDATA2LSB FileEndian = 1
	ELFDATA2MSB FileEndian = 2
)

type FileABI uint8

type FileType uint16

const (
	ET_NONE   FileType = 0
	ET_REL    FileType = 1
	ET_EXEC   FileType = 2
	ET_DYN    FileType = 3
	ET_CORE   FileType = 4
	ET_LOOS   FileType = 0xFE00
	ET_HIOS   FileType = 0xFEFF
	ET_LOPROC FileType = 0xFF00
	ET_HIPROC FileType = 0xFFFF
)

type MachineType uint16

const (
	EM_NONE MachineType = 0  // None.
	EM_386  MachineType = 3  // 386-compatible processor; also used by gcc-ia16 to denote 8086-compatible processor.
	EM_MIPS MachineType = 8  // MIPS processor
	EM_ARM  MachineType = 40 // ARM processor
)

// Section header index
const (
	SHN_UNDEF     = 0
	SHN_LORESERVE = 0xFF00
	SHN_ABS       = 0xFFF1
	SHN_COMMON    = 0xFFF2
	SHN_XINDEX    = 0xFFFF
)

type SectionHeaderType uint32

const (
	SHT_NULL          SectionHeaderType = 0
	SHT_PROGBITS      SectionHeaderType = 1
	SHT_SYMTAB        SectionHeaderType = 2
	SHT_STRTAB        SectionHeaderType = 3
	SHT_RELA          SectionHeaderType = 4
	SHT_HASH          SectionHeaderType = 5
	SHT_DYNAMIC       SectionHeaderType = 6
	SHT_NOTE          SectionHeaderType = 7
	SHT_NOBITS        SectionHeaderType = 8
	SHT_REL           SectionHeaderType = 9
	SHT_SHLIB         SectionHeaderType = 10
	SHT_DYNSYM        SectionHeaderType = 11
	SHT_INIT_ARRAY    SectionHeaderType = 14
	SHT_FINI_ARRAY    SectionHeaderType = 15
	SHT_PREINIT_ARRAY SectionHeaderType = 16
	SHT_GROUP         SectionHeaderType = 17
	SHT_SYMTAB_SHNDX  SectionHeaderType = 18
)

func (s SectionHeaderType) HasSectionInInfo() bool {
	return s == SHT_REL || s == SHT_RELA
}

func (s SectionHeaderType) HasDataInFile() bool {
	return s != SHT_NOBITS
}

// Section header flags
type SectionHeaderFlag uint32

const (
	SHF_WRITE            SectionHeaderFlag = 0x00000001
	SHF_ALLOC            SectionHeaderFlag = 0x00000002
	SHF_EXECINSTR        SectionHeaderFlag = 0x00000004
	SHF_MERGE            SectionHeaderFlag = 0x00000010
	SHF_STRINGS          SectionHeaderFlag = 0x00000020
	SHF_INFO_LINK        SectionHeaderFlag = 0x00000040
	SHF_LINK_ORDER       SectionHeaderFlag = 0x00000080
	SHF_OS_NONCONFORMING SectionHeaderFlag = 0x00000100
	SHF_GROUP            SectionHeaderFlag = 0x00000200
	SHF_TLS              SectionHeaderFlag = 0x00000400
	SHF_GNU_RETAIN       SectionHeaderFlag = 0x00200000
	SHF_EXCLUDE          SectionHeaderFlag = 0x80000000
)

// Symbol table type
type SymbolType int

const (
	STT_NOTYPE  SymbolType = 0
	STT_OBJECT  SymbolType = 1
	STT_FUNC    SymbolType = 2
	STT_SECTION SymbolType = 3
	STT_FILE    SymbolType = 4
	STT_COMMON  SymbolType = 5
)

type SymbolBinding int

const (
	STB_LOCAL  SymbolBinding = 0
	STB_GLOBAL SymbolBinding = 1
	STB_WEAK   SymbolBinding = 2
)

const (
	GRP_COMDAT = 1
)

type ProgramHeaderType uint32

const (
	PT_NULL    ProgramHeaderType = 0
	PT_LOAD    ProgramHeaderType = 1
	PT_DYNAMIC ProgramHeaderType = 2
	PT_INTERP  ProgramHeaderType = 3
	PT_NOTE    ProgramHeaderType = 4
	PT_SHLIB   ProgramHeaderType = 5
	PT_PHDR    ProgramHeaderType = 6
	PT_TLS     ProgramHeaderType = 7
)

type ProgramHeaderFlag uint32

type R_386 int

const (
	R_386_NONE        R_386 = 0
	R_386_32          R_386 = 1
	R_386_16          R_386 = 20
	R_386_PC16        R_386 = 21
	R_386_SEG16       R_386 = 45
	R_386_SUB16       R_386 = 46
	R_386_SUB32       R_386 = 47
	R_386_SEGRELATIVE R_386 = 48
	R_386_OZSEG16     R_386 = 80
	R_386_OZRELSEG16  R_386 = 81
)

type R_ARM int

const (
	R_ARM_NONE R_ARM = 0
)
