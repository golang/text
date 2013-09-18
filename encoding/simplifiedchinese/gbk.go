// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package simplifiedchinese

import (
	"errors"
	"unicode/utf8"

	"code.google.com/p/go.text/encoding"
	"code.google.com/p/go.text/transform"
)

// GBK is the GBK encoding. It encodes an extension of the GB2312 character set
// and is also known as Code Page 936.
var GBK encoding.Encoding = gbk{}

type gbk struct{}

func (gbk) NewDecoder() transform.Transformer {
	return gbkDecoder{}
}

func (gbk) NewEncoder() transform.Transformer {
	return gbkEncoder{}
}

func (gbk) String() string {
	return "GBK"
}

var errInvalidGBK = errors.New("simplifiedchinese: invalid GBK encoding")

type gbkDecoder struct{}

func (gbkDecoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	r, size := rune(0), 0
loop:
	for ; nSrc < len(src); nSrc += size {
		switch c0 := src[nSrc]; {
		case c0 < utf8.RuneSelf:
			r, size = rune(c0), 1

		// Microsoft's Code Page 936 extends GBK 1.0 to encode the euro sign U+20AC
		// as 0x80. The HTML5 specification at http://encoding.spec.whatwg.org/#gbk
		// says to treat "gbk" as Code Page 936.
		case c0 == 0x80:
			r, size = '€', 1

		case c0 < 0xff:
			if nSrc+1 >= len(src) {
				err = transform.ErrShortSrc
				break loop
			}
			c1 := src[nSrc+1]
			switch {
			case 0x40 <= c1 && c1 < 0x7f:
				c1 -= 0x40
			case 0x80 <= c1 && c1 < 0xff:
				c1 -= 0x41
			default:
				err = errInvalidGBK
				break loop
			}
			r, size = encoding.ASCIISub, 2
			if i := int(c0-0x81)*190 + int(c1); i < len(decode) {
				r = rune(decode[i])
				if r == 0 {
					r = encoding.ASCIISub
				}
			}

		default:
			err = errInvalidGBK
			break loop
		}

		if nDst+utf8.RuneLen(r) > len(dst) {
			err = transform.ErrShortDst
			break loop
		}
		nDst += utf8.EncodeRune(dst[nDst:], r)
	}
	if atEOF && err == transform.ErrShortSrc {
		err = errInvalidGBK
	}
	return nDst, nSrc, err
}

type gbkEncoder struct{}

func (gbkEncoder) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
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

			// func init checks that the switch covers all tables.
			switch {
			case encode0Low <= r && r < encode0High:
				if r = rune(encode0[r-encode0Low]); r != 0 {
					goto write2
				}
			case encode1Low <= r && r < encode1High:
				// Microsoft's Code Page 936 extends GBK 1.0 to encode the euro sign U+20AC
				// as 0x80. The HTML5 specification at http://encoding.spec.whatwg.org/#gbk
				// says to treat "gbk" as Code Page 936.
				if r == '€' {
					r = 0x80
					goto write1
				}
				if r = rune(encode1[r-encode1Low]); r != 0 {
					goto write2
				}
			case encode2Low <= r && r < encode2High:
				if r = rune(encode2[r-encode2Low]); r != 0 {
					goto write2
				}
			case encode3Low <= r && r < encode3High:
				if r = rune(encode3[r-encode3Low]); r != 0 {
					goto write2
				}
			case encode4Low <= r && r < encode4High:
				if r = rune(encode4[r-encode4Low]); r != 0 {
					goto write2
				}
			}
			r = encoding.ASCIISub
		}

	write1:
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

func init() {
	// Check that the hard-coded encode switch covers all tables.
	if numEncodeTables != 5 {
		panic("bad numEncodeTables")
	}
}
