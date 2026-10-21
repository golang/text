// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package traditionalchinese

import (
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/internal"
	"golang.org/x/text/encoding/internal/identifier"
	"golang.org/x/text/transform"
)

// AllHK is a list of all defined encodings in this package.
var AllHK = []encoding.Encoding{Big5HK}

// Big5HK is the Big5HK encoding, also known as Code Page 950.
var Big5HK encoding.Encoding = &big5HK

var big5HK = internal.Encoding{
	&internal.SimpleEncoding{big5DecoderHK{}, big5EncoderHK{}},
	"Big5",
	identifier.Big5,
}

type big5DecoderHK struct{ transform.NopResetter }

func (big5DecoderHK) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	r, size, s := rune(0), 0, ""
loop:
	for ; nSrc < len(src); nSrc += size {
		switch c0 := src[nSrc]; {
		case c0 < utf8.RuneSelf:
			r, size = rune(c0), 1

		case 0x81 <= c0 && c0 < 0xff:
			if nSrc+1 >= len(src) {
				if !atEOF {
					err = transform.ErrShortSrc
					break loop
				}
				r, size = utf8.RuneError, 1
				goto write
			}
			c1 := src[nSrc+1]
			switch {
			case 0x40 <= c1 && c1 < 0x7f:
				c1 -= 0x40
			case 0xa1 <= c1 && c1 < 0xff:
				c1 -= 0x62
			case c1 < 0x40:
				r, size = utf8.RuneError, 1
				goto write
			default:
				r, size = utf8.RuneError, 2
				goto write
			}
			r, size = '\ufffd', 2
			if i := int(c0-0x81)*157 + int(c1); i < len(decodeHK) {
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
				r = rune(decodeHK[i])
				if r == 0 {
					r = '\ufffd'
				}
			}

		default:
			r, size = utf8.RuneError, 1
		}

	write:
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
	return nDst, nSrc, err
}

type big5EncoderHK struct{ transform.NopResetter }

func (big5EncoderHK) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	r, size := rune(0), 0
	for ; nSrc < len(src); nSrc += size {
		r = rune(src[nSrc])

		// Decode a 1-byte rune.
		if r < utf8.RuneSelf {
			size = 1
			if nDst >= len(dst) {
				err = transform.ErrShortDst
				break
			}
			dst[nDst] = uint8(r)
			nDst++
			continue

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
			// func init checks that the switch covers all tables.
			switch {
			case encode0LowHK <= r && r < encode0HighHK:
				if r = rune(encode0HK[r-encode0LowHK]); r != 0 {
					goto write2
				}
			case encode1LowHK <= r && r < encode1HighHK:
				if r = rune(encode1HK[r-encode1LowHK]); r != 0 {
					goto write2
				}
			case encode2LowHK <= r && r < encode2HighHK:
				if r = rune(encode2HK[r-encode2LowHK]); r != 0 {
					goto write2
				}
			case encode3LowHK <= r && r < encode3HighHK:
				if r = rune(encode3HK[r-encode3LowHK]); r != 0 {
					goto write2
				}
			case encode4LowHK <= r && r < encode4HighHK:
				if r = rune(encode4HK[r-encode4LowHK]); r != 0 {
					goto write2
				}
			case encode5LowHK <= r && r < encode5HighHK:
				if r = rune(encode5HK[r-encode5LowHK]); r != 0 {
					goto write2
				}
			case encode6LowHK <= r && r < encode6HighHK:
				if r = rune(encode6HK[r-encode6LowHK]); r != 0 {
					goto write2
				}
			case encode7LowHK <= r && r < encode7HighHK:
				if r = rune(encode7HK[r-encode7LowHK]); r != 0 {
					goto write2
				}
			}
			err = internal.ErrASCIIReplacement
			break
		}

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

func init() {
	// Check that the hard-coded encode switch covers all tables.
	if numEncodeTablesHK != 8 {
		panic("bad numEncodeTables")
	}
}
