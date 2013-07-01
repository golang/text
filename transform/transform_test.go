// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package transform

import (
	"errors"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
)

type lowerCaseASCII struct{}

func (lowerCaseASCII) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	n := len(src)
	if n > len(dst) {
		n, err = len(dst), ErrShortDst
	}
	for i, c := range src[:n] {
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		dst[i] = c
	}
	return n, n, err
}

var errYouMentionedX = errors.New("you mentioned X")

type dontMentionX struct{}

func (dontMentionX) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	n := len(src)
	if n > len(dst) {
		n, err = len(dst), ErrShortDst
	}
	for i, c := range src[:n] {
		if c == 'X' {
			return i, i, errYouMentionedX
		}
		dst[i] = c
	}
	return n, n, err
}

// rleDecode and rleEncode implement a toy run-length encoding: "aabbbbbbbbbb"
// is encoded as "2a10b". The decoding is assumed to not contain any numbers.

type rleDecode struct{}

func (rleDecode) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
loop:
	for len(src) > 0 {
		n := 0
		for i, c := range src {
			if '0' <= c && c <= '9' {
				n = 10*n + int(c-'0')
				continue
			}
			if i == 0 {
				return nDst, nSrc, errors.New("rleDecode: bad input")
			}
			if n > len(dst) {
				return nDst, nSrc, ErrShortDst
			}
			for j := 0; j < n; j++ {
				dst[j] = c
			}
			dst, src = dst[n:], src[i+1:]
			nDst, nSrc = nDst+n, nSrc+i+1
			continue loop
		}
		if atEOF {
			return nDst, nSrc, errors.New("rleDecode: bad input")
		}
		return nDst, nSrc, ErrShortSrc
	}
	return nDst, nSrc, nil
}

type rleEncode struct {
	// allowStutter means that "xxxxxxxx" can be encoded as "5x3x"
	// instead of always as "8x".
	allowStutter bool
}

func (e rleEncode) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	for len(src) > 0 {
		n, c0 := len(src), src[0]
		for i, c := range src[1:] {
			if c != c0 {
				n = i + 1
				break
			}
		}
		if n == len(src) && !atEOF && !e.allowStutter {
			return nDst, nSrc, ErrShortSrc
		}
		s := strconv.Itoa(n)
		if len(s) >= len(dst) {
			return nDst, nSrc, ErrShortDst
		}
		copy(dst, s)
		dst[len(s)] = c0
		dst, src = dst[len(s)+1:], src[n:]
		nDst, nSrc = nDst+len(s)+1, nSrc+n
	}
	return nDst, nSrc, nil
}

func TestReader(t *testing.T) {
	testCases := []struct {
		desc    string
		t       Transformer
		src     string
		dstSize int
		srcSize int
		wantStr string
		wantErr error
	}{
		{
			"lowerCaseASCII",
			lowerCaseASCII{},
			"Hello WORLD.",
			100,
			100,
			"hello world.",
			nil,
		},

		{
			"lowerCaseASCII; small dst",
			lowerCaseASCII{},
			"Hello WORLD.",
			3,
			100,
			"hello world.",
			nil,
		},

		{
			"lowerCaseASCII; small src",
			lowerCaseASCII{},
			"Hello WORLD.",
			100,
			4,
			"hello world.",
			nil,
		},

		{
			"lowerCaseASCII; small buffers",
			lowerCaseASCII{},
			"Hello WORLD.",
			3,
			4,
			"hello world.",
			nil,
		},

		{
			"lowerCaseASCII; very small buffers",
			lowerCaseASCII{},
			"Hello WORLD.",
			1,
			1,
			"hello world.",
			nil,
		},

		{
			"dontMentionX",
			dontMentionX{},
			"The First Rule of Transform Club: don't mention Mister X, ever.",
			100,
			100,
			"The First Rule of Transform Club: don't mention Mister ",
			errYouMentionedX,
		},

		{
			"dontMentionX; small buffers",
			dontMentionX{},
			"The First Rule of Transform Club: don't mention Mister X, ever.",
			10,
			10,
			"The First Rule of Transform Club: don't mention Mister ",
			errYouMentionedX,
		},

		{
			"dontMentionX; very small buffers",
			dontMentionX{},
			"The First Rule of Transform Club: don't mention Mister X, ever.",
			1,
			1,
			"The First Rule of Transform Club: don't mention Mister ",
			errYouMentionedX,
		},

		{
			"rleDecode",
			rleDecode{},
			"1a2b3c10d11e0f1g",
			100,
			100,
			"abbcccddddddddddeeeeeeeeeeeg",
			nil,
		},

		{
			"rleDecode; long",
			rleDecode{},
			"12a23b34c45d56e99z",
			100,
			100,
			strings.Repeat("a", 12) +
				strings.Repeat("b", 23) +
				strings.Repeat("c", 34) +
				strings.Repeat("d", 45) +
				strings.Repeat("e", 56) +
				strings.Repeat("z", 99),
			nil,
		},

		{
			"rleDecode; tight buffers",
			rleDecode{},
			"1a2b3c10d11e0f1g",
			11,
			3,
			"abbcccddddddddddeeeeeeeeeeeg",
			nil,
		},

		{
			"rleDecode; short dst",
			rleDecode{},
			"1a2b3c10d11e0f1g",
			10,
			3,
			"abbcccdddddddddd",
			ErrShortDst,
		},

		{
			"rleDecode; short src",
			rleDecode{},
			"1a2b3c10d11e0f1g",
			11,
			2,
			"abbccc",
			ErrShortSrc,
		},

		{
			"rleEncode",
			rleEncode{},
			"abbcccddddddddddeeeeeeeeeeeg",
			100,
			100,
			"1a2b3c10d11e1g",
			nil,
		},

		{
			"rleEncode; long",
			rleEncode{},
			strings.Repeat("a", 12) +
				strings.Repeat("b", 23) +
				strings.Repeat("c", 34) +
				strings.Repeat("d", 45) +
				strings.Repeat("e", 56) +
				strings.Repeat("z", 99),
			100,
			100,
			"12a23b34c45d56e99z",
			nil,
		},

		{
			"rleEncode; tight buffers",
			rleEncode{},
			"abbcccddddddddddeeeeeeeeeeeg",
			3,
			12,
			"1a2b3c10d11e1g",
			nil,
		},

		{
			"rleEncode; short dst",
			rleEncode{},
			"abbcccddddddddddeeeeeeeeeeeg",
			2,
			12,
			"1a2b3c",
			ErrShortDst,
		},

		{
			"rleEncode; short src",
			rleEncode{},
			"abbcccddddddddddeeeeeeeeeeeg",
			3,
			11,
			"1a2b3c10d",
			ErrShortSrc,
		},

		{
			"rleEncode; allowStutter = false",
			rleEncode{allowStutter: false},
			"aaaabbbbbbbbccccddddd",
			10,
			10,
			"4a8b4c5d",
			nil,
		},

		{
			"rleEncode; allowStutter = true",
			rleEncode{allowStutter: true},
			"aaaabbbbbbbbccccddddd",
			10,
			10,
			"4a6b2b4c4d1d",
			nil,
		},
	}

	for _, tc := range testCases {
		r := NewReader(strings.NewReader(tc.src), tc.t)
		// Differently sized dst and src buffers are not part of the
		// exported API. We override them manually.
		r.dst = make([]byte, tc.dstSize)
		r.src = make([]byte, tc.srcSize)
		got, err := ioutil.ReadAll(r)
		str := string(got)
		if str != tc.wantStr || err != tc.wantErr {
			t.Errorf("%s:\ngot  %q, %v\nwant %q, %v", tc.desc, str, err, tc.wantStr, tc.wantErr)
		}
	}
}
