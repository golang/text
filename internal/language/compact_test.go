// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package language

import "testing"

func TestCompactCoreRoundTrip(t *testing.T) {
	// Cover every indexed language, including IDs above 4095 with CLDR 49,
	// and boundaries of the script and region bit fields.
	seen := make(map[CompactCoreInfo]Tag)
	for l := 0; l < langNoIndexOffset; l++ {
		for _, s := range []Script{0, 255, 256, 511} {
			for _, r := range []Region{0, 255, 256, 511} {
				tag := Tag{LangID: Language(l), ScriptID: s, RegionID: r}
				core, ok := GetCompactCore(tag)
				if !ok {
					t.Fatalf("indexed language %d rejected", l)
				}
				if got := core.Tag(); got != tag {
					t.Fatalf("round trip: got %#v; want %#v", got, tag)
				}
				if old, exists := seen[core]; exists && old != tag {
					t.Fatalf("collision between %#v and %#v", old, tag)
				}
				seen[core] = tag
			}
		}
	}
}

func TestCompactCoreNonIndexedLanguage(t *testing.T) {
	for _, id := range []Language{Language(langNoIndexOffset), 65535} {
		if _, ok := GetCompactCore(Tag{LangID: id}); ok {
			t.Errorf("non-indexed language %d accepted", id)
		}
	}
}

func TestCompactCoreCapacity(t *testing.T) {
	for _, tag := range []Tag{
		{LangID: 1 << 14},
		{ScriptID: 1 << 9},
		{RegionID: 1 << 9},
	} {
		if _, ok := GetCompactCore(tag); ok {
			t.Errorf("out-of-range tag accepted: %#v", tag)
		}
	}
}
