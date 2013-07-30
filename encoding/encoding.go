// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package encoding provides mappings from simple character encodings, such as
// IBM Code Page 437 and Windows 1252, to and from UTF-8. Mappings for rare or
// large encodings are provided in other packages, such as
// code.google.com/p/go.text/encoding/shiftjis.
package encoding

import (
	"errors"
	"unicode/utf8"

	"code.google.com/p/go.text/transform"
)

// Encoding is a character set encoding that can be transformed to and from
// UTF-8.
type Encoding interface {
	// NewDecoder returns a transformer that converts to UTF-8.
	//
	// Transforming source bytes that are not of that encoding will not
	// result in an error per se. Each byte that cannot be transcoded will
	// be represented in the output by the UTF-8 encoding of '\uFFFD', the
	// replacement rune.
	NewDecoder() transform.Transformer

	// NewEncoder returns a transformer that converts from UTF-8.
	//
	// Transforming source bytes that are not valid UTF-8 will not result in
	// an error per se. Each rune that cannot be transcoded will be
	// represented in the output by an encoding-specific replacement such as
	// "\x1a" (the ASCII substitute character) or "\xff\xfd". To return
	// early with error instead, use transform.Chain to preprocess the data
	// with a UTF8Validator.
	NewEncoder() transform.Transformer
}

// ErrInvalidUTF8 means that a transformer encountered invalid UTF-8.
var ErrInvalidUTF8 = errors.New("encoding: invalid UTF-8")

// UTF8Validator is a transformer that returns ErrInvalidUTF8 on the first
// input byte that is not valid UTF-8.
var UTF8Validator transform.Transformer = utf8Validator{}

type utf8Validator struct{}

func (utf8Validator) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	n := len(src)
	if n > len(dst) {
		n = len(dst)
	}
	for i := 0; i < n; {
		if c := src[i]; c < utf8.RuneSelf {
			dst[i] = c
			i++
			continue
		}
		_, size := utf8.DecodeRune(src[i:])
		if size == 1 {
			// All valid runes of size 1 (those below utf8.RuneSelf) were
			// handled above. We have invalid UTF-8.
			err = ErrInvalidUTF8
			if !atEOF && !utf8.FullRune(src[i:]) {
				err = transform.ErrShortSrc
			}
			return i, i, err
		}
		if i+size > len(dst) {
			return i, i, transform.ErrShortDst
		}
		for ; size > 0; size-- {
			dst[i] = src[i]
			i++
		}
	}
	if len(src) > len(dst) {
		err = transform.ErrShortDst
	}
	return n, n, err
}
