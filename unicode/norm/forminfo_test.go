// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build test

package norm

import "testing"

func TestProperties(t *testing.T) {
	var d runeData
	CK := [2]string{"C", "K"}
	for k, r := 1, rune(0); r < 0x2ffff; r++ {
		if k < len(testData) && r == testData[k].r {
			d = testData[k]
			k++
		}
		s := string(r)
		for j, p := range []Properties{NFC.PropertiesString(s), NFKC.PropertiesString(s)} {
			f := d.f[j]
			// Skip Hangul as it is algorithmically computed.
			if r >= hangulBase && r <= hangulEnd {
				// TODO: support Hangul in the API as well.
				continue
			}
			if p.isYesC() != (f.qc == Yes) {
				t.Errorf("%U: YesC(%s): was %v; want %v", r, CK[j], p.isYesC(), f.qc == Yes)
			}
			if p.combinesBackward() != (f.qc == Maybe) {
				t.Errorf("%U: combines backwards(%s): was %v; want %v", r, CK[j], p.combinesBackward(), f.qc == Maybe)
			}
			if p.combinesForward() != f.combinesForward {
				t.Errorf("%U: combines forward(%s): was %v; want %v", r, CK[j], p.combinesForward(), f.combinesForward)
			}
			if p.hasDecomposition() {
				if has := f.decomposition != ""; !has {
					t.Errorf("%U: hasDecomposition(%s): was %v; want %v", r, CK[j], p.hasDecomposition(), has)
				}
				if string(p.Decomposition()) != f.decomposition {
					t.Errorf("%U: decomp(%s): was %+q; want %+q", r, CK[j], p.Decomposition(), f.decomposition)
				}
			}
		}
	}
}
