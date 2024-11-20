// SPDX-License-Identifier: MIT
//
// Copyright (c) 2024 Adrian "asie" Siekierka

package relocation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockRegionEntry struct {
	offset uint64
	size   uint64
	align  uint64
}

func (r MockRegionEntry) Offset() uint64 {
	return r.offset
}

func (r *MockRegionEntry) SetOffset(offset uint64) {
	r.offset = offset
}

func (r MockRegionEntry) Size() uint64 {
	return r.size
}

func (r MockRegionEntry) Alignment() uint64 {
	return r.align
}

func NewMockRegionEntry(size uint64, align uint64) *MockRegionEntry {
	return &MockRegionEntry{
		offset: 0,
		size:   size,
		align:  align,
	}
}

func TestAddEntries(t *testing.T) {
	e1 := NewMockRegionEntry(64, 1)
	e2 := NewMockRegionEntry(32, 1)
	r := NewRegion[*MockRegionEntry](0, 1000, false)
	ok, _ := r.Place(e1, nil, false)
	assert.True(t, ok, "first entry placement")
	ok, _ = r.Place(e2, nil, false)
	assert.True(t, ok, "second entry placement")
	assert.Equal(t, uint64(0), e1.Offset(), "first entry offset")
	assert.Equal(t, uint64(64), e2.Offset(), "second entry offset")
}

func TestAddEntriesDescending(t *testing.T) {
	e1 := NewMockRegionEntry(64, 1)
	e2 := NewMockRegionEntry(32, 1)
	r := NewRegion[*MockRegionEntry](0, 1000, true)
	ok, _ := r.Place(e1, nil, false)
	assert.True(t, ok, "first entry placement")
	ok, _ = r.Place(e2, nil, false)
	assert.True(t, ok, "second entry placement")
	assert.Equal(t, uint64(936), e1.Offset(), "first entry offset")
	assert.Equal(t, uint64(904), e2.Offset(), "second entry offset")
}

func TestAddEntriesAlignment(t *testing.T) {
	// e1, e4, e3, e2, e6, e5
	e1 := NewMockRegionEntry(61, 4)
	e2 := NewMockRegionEntry(30, 4)
	e3 := NewMockRegionEntry(1, 2)
	e4 := NewMockRegionEntry(1, 1)
	e5 := NewMockRegionEntry(1, 128)
	e6 := NewMockRegionEntry(1, 16)
	r := NewRegion[*MockRegionEntry](0, 1000, false)
	ok, _ := r.Place(e1, nil, false)
	assert.True(t, ok, "first entry placement")
	ok, _ = r.Place(e2, nil, false)
	assert.True(t, ok, "second entry placement")
	ok, _ = r.Place(e3, nil, false)
	assert.True(t, ok, "third entry placement")
	ok, _ = r.Place(e4, nil, false)
	assert.True(t, ok, "fourth entry placement")
	ok, _ = r.Place(e5, nil, false)
	assert.True(t, ok, "fifth entry placement")
	ok, _ = r.Place(e6, nil, false)
	assert.True(t, ok, "sixth entry placement")
	assert.Equal(t, uint64(0), e1.Offset(), "first entry offset")
	assert.Equal(t, uint64(64), e2.Offset(), "second entry offset")
	assert.Equal(t, uint64(62), e3.Offset(), "third entry offset")
	assert.Equal(t, uint64(61), e4.Offset(), "fourth entry offset")
	assert.Equal(t, uint64(128), e5.Offset(), "fifth entry offset")
	assert.Equal(t, uint64(96), e6.Offset(), "sixth entry offset")
}
