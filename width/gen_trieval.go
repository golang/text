// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

// elem is an entry of the width trie.
type elem uint16

const (
	hasMappingMask elem = 0x8000
	tagMappingMask      = 0xc000
	tagHalfwidth        = 0xc000
	tagFullwidth        = 0x8000
	tagNeutral          = 0x0000
	tagAmbiguous        = 0x0001
	tagWide             = 0x0002
	tagNarrow           = 0x0003

	// The Korean Won sign is halfwidth, but SHOULD NOT be mapped to a wide
	// variant.
	wonSign = 0x20A9
)
