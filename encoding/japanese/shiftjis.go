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

// ShiftJIS is the Shift JIS (Japanese Industrial Standards) encoding, also
// known as Code Page 932 and Windows-31J.
var ShiftJIS encoding.Encoding = shiftJIS{}

type shiftJIS struct{}

func (shiftJIS) NewDecoder() transform.Transformer {
	return shiftJISDecoder{}
}

func (shiftJIS) NewEncoder() transform.Transformer {
	return shiftJISEncoder{}
}

func (shiftJIS) String() string {
	return "Shift JIS"
}

var errInvalidShiftJIS = errors.New("japanese: invalid Shift JIS encoding")

type shiftJISDecoder struct{}

func (shiftJISDecoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	r, size := rune(0), 0
loop:
	for ; nSrc < len(src); nSrc += size {
		switch c0 := src[nSrc]; {
		case c0 < utf8.RuneSelf:
			r, size = rune(c0), 1

		case 0xa1 <= c0 && c0 <= 0xdf:
			r, size = rune(c0)+(0xff61-0xa1), 1

		case (0x81 <= c0 && c0 <= 0x9f) || (0xe0 <= c0 && c0 <= 0xef):
			if c0 <= 0x9f {
				c0 -= 0x70
			} else {
				c0 -= 0xb0
			}
			c0 = 2*c0 - 0x21

			if nSrc+1 >= len(src) {
				err = transform.ErrShortSrc
				break loop
			}
			c1 := src[nSrc+1]
			switch {
			case c1 < 0x40:
				err = errInvalidShiftJIS
				break loop
			case c1 < 0x7f:
				c0--
				c1 -= 0x40
			case c1 == 0x7f:
				err = errInvalidShiftJIS
				break loop
			case c1 < 0x9f:
				c0--
				c1 -= 0x41
			case c1 < 0xfd:
				c1 -= 0x9f
			default:
				err = errInvalidShiftJIS
				break loop
			}
			r, size = encoding.ASCIISub, 2
			if i := int(c0)*94 + int(c1); i < len(jis0208Decode) {
				r = rune(jis0208Decode[i])
				if r == 0 {
					r = encoding.ASCIISub
				}
			}

		default:
			err = errInvalidShiftJIS
			break loop
		}

		if nDst+utf8.RuneLen(r) > len(dst) {
			err = transform.ErrShortDst
			break loop
		}
		nDst += utf8.EncodeRune(dst[nDst:], r)
	}
	if atEOF && err == transform.ErrShortSrc {
		err = errInvalidShiftJIS
	}
	return nDst, nSrc, err
}

type shiftJISEncoder struct{}

func (shiftJISEncoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
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
			// r is an ASCII rune.

		case 0xff61 <= r && r <= 0xff9f:
			r -= 0xff61 - 0xa1

		case 0xffff < r:
			r = encoding.ASCIISub

		default:
			e := jisEncode[uint16(r)]
			if e == 0 {
				r = encoding.ASCIISub
				break
			}
			if e>>tableShift != jis0208 {
				r = encoding.ASCIISub
				break
			}
			j1 := uint8(e>>codeShift) & codeMask
			j2 := uint8(e) & codeMask
			if nDst+2 > len(dst) {
				err = transform.ErrShortDst
				break loop
			}
			if j1 <= 61 {
				dst[nDst+0] = 129 + j1/2
			} else {
				dst[nDst+0] = 193 + j1/2
			}
			if j1&1 == 0 {
				dst[nDst+1] = j2 + j2/63 + 64
			} else {
				dst[nDst+1] = j2 + 159
			}
			nDst += 2
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
