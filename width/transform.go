// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package width

import (
	"unicode/utf8"

	"golang.org/x/text/transform"
)

type foldTransform struct {
	transform.NopResetter
}

func (foldTransform) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for nSrc < len(src) {
		// TODO: ASCII fast path.
		v, size := trie.lookup(src[nSrc:])
		if size == 0 { // incomplete UTF-8 encoding
			if !atEOF {
				return nDst, nSrc, transform.ErrShortSrc
			}
			size = 1 // gobble 1 byte
		}
		if elem(v)&hasMappingMask == 0 || v == tagHalfwidth /* Won Sign */ {
			if size != copy(dst[nDst:], src[nSrc:nSrc+size]) {
				return nDst, nSrc, transform.ErrShortDst
			}
			nDst += size
		} else {
			r := rune(v &^ tagMappingMask)
			if utf8.RuneLen(r) > len(dst)-nDst {
				return nDst, nSrc, transform.ErrShortDst
			}
			nDst += utf8.EncodeRune(dst[nDst:], r)
		}
		nSrc += size
	}
	return nDst, nSrc, nil
}
