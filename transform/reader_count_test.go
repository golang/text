// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transform

import (
	"errors"
	"strings"
	"testing"
	"testing/iotest"
)

type invalidCountTransformer struct {
	NopResetter
	nDst, nSrc int
	err        error
	calls      int
}

func (t *invalidCountTransformer) Transform(dst, src []byte, atEOF bool) (int, int, error) {
	t.calls++
	return t.nDst, t.nSrc, t.err
}

func TestReaderInvalidCounts(t *testing.T) {
	for _, counts := range []struct {
		name       string
		nDst, nSrc int
	}{
		{"source overflow", 0, 4},
		{"negative source", 0, -1},
		{"destination overflow", 5, 0},
		{"negative destination", -1, 0},
	} {
		t.Run(counts.name, func(t *testing.T) {
			for _, transformErr := range []error{nil, ErrShortDst, ErrShortSrc} {
				tr := &invalidCountTransformer{nDst: counts.nDst, nSrc: counts.nSrc, err: transformErr}
				r := NewReader(strings.NewReader("abc"), tr)
				r.dst = make([]byte, 4)
				var buf [8]byte
				for i := 0; i < 2; i++ {
					if n, err := r.Read(buf[:]); n != 0 || err != errInconsistentByteCount {
						t.Errorf("Read with Transform error %v = %d, %v; want 0, %v", transformErr, n, err, errInconsistentByteCount)
					}
				}
				if tr.calls != 1 {
					t.Errorf("Transform called %d times, want 1", tr.calls)
				}
			}
		})
	}
}

func TestReaderInvalidCountsReadError(t *testing.T) {
	readErr := errors.New("read error")
	r := NewReader(iotest.ErrReader(readErr), &invalidCountTransformer{nSrc: 1, err: ErrShortDst})
	var buf [1]byte
	if n, err := r.Read(buf[:]); n != 0 || err != readErr {
		t.Errorf("Read = %d, %v; want 0, %v", n, err, readErr)
	}
}
