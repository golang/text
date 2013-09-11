// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package korean

import (
	"errors"
	"unicode/utf8"

	"code.google.com/p/go.text/encoding"
	"code.google.com/p/go.text/transform"
)

// EUCKR is the EUC-KR encoding, also known as Code Page 949.
var EUCKR encoding.Encoding = eucKR{}

type eucKR struct{}

func (eucKR) NewDecoder() transform.Transformer {
	return eucKRDecoder{}
}

func (eucKR) NewEncoder() transform.Transformer {
	return eucKREncoder{}
}

func (eucKR) String() string {
	return "EUC-KR"
}

var errInvalidEUCKR = errors.New("korean: invalid EUC-KR encoding")

type eucKRDecoder struct{}

func (eucKRDecoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	r, size := rune(0), 0
loop:
	for ; nSrc < len(src); nSrc += size {
		switch c0 := src[nSrc]; {
		case c0 < utf8.RuneSelf:
			r, size = rune(c0), 1

		case 0x81 <= c0 && c0 < 0xff:
			if nSrc+1 >= len(src) {
				err = transform.ErrShortSrc
				break loop
			}
			c1 := src[nSrc+1]
			if c0 < 0xc7 {
				r = 178 * rune(c0-0x81)
				switch {
				case 0x41 <= c1 && c1 < 0x5b:
					r += rune(c1) - (0x41 - 0*26)
				case 0x61 <= c1 && c1 < 0x7b:
					r += rune(c1) - (0x61 - 1*26)
				case 0x81 <= c1 && c1 < 0xff:
					r += rune(c1) - (0x81 - 2*26)
				default:
					err = errInvalidEUCKR
					break loop
				}
			} else if 0xa1 <= c1 && c1 < 0xff {
				r = 178*(0xc7-0x81) + rune(c0-0xc7)*94 + rune(c1-0xa1)
			} else {
				err = errInvalidEUCKR
				break loop
			}
			if int(r) < len(eucKRDecode) {
				r = rune(eucKRDecode[r])
				if r == 0 {
					r = encoding.ASCIISub
				}
			} else {
				r = encoding.ASCIISub
			}
			size = 2

		default:
			err = errInvalidEUCKR
			break loop
		}

		if nDst+utf8.RuneLen(r) > len(dst) {
			err = transform.ErrShortDst
			break loop
		}
		nDst += utf8.EncodeRune(dst[nDst:], r)
	}
	if atEOF && err == transform.ErrShortSrc {
		err = errInvalidEUCKR
	}
	return nDst, nSrc, err
}

type eucKREncoder struct{}

func (eucKREncoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
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

		case 0xffff < r:
			r = encoding.ASCIISub

		default:
			e := eucKREncode[uint16(r)]
			if e == 0 {
				r = encoding.ASCIISub
				break
			}
			if nDst+2 > len(dst) {
				err = transform.ErrShortDst
				break loop
			}
			dst[nDst+0] = uint8(e >> 8)
			dst[nDst+1] = uint8(e)
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
