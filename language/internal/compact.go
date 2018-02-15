// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package language

// GetCoreKey generates a uint32 value that is guaranteed to be unique for
// different language, region, and script values.
func GetCoreKey(t Tag) (key uint32) {
	if t.LangID > langNoIndexOffset {
		return 0xfff00000
	}
	key |= uint32(t.LangID) << (8 + 12)
	key |= uint32(t.ScriptID) << 12
	key |= uint32(t.RegionID)
	return key
}
