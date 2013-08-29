// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package japanese

import (
	"errors"
	"unicode/utf8"

	"code.google.com/p/go.text/encoding"
	"code.google.com/p/go.text/transform"
)

// EUCJP is the EUC-JP (Extended Unix Code Japanese) encoding.
var EUCJP encoding.Encoding = eucJP{}

type eucJP struct{}

func (eucJP) NewDecoder() transform.Transformer {
	return eucJPDecoder{}
}

func (eucJP) NewEncoder() transform.Transformer {
	return eucJPEncoder{}
}

func (eucJP) String() string {
	return "EUC-JP"
}

var errInvalidEUCJP = errors.New("japanese: invalid EUC-JP encoding")

type eucJPDecoder struct{}

func (eucJPDecoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	r, size := rune(0), 0
loop:
	for ; nSrc < len(src); nSrc += size {
		switch c0 := src[nSrc]; {
		case c0 < utf8.RuneSelf:
			r, size = rune(c0), 1

		case c0 == 0x8e:
			if nSrc+1 >= len(src) {
				err = transform.ErrShortSrc
				break loop
			}
			c1 := src[nSrc+1]
			if c1 < 0xa1 || 0xdf < c1 {
				err = errInvalidEUCJP
				break loop
			}
			r, size = rune(c1)+(0xff61-0xa1), 2

		case c0 == 0x8f:
			if nSrc+2 >= len(src) {
				err = transform.ErrShortSrc
				break loop
			}
			c1 := src[nSrc+1]
			if c1 < 0xa1 || 0xfe < c1 {
				err = errInvalidEUCJP
				break loop
			}
			c2 := src[nSrc+2]
			if c2 < 0xa1 || 0xfe < c2 {
				err = errInvalidEUCJP
				break loop
			}
			r, size = encoding.ASCIISub, 3
			if i := int(c1-0xa1)*94 + int(c2-0xa1); i < len(jis0212Decode) {
				r = rune(jis0212Decode[i])
				if r == 0 {
					r = encoding.ASCIISub
				}
			}

		case 0xa1 <= c0 && c0 <= 0xfe:
			if nSrc+1 >= len(src) {
				err = transform.ErrShortSrc
				break loop
			}
			c1 := src[nSrc+1]
			if c1 < 0xa1 || 0xfe < c1 {
				err = errInvalidEUCJP
				break loop
			}
			r, size = encoding.ASCIISub, 2
			if i := int(c0-0xa1)*94 + int(c1-0xa1); i < len(jis0208Decode) {
				r = rune(jis0208Decode[i])
				if r == 0 {
					r = encoding.ASCIISub
				}
			}

		default:
			err = errInvalidEUCJP
			break loop
		}

		if nDst+utf8.RuneLen(r) >= len(dst) {
			err = transform.ErrShortDst
			break loop
		}
		nDst += utf8.EncodeRune(dst[nDst:], r)
	}
	if atEOF && err == transform.ErrShortSrc {
		err = errInvalidEUCJP
	}
	return nDst, nSrc, err
}

type eucJPEncoder struct{}

func (eucJPEncoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	r, size := rune(0), 0
loop:
	for ; nSrc < len(src); nSrc += size {
		r = rune(src[nSrc])

		// Decode a 1-byte rune.
		if r < utf8.RuneSelf {
			size = 1

		} else {
			// Decode a multi-byte rune.
			r, size = utf8.DecodeRune(src[nSrc:])
			if size == 1 {
				// All valid runes of size 1 (those below utf8.RuneSelf) were
				// handled above. We have invalid UTF-8 or we haven't seen the
				// full character yet.
				if !atEOF && !utf8.FullRune(src[nSrc:]) {
					err = transform.ErrShortSrc
					break loop
				}
			}
		}

		switch {
		case r < utf8.RuneSelf:
			// No-op.

		case 0xff61 <= r && r <= 0xff9f:
			if nDst+2 > len(dst) {
				err = transform.ErrShortDst
				break loop
			}
			dst[nDst+0] = 0x8e
			dst[nDst+1] = uint8(r - (0xff61 - 0xa1))
			nDst += 2
			continue loop

		case 0xffff < r:
			r = encoding.ASCIISub

		default:
			e := jisEncode[uint16(r)]
			if e == 0 {
				r = encoding.ASCIISub
				break
			}
			switch e >> tableShift {
			case jis0208:
				if nDst+2 > len(dst) {
					err = transform.ErrShortDst
					break loop
				}
				dst[nDst+0] = 0xa1 + uint8(e>>codeShift)&codeMask
				dst[nDst+1] = 0xa1 + uint8(e)&codeMask
				nDst += 2
			case jis0212:
				if nDst+3 > len(dst) {
					err = transform.ErrShortDst
					break loop
				}
				dst[nDst+0] = 0x8f
				dst[nDst+1] = 0xa1 + uint8(e>>codeShift)&codeMask
				dst[nDst+2] = 0xa1 + uint8(e)&codeMask
				nDst += 3
			}
			continue loop
		}

		// r is encoded as a single byte.
		if nDst >= len(dst) {
			err = transform.ErrShortDst
			break loop
		}
		dst[nDst] = uint8(r)
		nDst++
	}
	return nDst, nSrc, err
}
