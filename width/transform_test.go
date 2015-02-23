// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package width

import (
	"bytes"
	"testing"

	"golang.org/x/text/internal/testtext"
	"golang.org/x/text/transform"
)

func TestFoldSingleRunes(t *testing.T) {
	for r := rune(0); r < 0x1FFFF; r++ {
		if loSurrogate <= r && r <= hiSurrogate {
			continue
		}
		want := string(r)
		if alt, ok := foldRunes[r]; ok {
			want = string(alt)
		}
		got := Fold().String(string(r))
		if got != want {
			t.Errorf("Fold().String(%U) = %+q; want %+q", r, got, want)
		}
	}
}

func TestFold(t *testing.T) {
	var fold Transformer
	if n := testing.AllocsPerRun(1, func() { fold = Fold() }); n > 0 {
		t.Errorf("#allocs was %f; want 0", n)
	}
	for _, tc := range []struct {
		desc  string
		src   string
		nDst  int
		atEOF bool
		dst   string
		nSrc  int
		err   error
	}{{
		desc:  "empty",
		src:   "",
		dst:   "",
		nDst:  10,
		nSrc:  0,
		atEOF: false,
		err:   nil,
	}, {
		desc:  "short source 1",
		src:   "a\xc0",
		dst:   "a",
		nDst:  10,
		nSrc:  1,
		atEOF: false,
		err:   transform.ErrShortSrc,
	}, {
		desc:  "short source 2",
		src:   "a\xe0\x80",
		dst:   "a",
		nDst:  10,
		nSrc:  1,
		atEOF: false,
		err:   transform.ErrShortSrc,
	}, {
		desc:  "incomplete but terminated source 1",
		src:   "a\xc0",
		dst:   "a\xc0",
		nDst:  10,
		nSrc:  2,
		atEOF: true,
		err:   nil,
	}, {
		desc:  "incomplete but terminated source 2",
		src:   "a\xe0\x80",
		dst:   "a\xe0\x80",
		nDst:  10,
		nSrc:  3,
		atEOF: true,
		err:   nil,
	}, {
		desc:  "exact fit dst",
		src:   "a\uff01",
		dst:   "a!",
		nDst:  2,
		nSrc:  4,
		atEOF: false,
		err:   nil,
	}, {
		desc:  "short dst 1",
		src:   "a\uffe0",
		dst:   "a",
		nDst:  2,
		nSrc:  1,
		atEOF: false,
		err:   transform.ErrShortDst,
	}} {
		b := make([]byte, tc.nDst)
		nDst, nSrc, err := fold.Transform(b, []byte(tc.src), tc.atEOF)
		if got := string(b[:nDst]); got != tc.dst {
			t.Errorf("%s: dst was %+q; want %+q", tc.desc, got, tc.dst)
		}
		if nSrc != tc.nSrc {
			t.Errorf("%s: nSrc was %d; want %d", tc.desc, nSrc, tc.nSrc)
		}
		if err != tc.err {
			t.Errorf("%s: error was %v; want %v", tc.desc, err, tc.err)
		}
	}
}

func benchFold(b *testing.B, s string) {
	b.StopTimer()
	dst := make([]byte, 1024)
	src := []byte(s)
	b.SetBytes(int64(len(src)))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		foldTransform{}.Transform(dst, src, true)
	}
}

func BenchmarkFoldASCII(b *testing.B) {
	benchFold(b, testtext.ASCII)
}

func BenchmarkFoldCJK(b *testing.B) {
	benchFold(b, testtext.CJK)
}

var foldingRunes string

func init() {
	buf := &bytes.Buffer{}
	for r := rune(0xFF00); r <= 0xFFFF; r++ {
		if _, ok := foldRunes[r]; ok {
			buf.WriteRune(r)
		}
	}
	foldingRunes = buf.String()
}

func BenchmarkFoldNonCanonical(b *testing.B) {
	benchFold(b, foldingRunes)
}

func BenchmarkFoldOther(b *testing.B) {
	benchFold(b, testtext.TwoByteUTF8+testtext.ThreeByteUTF8)
}
