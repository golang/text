// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package norm

import (
	"testing"

	"code.google.com/p/go.text/transform"
)

func TestTransform(t *testing.T) {
	tests := []struct {
		f       Form
		in, out string
		eof     bool
		dstSize int
		err     error
	}{
		{NFC, "ab", "ab", true, 2, nil},
		{NFC, "qx", "qx", true, 2, nil},
		{NFD, "qx", "qx", true, 2, nil},
		{NFC, "", "", true, 1, nil},
		{NFD, "", "", true, 1, nil},
		{NFC, "", "", false, 1, nil},
		{NFD, "", "", false, 1, nil},

		// Normalized segment does not fit in destination.
		{NFD, "ö", "", true, 1, transform.ErrShortDst},
		{NFD, "ö", "", true, 2, transform.ErrShortDst},

		// As an artifact of the algorithm, only full segments are written.
		// This is not strictly required, and some bytes could be written.
		// In practice, for Transform to not block, the destination buffer
		// should be at least MaxSegmentSize to work anyway and these edge
		// conditions will be relatively rare.
		{NFC, "ab", "", true, 1, transform.ErrShortDst},
		// This is even true for inert runes.
		{NFC, "qx", "", true, 1, transform.ErrShortDst},
		{NFC, "a\u0300abc", "\u00e0a", true, 4, transform.ErrShortDst},

		// We cannot write a segment if succesive runes could still change the result.
		{NFD, "ö", "", false, 3, transform.ErrShortSrc},
		{NFC, "a\u0300", "", false, 4, transform.ErrShortSrc},
		{NFD, "a\u0300", "", false, 4, transform.ErrShortSrc},
		{NFC, "ö", "", false, 3, transform.ErrShortSrc},

		{NFC, "a\u0300", "", true, 1, transform.ErrShortDst},
		// Theoretically could fit, but won't due to simplified checks.
		{NFC, "a\u0300", "", true, 2, transform.ErrShortDst},
		{NFC, "a\u0300", "", true, 3, transform.ErrShortDst},
		{NFC, "a\u0300", "\u00e0", true, 4, nil},

		{NFD, "öa\u0300", "o\u0308", false, 8, transform.ErrShortSrc},
		{NFD, "öa\u0300ö", "o\u0308a\u0300", true, 8, transform.ErrShortDst},
		{NFD, "öa\u0300ö", "o\u0308a\u0300", false, 12, transform.ErrShortSrc},

		// Illegal input is copied verbatim.
		{NFD, "\xbd\xb2=\xbc ", "\xbd\xb2=\xbc ", true, 8, nil},
	}
	b := make([]byte, 100)
	for i, tt := range tests {
		nDst, _, err := tt.f.Transform(b[:tt.dstSize], []byte(tt.in), tt.eof)
		out := string(b[:nDst])
		if out != tt.out || err != tt.err {
			t.Errorf("%d: was %+q (%v); want %+q (%v)", i, out, err, tt.out, tt.err)
		}
		if want := tt.f.String(tt.in)[:nDst]; want != out {
			t.Errorf("%d: incorect normalization: was %+q; want %+q", i, out, want)
		}
	}
}

var transBufSizes = []int{
	MaxTransformChunkSize,
	3 * MaxTransformChunkSize / 2,
	2 * MaxTransformChunkSize,
	3 * MaxTransformChunkSize,
	100 * MaxTransformChunkSize,
}

func doTransNorm(f Form, buf []byte, s string) []byte {
	acc := []byte{}
	b := []byte(s)
	for p := 0; p < len(b); {
		nd, ns, _ := f.Transform(buf[:], b[p:], true)
		p += ns
		acc = append(acc, buf[:nd]...)
	}
	return acc
}

func runTransformTests(t *testing.T, name string, f Form, tests []AppendTest, norm bool) {
	for i, test := range tests {
		in := test.left + test.right
		gold := test.out
		if norm {
			gold = string(f.AppendString(nil, test.out))
		}
		for _, sz := range transBufSizes {
			buf := make([]byte, sz)
			out := string(doTransNorm(f, buf, in))
			if len(out) != len(gold) {
				const msg = "%s:%d:%d: length is %d; want %d"
				t.Errorf(msg, name, i, sz, len(out), len(gold))
			}
			if out != gold {
				k, pf := pidx(out, gold)
				t.Errorf("%s:%d: \nwas  %s%+q; \nwant %s%+q", name, i, pf, pc(out[k:]), pf, pc(gold[k:]))
			}
		}
	}
}

func TestTransformD(t *testing.T) {
	runTransformTests(t, "TransformD1", NFKD, appendTests, true)
	runTransformTests(t, "TransformD2", NFKD, iterTests, true)
}

func TestTransformC(t *testing.T) {
	runTransformTests(t, "TransformC1", NFKC, appendTests, true)
	runTransformTests(t, "TransformC2", NFKC, iterTests, true)
}
