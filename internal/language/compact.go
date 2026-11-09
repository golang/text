// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package language

// CompactCoreInfo packs a 14-bit language, 9-bit script, and 9-bit region.
type CompactCoreInfo uint32

// GetCompactCore generates a uint32 value that is guaranteed to be unique for
// different language, region, and script values.
func GetCompactCore(t Tag) (cci CompactCoreInfo, ok bool) {
	if t.LangID >= langNoIndexOffset || t.LangID >= 1<<14 ||
		t.ScriptID >= 1<<9 || t.RegionID >= 1<<9 {
		return 0, false
	}
	cci |= CompactCoreInfo(t.LangID) << 18
	cci |= CompactCoreInfo(t.ScriptID) << 9
	cci |= CompactCoreInfo(t.RegionID)
	return cci, true
}

// Tag generates a tag from c.
func (c CompactCoreInfo) Tag() Tag {
	return Tag{
		LangID:   Language(c >> 18),
		RegionID: Region(c & 0x1ff),
		ScriptID: Script(c>>9) & 0x1ff,
	}
}
