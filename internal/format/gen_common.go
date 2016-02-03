// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import "unicode/utf8"

// A numberSystem identifies a CLDR numbering system.
type numberSystem byte

type numberSystemData struct {
	id        numberSystem
	digitSize byte              // number of UTF-8 bytes per digit
	zero      [utf8.UTFMax]byte // UTF-8 sequence of zero digit.
}

type symbolType int

const (
	symDecimal symbolType = iota
	symGroup
	symList
	symPercentSign
	symPlusSign
	symMinusSign
	symExponential
	symSuperscriptingExponent
	symPerMille
	symInfinity
	symNan
	symTimeSeparator

	numSymbolTypes
)

type altSymData struct {
	compactTag   uint16
	numberSystem numberSystem
	symIndex     byte
}
