// SPDX-License-Identifier: MIT
//
// Copyright (c) 2024 Adrian "asie" Siekierka

package elf

func (e *Elf) GarbageCollectSections(retainedSections map[*SectionHeader]bool) {
	// Build a tree of relations where child section is dependent on a parent section
	sectionChildren := make(map[*SectionHeader]map[*SectionHeader]bool)

	for _, section := range e.Sections {
		if (section.Flags & SHF_GNU_RETAIN) != 0 {
			retainedSections[section] = true
		}
	}

	for parentSection, relocations := range e.Relocations {
		for _, relocation := range relocations {
			if relocation.Symbol != nil && relocation.Symbol.Section != nil {
				childSection := relocation.Symbol.Section
				children, ok := sectionChildren[parentSection]
				if !ok {
					children = make(map[*SectionHeader]bool)
					sectionChildren[parentSection] = children
				}
				children[childSection] = true
			}
		}
	}

	// Traverse the tree of parent<->child relations
	newlyRetainedSections := retainedSections
	retainedSections = make(map[*SectionHeader]bool)

	for len(newlyRetainedSections) > 0 {
		nextRetainedSections := make(map[*SectionHeader]bool)
		for retainedSection := range newlyRetainedSections {
			retainedSections[retainedSection] = true

			if sectionChildren[retainedSection] != nil {
				for childOfRetainedSection := range sectionChildren[retainedSection] {
					if _, ok := retainedSections[childOfRetainedSection]; !ok {
						nextRetainedSections[childOfRetainedSection] = true
					}
				}
			}
		}

		newlyRetainedSections = nextRetainedSections
	}

	allocatedSections := make([]*SectionHeader, 0)
	for _, section := range e.Sections {
		if _, ok := retainedSections[section]; ok {
			allocatedSections = append(allocatedSections, section)
		} else {
			// TODO
			// log.info("gc: removing section " .. v.name)
		}
	}
	e.Sections = allocatedSections
}
