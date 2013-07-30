// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encoding

import (
	"bytes"
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"code.google.com/p/go.text/transform"
)

func trim(s string) string {
	if len(s) < 120 {
		return s
	}
	return s[:50] + "..." + s[len(s)-50:]
}

var basicTestCases = []struct {
	e         Encoding
	encPrefix string
	encoded   string
	utf8      string
}{
	{
		e:       CodePage437,
		encoded: "H\x82ll\x93 \x9d\xa7\xf4\x9c\xbe",
		utf8:    "Héllô ¥º⌠£╛",
	},
	{
		e:       Windows1252,
		encoded: "H\xe9ll\xf4 \xa5\xbA\xae\xa3\xd0",
		utf8:    "Héllô ¥º®£Ð",
	},
	{
		e:       UTF16(BigEndian, IgnoreBOM),
		encoded: "\x00\x57\x00\xe4\xd8\x35\xdd\x65",
		utf8:    "\x57\u00e4\U0001d565",
	},
	{
		e:         UTF16(BigEndian, ExpectBOM),
		encPrefix: "\xfe\xff",
		encoded:   "\x00\x57\x00\xe4\xd8\x35\xdd\x65",
		utf8:      "\x57\u00e4\U0001d565",
	},
	{
		e:       UTF16(LittleEndian, IgnoreBOM),
		encoded: "\x57\x00\xe4\x00\x35\xd8\x65\xdd",
		utf8:    "\x57\u00e4\U0001d565",
	},
	{
		e:         UTF16(LittleEndian, ExpectBOM),
		encPrefix: "\xff\xfe",
		encoded:   "\x57\x00\xe4\x00\x35\xd8\x65\xdd",
		utf8:      "\x57\u00e4\U0001d565",
	},
}

func TestBasics(t *testing.T) {
	for _, tc := range basicTestCases {
		for _, direction := range []string{"Decode", "Encode"} {
			newTransformer, want, src := (func() transform.Transformer)(nil), "", ""
			wPrefix, sPrefix := "", ""
			if direction == "Decode" {
				newTransformer, want, src = tc.e.NewDecoder, tc.utf8, tc.encoded
				wPrefix, sPrefix = "", tc.encPrefix
			} else {
				newTransformer, want, src = tc.e.NewEncoder, tc.encoded, tc.utf8
				wPrefix, sPrefix = tc.encPrefix, ""
			}

			dst := make([]byte, 1024)
			nDst, nSrc, err := newTransformer().Transform(dst, []byte(sPrefix+src), true)
			if err != nil {
				t.Errorf("%v: %s: %v", tc.e, direction, err)
				continue
			}
			if nSrc != len(sPrefix)+len(src) {
				t.Errorf("%v: %s: nSrc got %d, want %d",
					tc.e, direction, nSrc, len(sPrefix)+len(src))
				continue
			}
			if got := string(dst[:nDst]); got != wPrefix+want {
				t.Errorf("%v: %s:\ngot  %q\nwant %q",
					tc.e, direction, got, wPrefix+want)
				continue
			}

			for _, n := range []int{0, 1, 2, 10, 123, 4567} {
				sr := strings.NewReader(sPrefix + strings.Repeat(src, n))
				g, err := ioutil.ReadAll(transform.NewReader(sr, newTransformer()))
				if err != nil {
					t.Errorf("%v: %s: ReadAll: n=%d: %v", tc.e, direction, n, err)
					continue
				}
				got1, want1 := string(g), wPrefix+strings.Repeat(want, n)
				if got1 != want1 {
					t.Errorf("%v: %s: ReadAll: n=%d\ngot  %q\nwant %q",
						tc.e, direction, n, trim(got1), trim(want1))
					continue
				}
			}
		}
	}
}

func TestEncodeInvalidUTF8(t *testing.T) {
	inputs := []string{
		"hello.",
		"wo\ufffdld.",
		"ABC\xff\x80\x80", // Invalid UTF-8.
		"\x80\x80\x80\x80\x80",
		"\x80\x80D\x80\x80",          // Synchronization at "D".
		"E\xed\xa0\x80\xed\xbf\xbfF", // Two invalid UTF-8 runes (surrogates).
		"G",
		"H\xe2\x82",     // U+20AC in UTF-8 is "\xe2\x82\xac", which we split over two
		"\xacI\xe2\x82", // input lines. It maps to 0x80 in the Windows-1252 encoding.
	}
	want := strings.Replace("hello.wo?ld.ABC?D?E??FGH\x80I?", "?", "\x1a", -1)

	transformer := Windows1252.NewEncoder()
	gotBuf := make([]byte, 0, 1024)
	src := make([]byte, 0, 1024)
	for i, input := range inputs {
		dst := make([]byte, 1024)
		src = append(src, input...)
		atEOF := i == len(inputs)-1
		nDst, nSrc, err := transformer.Transform(dst, src, atEOF)
		gotBuf = append(gotBuf, dst[:nDst]...)
		src = src[nSrc:]
		if err != nil && err != transform.ErrShortSrc {
			t.Fatalf("i=%d: %v", i, err)
		}
		if atEOF && err != nil {
			t.Fatalf("atEOF: %v", i, err)
		}
	}
	if got := string(gotBuf); got != want {
		t.Fatalf("\ngot  %+q\nwant %+q", got, want)
	}
}

func TestUTF8Validator(t *testing.T) {
	testCases := []struct {
		desc    string
		dstSize int
		src     string
		atEOF   bool
		want    string
		wantErr error
	}{
		{
			"empty input",
			100,
			"",
			false,
			"",
			nil,
		},
		{
			"valid 1-byte 1-rune input",
			100,
			"a",
			false,
			"a",
			nil,
		},
		{
			"valid 3-byte 1-rune input",
			100,
			"\u1234",
			false,
			"\u1234",
			nil,
		},
		{
			"valid 5-byte 3-rune input",
			100,
			"a\u0100\u0101",
			false,
			"a\u0100\u0101",
			nil,
		},
		{
			"perfectly sized dst (non-ASCII)",
			5,
			"a\u0100\u0101",
			false,
			"a\u0100\u0101",
			nil,
		},
		{
			"short dst (non-ASCII)",
			4,
			"a\u0100\u0101",
			false,
			"a\u0100",
			transform.ErrShortDst,
		},
		{
			"perfectly sized dst (ASCII)",
			5,
			"abcde",
			false,
			"abcde",
			nil,
		},
		{
			"short dst (ASCII)",
			4,
			"abcde",
			false,
			"abcd",
			transform.ErrShortDst,
		},
		{
			"partial input (!EOF)",
			100,
			"a\u0100\xf1",
			false,
			"a\u0100",
			transform.ErrShortSrc,
		},
		{
			"invalid input (EOF)",
			100,
			"a\u0100\xf1",
			true,
			"a\u0100",
			ErrInvalidUTF8,
		},
		{
			"invalid input (!EOF)",
			100,
			"a\u0100\x80",
			false,
			"a\u0100",
			ErrInvalidUTF8,
		},
		{
			"invalid input (above U+10FFFF)",
			100,
			"a\u0100\xf7\xbf\xbf\xbf",
			false,
			"a\u0100",
			ErrInvalidUTF8,
		},
		{
			"invalid input (surrogate half)",
			100,
			"a\u0100\xed\xa0\x80",
			false,
			"a\u0100",
			ErrInvalidUTF8,
		},
	}
	for _, tc := range testCases {
		dst := make([]byte, tc.dstSize)
		nDst, nSrc, err := utf8Validator{}.Transform(dst, []byte(tc.src), tc.atEOF)
		if nDst < 0 || len(dst) < nDst {
			t.Errorf("%s: nDst=%d out of range", tc.desc, nDst)
			continue
		}
		got := string(dst[:nDst])
		if got != tc.want || nSrc != len(tc.want) || err != tc.wantErr {
			t.Errorf("%s:\ngot  %+q, %d, %v\nwant %+q, %d, %v",
				tc.desc, got, nSrc, err, tc.want, len(tc.want), tc.wantErr)
			continue
		}
	}
}

// TODO: UTF-16-specific tests:
// - inputs with multiple U+FEFF and U+FFFE runes. These should not be replaced
//   by U+FFFD.
// - malformed input: an odd number of bytes (and atEOF), or unmatched
//   surrogates. These should be replaced with U+FFFD.

func benchmark(b *testing.B, dstFile, srcFile string, newTransformer func() transform.Transformer) {
	dst, err := ioutil.ReadFile(dstFile)
	if err != nil {
		b.Fatal(err)
	}
	src, err := ioutil.ReadFile(srcFile)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := transform.NewReader(bytes.NewReader(src), newTransformer())
		n, err := io.Copy(ioutil.Discard, r)
		if err != nil {
			b.Fatal(err)
		}
		if n != int64(len(dst)) {
			b.Fatalf("copied %d bytes, want %d", n, len(dst))
		}
	}
}

func BenchmarkCharmapDecoder(b *testing.B) {
	benchmark(
		b,
		"testdata/candide-utf-8.txt",
		"testdata/candide-windows-1252.txt",
		Windows1252.NewDecoder,
	)
}

func BenchmarkCharmapEncoder(b *testing.B) {
	benchmark(
		b,
		"testdata/candide-windows-1252.txt",
		"testdata/candide-utf-8.txt",
		Windows1252.NewEncoder,
	)
}

func BenchmarkUTF16Decoder(b *testing.B) {
	benchmark(
		b,
		"testdata/candide-utf-8.txt",
		"testdata/candide-utf-16le.txt",
		UTF16(LittleEndian, IgnoreBOM).NewDecoder,
	)
}

func BenchmarkUTF16Encoder(b *testing.B) {
	benchmark(
		b,
		"testdata/candide-utf-16le.txt",
		"testdata/candide-utf-8.txt",
		UTF16(LittleEndian, IgnoreBOM).NewEncoder,
	)
}
