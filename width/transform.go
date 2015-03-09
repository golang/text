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
		if src[nSrc] < utf8.RuneSelf {
			// ASCII fast path.
			start, end := nSrc, len(src)
			if d := len(dst) - nDst; d < end-start {
				end = nSrc + d
			}
			for nSrc++; nSrc < end && src[nSrc] < utf8.RuneSelf; nSrc++ {
			}
			nDst += copy(dst[nDst:], src[start:nSrc])
			if nDst == len(dst) && nSrc < len(src) && src[nSrc] < utf8.RuneSelf {
				return nDst, nSrc, transform.ErrShortDst
			}
			continue
		}
		v, size := trie.lookup(src[nSrc:])
		if size == 0 { // incomplete UTF-8 encoding
			if !atEOF {
				return nDst, nSrc, transform.ErrShortSrc
			}
			size = 1 // gobble 1 byte
		}
		if elem(v)&tagNeedsFold == 0 {
			if size != copy(dst[nDst:], src[nSrc:nSrc+size]) {
				return nDst, nSrc, transform.ErrShortDst
			}
			nDst += size
		} else {
			data := inverseData[byte(v)]
			if len(dst)-nDst < int(data[0]) {
				return nDst, nSrc, transform.ErrShortDst
			}
			i := 1
			for end := int(data[0]); i < end; i++ {
				dst[nDst] = data[i]
				nDst++
			}
			dst[nDst] = data[i] ^ src[nSrc+size-1]
			nDst++
		}
		nSrc += size
	}
	return nDst, nSrc, nil
}
