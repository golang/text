// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package traditionalchinese

import (
	"errors"
	"unicode/utf8"

	"code.google.com/p/go.text/encoding"
	"code.google.com/p/go.text/transform"
)

// Big5 is the Big5 encoding, also known as Code Page 950.
var Big5 encoding.Encoding = big5{}

type big5 struct{}

func (big5) NewDecoder() transform.Transformer {
	return big5Decoder{}
}

func (big5) NewEncoder() transform.Transformer {
	return big5Encoder{}
}

func (big5) String() string {
	return "Big5"
}

var errInvalidBig5 = errors.New("traditionalchinese: invalid Big5 encoding")

type big5Decoder struct{}

func (big5Decoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	r, size, s := rune(0), 0, ""
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
			switch {
			case 0x40 <= c1 && c1 < 0x7f:
				c1 -= 0x40
			case 0xa1 <= c1 && c1 < 0xff:
				c1 -= 0x62
			default:
				err = errInvalidBig5
				break loop
			}
			r, size = encoding.ASCIISub, 2
			if i := int(c0-0x81)*157 + int(c1); i < len(big5Decode) {
				if 1133 <= i && i < 1167 {
					// The two-rune special cases for LATIN CAPITAL / SMALL E WITH CIRCUMFLEX
					// AND MACRON / CARON are from http://encoding.spec.whatwg.org/#big5
					switch i {
					case 1133:
						s = "\u00CA\u0304"
						goto writeStr
					case 1135:
						s = "\u00CA\u030C"
						goto writeStr
					case 1164:
						s = "\u00EA\u0304"
						goto writeStr
					case 1166:
						s = "\u00EA\u030C"
						goto writeStr
					}
				}
				r = rune(big5Decode[i])
				if r == 0 {
					r = encoding.ASCIISub
				}
			}

		default:
			err = errInvalidBig5
			break loop
		}

		if nDst+utf8.RuneLen(r) > len(dst) {
			err = transform.ErrShortDst
			break loop
		}
		nDst += utf8.EncodeRune(dst[nDst:], r)
		continue loop

	writeStr:
		if nDst+len(s) > len(dst) {
			err = transform.ErrShortDst
			break loop
		}
		nDst += copy(dst[nDst:], s)
		continue loop
	}
	if atEOF && err == transform.ErrShortSrc {
		err = errInvalidBig5
	}
	return nDst, nSrc, err
}

type big5Encoder struct{}

func (big5Encoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	r, size := rune(0), 0
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
					break
				}
			}
		}

		if r >= utf8.RuneSelf {
			switch {
			case big5Encode0Low <= r && r < big5Encode0High:
				if r = rune(big5Encode0[r-big5Encode0Low]); r != 0 {
					goto write2
				}
			case big5Encode1Low <= r && r < big5Encode1High:
				if r = rune(big5Encode1[r-big5Encode1Low]); r != 0 {
					goto write2
				}
			case big5Encode2Low <= r && r < big5Encode2High:
				if r = rune(big5Encode2[r-big5Encode2Low]); r != 0 {
					goto write2
				}
			case big5Encode3Low <= r && r < big5Encode3High:
				if r = rune(big5Encode3[r-big5Encode3Low]); r != 0 {
					goto write2
				}
			case big5Encode4Low <= r && r < big5Encode4High:
				if r = rune(big5Encode4[r-big5Encode4Low]); r != 0 {
					goto write2
				}
			case big5Encode5Low <= r && r < big5Encode5High:
				if r = rune(big5Encode5[r-big5Encode5Low]); r != 0 {
					goto write2
				}
			case big5Encode6Low <= r && r < big5Encode6High:
				if r = rune(big5Encode6[r-big5Encode6Low]); r != 0 {
					goto write2
				}
			case big5Encode7Low <= r && r < big5Encode7High:
				if r = rune(big5Encode7[r-big5Encode7Low]); r != 0 {
					goto write2
				}
			}
			r = encoding.ASCIISub
		}

		if nDst >= len(dst) {
			err = transform.ErrShortDst
			break
		}
		dst[nDst] = uint8(r)
		nDst++
		continue

	write2:
		if nDst+2 > len(dst) {
			err = transform.ErrShortDst
			break
		}
		dst[nDst+0] = uint8(r >> 8)
		dst[nDst+1] = uint8(r)
		nDst += 2
		continue
	}
	return nDst, nSrc, err
}
