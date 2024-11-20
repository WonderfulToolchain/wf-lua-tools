// SPDX-License-Identifier: MIT
//
// Copyright (c) 2024 Adrian "asie" Siekierka

package elf

import "io"

func readString(r io.Reader) (error, string) {
	s := ""
	buf := make([]byte, 1)

	for {
		if _, err := r.Read(buf); err != nil {
			return err, s
		}
		if buf[0] == 0 {
			return nil, s
		}
		s = s + string(buf)
	}
}
