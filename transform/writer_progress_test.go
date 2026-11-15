// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transform

import (
	"bytes"
	"testing"
)

// shortSrcProgress copies a little input while requesting additional input.
type shortSrcProgress struct{ NopResetter }

func (shortSrcProgress) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	if !atEOF && len(src) > 5 {
		n := copy(dst, src[:5])
		if n < 5 {
			return n, n, ErrShortDst
		}
		return n, n, ErrShortSrc
	}
	n := copy(dst, src)
	if n < len(src) {
		return n, n, ErrShortDst
	}
	return n, n, nil
}

func TestWriterShortSrcProgress(t *testing.T) {
	for _, test := range []struct {
		name           string
		bufSize, first int
		total          int
	}{
		{"default buffer", defaultBufSize, 2058, 4110},
		{"large write", defaultBufSize, 2058, 3*defaultBufSize + 41},
		{"small buffer", 7, 6, 40},
	} {
		t.Run(test.name, func(t *testing.T) {
			var dst bytes.Buffer
			w := NewWriter(&dst, shortSrcProgress{})
			w.src = make([]byte, test.bufSize)
			want := make([]byte, test.total)
			for i := range want {
				want[i] = byte(i)
			}
			for _, data := range [][]byte{want[:test.first], want[test.first:]} {
				if n, err := w.Write(data); n != len(data) || err != nil {
					t.Errorf("Write(%d bytes) = %d, %v", len(data), n, err)
				}
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(dst.Bytes(), want) {
				t.Errorf("output differs: got %d bytes, want %d", dst.Len(), len(want))
			}
		})
	}
}
