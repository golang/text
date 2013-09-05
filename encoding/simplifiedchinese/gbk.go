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
			case 0x40 <= c1 && c1 <= 0x7e:
				c1 -= 0x40
			case 0x80 <= c1 && c1 <= 0xfe:
				c1 -= 0x41
			default:
				err = errInvalidGBK
				break loop
			}
			r, size = encoding.ASCIISub, 2
			if i := int(c0-0x81)*190 + int(c1); i < len(gbkDecode) {
				r = rune(gbkDecode[i])
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

		// Microsoft's Code Page 936 extends GBK 1.0 to encode the euro sign U+20AC
		// as 0x80. The HTML5 specification at http://encoding.spec.whatwg.org/#gbk
		// says to treat "gbk" as Code Page 936.
		case r == '€':
			r = 0x80

		case 0xffff < r:
			r = encoding.ASCIISub

		default:
			e := gbkEncode[uint16(r)]
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
