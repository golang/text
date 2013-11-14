// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package norm

import (
	"unicode/utf8"

	"code.google.com/p/go.text/transform"
)

// Transform implements the transform.Transformer interface. It may need to
// write segments of up to MaxSegmentSize at once. Users should either catch
// ErrShortDst and allow dst to grow or have dst be at least of size
// MaxSegmentSize to be guaranteed of progress.
func (f Form) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	n := 0
	// Cap the maximum number of src bytes to check.
	b := src
	eof := atEOF
	if ns := len(dst); ns < len(b) {
		err = transform.ErrShortDst
		eof = false
		b = b[:ns]
	}
	i, ok := formTable[f].quickSpan(inputBytes(b), n, len(b), eof)
	n += copy(dst[n:], b[n:i])
	if !ok {
		nDst, nSrc, err = f.transform(dst[n:], src[n:], atEOF)
		return nDst + n, nSrc + n, err
	}
	if n < len(src) && !atEOF {
		err = transform.ErrShortSrc
	}
	return n, n, err
}

// transform implements the transform.Transformer interface. It is only called
// when quickSpan does not pass for a given string.
func (f Form) transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	// TODO: get rid of reorderBuffer. See CL 23460044.
	rb := reorderBuffer{}
	rb.init(f, src)
	for {
		// Load segment into reorder buffer.
		end := decomposeSegment(&rb, nSrc)
		if end == rb.nsrc && !atEOF {
			return nDst, nSrc, transform.ErrShortSrc
		}
		if rb.f.composing {
			rb.compose()
		}

		// Write out (must fully fit in dst, or else it is a ErrShortDst).
		if len(dst[nDst:]) < rb.nrune*utf8.UTFMax {
			return nDst, nSrc, transform.ErrShortDst
		}
		nSrc = end
		nDst += rb.flushCopy(dst[nDst:])

		// Next quickSpan.
		end = rb.nsrc
		eof := atEOF
		if n := nSrc + len(dst) - nDst; n < end {
			err = transform.ErrShortDst
			end = n
			eof = false
		}
		end, ok := rb.f.quickSpan(rb.src, nSrc, end, eof)
		n := copy(dst[nDst:], rb.src.bytes[nSrc:end])
		nSrc += n
		nDst += n
		if ok {
			if n < rb.nsrc && !atEOF {
				err = transform.ErrShortSrc
			}
			return nDst, nSrc, err
		}
	}
}
