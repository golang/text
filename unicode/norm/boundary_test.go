// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package norm

import "testing"

func TestCompatibilityJamoBoundaryBefore(t *testing.T) {
	for _, pair := range []struct {
		left, right string
	}{
		{"\u3131", "\u314f"}, // Compatibility leading consonant and vowel.
		{"\uac00", "\u3133"}, // Syllable and compatibility trailing consonant.
		{"\uffa1", "\uffc2"}, // Halfwidth leading consonant and vowel.
	} {
		if NFKC.String(pair.left+pair.right) == NFKC.String(pair.left)+NFKC.String(pair.right) {
			t.Fatalf("test pair %+q, %+q does not combine", pair.left, pair.right)
		}
		for _, f := range []Form{NFKC, NFKD} {
			if f.PropertiesString(pair.right).BoundaryBefore() {
				t.Errorf("form %v: PropertiesString(%+q).BoundaryBefore() = true", f, pair.right)
			}
			if f.Properties([]byte(pair.right)).BoundaryBefore() {
				t.Errorf("form %v: Properties(%+q).BoundaryBefore() = true", f, pair.right)
			}
		}
	}
}
