// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package charmap

import (
	"testing"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/internal"
	"golang.org/x/text/encoding/internal/enctest"
	"golang.org/x/text/transform"
)

func dec(e encoding.Encoding) (dir string, t transform.Transformer, err error) {
	return "Decode", e.NewDecoder(), nil
}

func encASCIISuperset(e encoding.Encoding) (dir string, t transform.Transformer, err error) {
	return "Encode", e.NewEncoder(), internal.ErrASCIIReplacement
}

func encEBCDIC(e encoding.Encoding) (dir string, t transform.Transformer, err error) {
	return "Encode", e.NewEncoder(), internal.RepertoireError(0x3f)
}

func TestNonRepertoire(t *testing.T) {
	testCases := []struct {
		init      func(e encoding.Encoding) (string, transform.Transformer, error)
		e         encoding.Encoding
		src, want string
	}{
		{dec, Windows1252, "\x81", "\ufffd"},

		{encEBCDIC, CodePage037, "갂", ""},

		{encEBCDIC, CodePage1047, "갂", ""},
		{encEBCDIC, CodePage1047, "a¤갂", "\x81\x9F"},

		{encEBCDIC, CodePage1140, "갂", ""},
		{encEBCDIC, CodePage1140, "a€갂", "\x81\x9F"},

		{encASCIISuperset, Windows1252, "갂", ""},
		{encASCIISuperset, Windows1252, "a갂", "a"},
		{encASCIISuperset, Windows1252, "\u00E9갂", "\xE9"},
	}
	for _, tc := range testCases {
		dir, tr, wantErr := tc.init(tc.e)

		dst, _, err := transform.String(tr, tc.src)
		if err != wantErr {
			t.Errorf("%s %v(%q): got %v; want %v", dir, tc.e, tc.src, err, wantErr)
		}
		if got := string(dst); got != tc.want {
			t.Errorf("%s %v(%q):\ngot  %q\nwant %q", dir, tc.e, tc.src, got, tc.want)
		}
	}
}

func TestBasics(t *testing.T) {
	testCases := []struct {
		e       encoding.Encoding
		encoded string
		utf8    string
	}{{
		e:       CodePage037,
		encoded: "\xc8\x51\xba\x93\xcf",
		utf8:    "Hé[lõ",
	}, {
		e:       CodePage437,
		encoded: "H\x82ll\x93 \x9d\xa7\xf4\x9c\xbe",
		utf8:    "Héllô ¥º⌠£╛",
	}, {
		e:       CodePage866,
		encoded: "H\xf3\xd3o \x98\xfd\x9f\xdd\xa1",
		utf8:    "Hє╙o Ш¤Я▌б",
	}, {
		e:       CodePage1047,
		encoded: "\xc8\x54\x93\x93\x9f",
		utf8:    "Hèll¤",
	}, {
		e:       CodePage1125,
		encoded: "Hello.\x20\x80\x81\x82\x83\xF2\x84\x85\xF4\xF0\x86\x87\x88\xF6\xF8\x89\x8A\x8B\x8C\x8D\x8E\x8F\x90\x91\x92\x93\x94\x95\x96\x97\x98\x99\x9A\x9B\x9C\x9D\x9E\x9F\xA0\xA1\xA2\xA3\xF3\xA4\xA5\xF5\xF1\xA6\xA7\xA8\xF7\xF9\xA9\xAA\xAB\xAC\xAD\xAE\xAF\xE0\xE1\xE2\xE3\xE4\xE5\xE6\xE7\xE8\xE9\xEA\xEB\xEC\xED\xEE\xEF\xB0\xB1\xB2\xB3\xB4\xB5\xB6\xB7\xB8\xB9\xBA\xBB\xBC\xBD\xBE\xBF\xC0\xC1\xC2\xC3\xC4\xC5\xC6\xC7\xC8\xC9\xCA\xCB\xCC\xCD\xCE\xCF\xD0\xD1\xD2\xD3\xD4\xD5\xD6\xD7\xD8\xD9\xDA\xDB\xDC\xDD\xDE\xDF\xFA\xFB\xFC\xFD\xFE",
		utf8:    "Hello. АБВГҐДЕЄЁЖЗИІЇЙКЛМНОПРСТУФХЦЧШЩЪЫЬЭЮЯабвгґдеєёжзиіїйклмнопрстуфхцчшщъыьэюя░▒▓│┤╡╢╖╕╣║╗╝╜╛┐└┴┬├─┼╞╟╚╔╩╦╠═╬╧╨╤╥╙╘╒╓╫╪┘┌█▄▌▐▀·√№¤■",
	}, {
		e:       CodePage1140,
		encoded: "\xc8\x9f\x93\x93\xcf",
		utf8:    "H€llõ",
	}, {
		e:       ISO8859_2,
		encoded: "Hel\xe5\xf5",
		utf8:    "Helĺő",
	}, {
		e:       ISO8859_3,
		encoded: "He\xbd\xd4",
		utf8:    "He½Ô",
	}, {
		e:       ISO8859_4,
		encoded: "Hel\xb6\xf8",
		utf8:    "Helļø",
	}, {
		e:       ISO8859_5,
		encoded: "H\xd7\xc6o",
		utf8:    "HзЦo",
	}, {
		e:       ISO8859_6,
		encoded: "Hel\xc2\xc9",
		utf8:    "Helآة",
	}, {
		e:       ISO8859_7,
		encoded: "H\xeel\xebo",
		utf8:    "Hξlλo",
	}, {
		e:       ISO8859_8,
		encoded: "Hel\xf5\xed",
		utf8:    "Helץם",
	}, {
		e:       ISO8859_9,
		encoded: "\xdeayet",
		utf8:    "Şayet",
	}, {
		e:       ISO8859_10,
		encoded: "H\xea\xbfo",
		utf8:    "Hęŋo",
	}, {
		e:       ISO8859_13,
		encoded: "H\xe6l\xf9o",
		utf8:    "Hęlło",
	}, {
		e:       ISO8859_14,
		encoded: "He\xfe\xd0o",
		utf8:    "HeŷŴo",
	}, {
		e:       ISO8859_15,
		encoded: "H\xa4ll\xd8",
		utf8:    "H€llØ",
	}, {
		e:       ISO8859_16,
		encoded: "H\xe6ll\xbd",
		utf8:    "Hællœ",
	}, {
		e:       KOI8R,
		encoded: "He\x93\xad\x9c",
		utf8:    "He⌠╜°",
	}, {
		e:       KOI8U,
		encoded: "He\x93\xad\x9c",
		utf8:    "He⌠ґ°",
	}, {
		e:       Macintosh,
		encoded: "He\xdf\xd7",
		utf8:    "Heﬂ◊",
	}, {
		e:       MacintoshCyrillic,
		encoded: "He\xbe\x94",
		utf8:    "HeЊФ",
	}, {
		e:       Windows874,
		encoded: "He\xb7\xf0",
		utf8:    "Heท๐",
	}, {
		e:       Windows1250,
		encoded: "He\xe5\xe5o",
		utf8:    "Heĺĺo",
	}, {
		e:       Windows1251,
		encoded: "H\xball\xfe",
		utf8:    "Hєllю",
	}, {
		e:       Windows1252,
		encoded: "H\xe9ll\xf4 \xa5\xbA\xae\xa3\xd0",
		utf8:    "Héllô ¥º®£Ð",
	}, {
		e:       Windows1253,
		encoded: "H\xe5ll\xd6",
		utf8:    "HεllΦ",
	}, {
		e:       Windows1254,
		encoded: "\xd0ello",
		utf8:    "Ğello",
	}, {
		e:       Windows1255,
		encoded: "He\xd4o",
		utf8:    "Heװo",
	}, {
		e:       Windows1256,
		encoded: "H\xdbllo",
		utf8:    "Hغllo",
	}, {
		e:       Windows1257,
		encoded: "He\xeflo",
		utf8:    "Heļlo",
	}, {
		e:       Windows1258,
		encoded: "Hell\xf5",
		utf8:    "Hellơ",
	}, {
		e:       XUserDefined,
		encoded: "\x00\x40\x7f\x80\xab\xff",
		utf8:    "\u0000\u0040\u007f\uf780\uf7ab\uf7ff",
	}}

	for _, tc := range testCases {
		enctest.TestEncoding(t, tc.e, tc.encoded, tc.utf8, "", "")
	}
}

var windows1255TestCases = []struct {
	b  byte
	ok bool
	r  rune
}{
	{'\x00', true, '\u0000'},
	{'\x1a', true, '\u001a'},
	{'\x61', true, '\u0061'},
	{'\x7f', true, '\u007f'},
	{'\x80', true, '\u20ac'},
	{'\x95', true, '\u2022'},
	{'\xa0', true, '\u00a0'},
	{'\xc0', true, '\u05b0'},
	{'\xfc', true, '\ufffd'},
	{'\xfd', true, '\u200e'},
	{'\xfe', true, '\u200f'},
	{'\xff', true, '\ufffd'},
	{encoding.ASCIISub, false, '\u0400'},
	{encoding.ASCIISub, false, '\u2603'},
	{encoding.ASCIISub, false, '\U0001f4a9'},
}

func TestDecodeByte(t *testing.T) {
	for _, tc := range windows1255TestCases {
		if !tc.ok {
			continue
		}

		got := Windows1255.DecodeByte(tc.b)
		want := tc.r
		if got != want {
			t.Errorf("DecodeByte(%#02x): got %#08x, want %#08x", tc.b, got, want)
		}
	}
}

func TestEncodeRune(t *testing.T) {
	for _, tc := range windows1255TestCases {
		// There can be multiple tc.b values that map to tc.r = '\ufffd'.
		if tc.r == '\ufffd' {
			continue
		}

		gotB, gotOK := Windows1255.EncodeRune(tc.r)
		wantB, wantOK := tc.b, tc.ok
		if gotB != wantB || gotOK != wantOK {
			t.Errorf("EncodeRune(%#08x): got (%#02x, %t), want (%#02x, %t)", tc.r, gotB, gotOK, wantB, wantOK)
		}
	}
}

func TestFiles(t *testing.T) { enctest.TestFile(t, Windows1252) }

func BenchmarkEncoding(b *testing.B) { enctest.Benchmark(b, Windows1252) }
