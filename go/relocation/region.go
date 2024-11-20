// SPDX-License-Identifier: MIT
//
// Copyright (c) 2023, 2024 Adrian "asie" Siekierka

package relocation

import (
	"slices"
)

type RegionPlaceable interface {
	Offset() uint64
	SetOffset(uint64)
	Size() uint64
	Alignment() uint64
}

type Region[T RegionPlaceable] struct {
	offset     uint64
	size       uint64
	entries    []T
	descending bool
}

type RegionList[T RegionPlaceable] struct {
	Regions []*Region[T]
}

func NewRegion[T RegionPlaceable](offset uint64, size uint64, descending bool) *Region[T] {
	r := Region[T]{
		offset:     offset,
		size:       size,
		entries:    make([]T, 0),
		descending: descending,
	}
	return &r
}

func (r Region[T]) Offset() uint64 {
	return r.offset
}

func (r Region[T]) Size() uint64 {
	return r.size
}

func (r Region[T]) Empty() bool {
	return len(r.entries) == 0
}

func (r Region[T]) Full() bool {
	if r.descending {
		return r.UsedStart() == r.offset
	} else {
		return r.UsedEnd() == r.offset+r.size-1
	}
}

func (r Region[T]) UsedStart() uint64 {
	if !r.Empty() {
		first := r.entries[0]
		return first.Offset()
	} else {
		return r.offset
	}
}

func (r Region[T]) UsedEnd() uint64 {
	if !r.Empty() {
		last := r.entries[len(r.entries)-1]
		return last.Offset() + last.Size() - 1
	} else {
		// if empty, UsedEnd() == UsedStart()
		return r.UsedStart()
	}
}

func calcEntryOffset(start uint64, end uint64, len uint64, descending bool, align uint64) (bool, uint64) {
	if descending {
		offset := end - len
		if align > 1 {
			offset -= (offset % align)
		}
		if offset >= start {
			return true, offset
		}
	} else {
		offset := start
		if align > 1 {
			offset += align - 1
			offset -= (offset % align)
		}
		if (offset + len) <= end {
			return true, offset
		}
	}

	return false, 0
}

const (
	RegionFindGapModeSmallest = iota
	RegionFindGapModeLargest
	RegionFindGapModeFirst
)

/*
type stringer interface {
	String() string
}

func name(r RegionPlaceable) string {
	if s, ok := r.(stringer); ok {
		return s.String()
	}
	return "unknown"
}
*/

func (r Region[T]) findGap(offsetMin uint64, offsetMax uint64, mode int, minimumSize int64, startIndex int) (bool, uint64, uint64, int) {
	if r.Empty() {
		if int64(offsetMax-offsetMin) >= minimumSize {
			return true, offsetMin, offsetMax, 0
		} else {
			return false, 0, 0, 0
		}
	}

	// TODO: skip up to startIndex
	if startIndex > 0 && mode != RegionFindGapModeFirst {
		panic("startIndex not supported with non-first gap find mode yet")
	}

	previous := r.entries[0]
	gapStart := max(offsetMin, r.Offset())
	gapSize := int64(previous.Offset() - offsetMin)
	gapIndex := 0

	newBestGap := func(newGapSize int64) bool {
		if newGapSize < minimumSize {
			return false
		}
		if mode == RegionFindGapModeFirst {
			return true
		} else if mode == RegionFindGapModeLargest {
			return newGapSize > gapSize
		} else {
			return newGapSize < gapSize
		}
	}

	if mode == RegionFindGapModeFirst && gapSize >= minimumSize && gapIndex >= startIndex {
		return true, gapStart, gapStart + uint64(gapSize), gapIndex
	}

	for i := 1; i <= len(r.entries); i++ {
		currentGapStart := max(offsetMin, previous.Offset()+previous.Size())
		currentGapEnd := min(offsetMax, r.Offset()+r.Size())

		if i < len(r.entries) {
			current := r.entries[i]
			currentGapEnd = min(currentGapEnd, current.Offset())
			previous = current
		}

		currentGap := int64(currentGapEnd - currentGapStart)
		if newBestGap(currentGap) {
			gapStart = currentGapStart
			gapSize = currentGap
			gapIndex = i

			if mode == RegionFindGapModeFirst && gapIndex >= startIndex {
				return true, gapStart, gapStart + uint64(gapSize), gapIndex
			}
		}
	}

	if gapSize < minimumSize {
		return false, 0, 0, 0
	} else {
		return true, gapStart, gapStart + uint64(gapSize), gapIndex
	}
}

func (r Region[T]) FindGap(offsetMin uint64, offsetMax uint64, mode int, minimumSize int64) (bool, uint64, uint64) {
	ok, gapStart, gapEnd, _ := r.findGap(offsetMin, offsetMax, mode, minimumSize, 0)
	return ok, gapStart, gapEnd
}

func (r Region[T]) FindAnyGap(mode int, minimumSize int64) (bool, uint64, uint64) {
	return r.FindGap(r.Offset(), r.Offset()+r.Size(), mode, minimumSize)
}

func (r *Region[T]) Place(entry T, offsetRange []uint64, simulate bool) (bool, uint64) {
	offsetMin := r.Offset()
	offsetMax := r.Offset() + r.Size()

	if offsetRange != nil {
		if len(offsetRange) == 2 {
			offsetMin = max(offsetMin, offsetRange[0])
			offsetMax = min(offsetMax, offsetRange[1]+1)
		} else if len(offsetRange) == 1 {
			offsetMin = max(offsetMin, offsetRange[0])
			offsetMax = min(offsetMax, offsetRange[0]+entry.Size())
		} else {
			panic("Unsupported offsetRange length")
		}
	}

	if r.Size() >= (offsetMax - offsetMin) {
		ok := true
		var gapStart uint64
		var gapEnd uint64
		gapIndex := -1
		for ok {
			ok, gapStart, gapEnd, gapIndex = r.findGap(offsetMin, offsetMax, RegionFindGapModeFirst, int64(entry.Size()), gapIndex+1)
			if !ok {
				return false, 0
			}

			ok, offset := calcEntryOffset(gapStart, gapEnd, entry.Size(), r.descending, entry.Alignment())
			if ok {
				if !simulate {
					entry.SetOffset(offset)
					r.entries = slices.Insert(r.entries, gapIndex, entry)

				}
				return true, r.offset
			}

			// loop starting from next index; alignment might have not been sufficient
		}
	}

	return false, 0
}

func (r *RegionList[T]) Place(entry T, offsetRange []uint64, simulate bool) (bool, uint64) {
	for i := 0; i < len(r.Regions); i++ {
		ok, offset := r.Place(entry, offsetRange, simulate)
		if ok {
			return ok, offset
		}
	}
	return false, 0
}
