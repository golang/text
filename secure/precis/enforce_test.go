// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package precis

import (
	"testing"

	"golang.org/x/text/secure/bidirule"
)

type testCase struct {
	input  string
	output string
	err    error
}

var testCases = []struct {
	name  string
	p     *Profile
	cases []testCase
}{
	{"Nickname", Nickname, []testCase{
		{"  Swan  of   Avon   ", "Swan of Avon", nil},
		{"", "", errEmptyString},
		{" ", "", errEmptyString},
		{"  ", "", errEmptyString},
		{"a\u00A0a\u1680a\u2000a\u2001a\u2002a\u2003a\u2004a\u2005a\u2006a\u2007a\u2008a\u2009a\u200Aa\u202Fa\u205Fa\u3000a", "a a a a a a a a a a a a a a a a a", nil},
		{"Foo", "Foo", nil},
		{"foo", "foo", nil},
		{"Foo Bar", "Foo Bar", nil},
		{"foo bar", "foo bar", nil},
		{"\u03C3", "\u03C3", nil},
		// Greek final sigma is left as is (do not fold!)
		{"\u03C2", "\u03C2", nil},
		{"\u265A", "♚", nil},
		{"Richard \u2163", "Richard IV", nil},
		{"\u212B", "Å", nil},
		{"\uFB00", "ff", nil}, // because of NFKC
		{"שa", "שa", nil},     // no bidi rule
		{"동일조건변경허락", "동일조건변경허락", nil},
	}},
	{"OpaqueString", OpaqueString, []testCase{
		{"  Swan  of   Avon   ", "  Swan  of   Avon   ", nil},
		{"", "", errEmptyString},
		{" ", " ", nil},
		{"  ", "  ", nil},
		{"a\u00A0a\u1680a\u2000a\u2001a\u2002a\u2003a\u2004a\u2005a\u2006a\u2007a\u2008a\u2009a\u200Aa\u202Fa\u205Fa\u3000a", "a a a a a a a a a a a a a a a a a", nil},
		{"Foo", "Foo", nil},
		{"foo", "foo", nil},
		{"Foo Bar", "Foo Bar", nil},
		{"foo bar", "foo bar", nil},
		{"\u03C3", "\u03C3", nil},
		{"Richard \u2163", "Richard \u2163", nil},
		{"\u212B", "Å", nil},
		{"Jack of \u2666s", "Jack of \u2666s", nil},
		{"my cat is a \u0009by", "", errDisallowedRune},
		{"·", "", errDisallowedRune}, // Middle dot
		{"͵", "", errDisallowedRune}, // Keraia
		{"׳", "", errDisallowedRune},
		{"׳ה", "", errDisallowedRune},
		{"a׳b", "", errDisallowedRune},
		// TODO: Requires context rule to work.
		// {OpaqueString, "ש׳", "ש", nil},  // U+05e9 U+05f3
		{"שa", "שa", nil}, // no bidi rule

		// Katakana Middle Dot
		{"abc・def", "", errDisallowedRune},
		// TODO: These require context rules to work.
		// {OpaqueString, "aヅc・def", "", nil},
		// {OpaqueString, "abc・dぶf", "", nil},
		// {OpaqueString, "⺐bc・def", "", nil},

		// Arabic Indic Digit
		// TODO: These require context rules to work.
		// {OpaqueString, "١٢٣٤٥", "١٢٣٤٥", nil},
		// {OpaqueString, "۱۲۳۴۵", "۱۲۳۴۵", nil},
		{"١٢٣٤٥۶", "", errDisallowedRune},
		{"۱۲۳۴۵٦", "", errDisallowedRune},
	}},
	{"UsernameCaseMapped", UsernameCaseMapped, []testCase{
		// TODO: Should this work?
		// {UsernameCaseMapped, "", "", errDisallowedRune},
		{"juliet@example.com", "juliet@example.com", nil},
		{"fussball", "fussball", nil},
		{"fu\u00DFball", "fussball", nil},
		{"\u03C0", "\u03C0", nil},
		{"\u03A3", "\u03C3", nil},
		{"\u03C3", "\u03C3", nil},
		{"\u03C2", "\u03C3", nil},
		{"\u0049", "\u0069", nil},
		{"\u0049", "\u0069", nil},
		{"\u03D2", "", errDisallowedRune},
		{"\u03B0", "\u03B0", nil},
		{"foo bar", "", bidirule.ErrInvalid},
		{"♚", "", bidirule.ErrInvalid},
		{"\u007E", "", bidirule.ErrInvalid}, // disallowed by bidi rule
		{"a", "a", nil},
		{"!", "", bidirule.ErrInvalid}, // disallowed by bidi rule
		{"²", "", bidirule.ErrInvalid},
		{"\t", "", bidirule.ErrInvalid},
		{"\n", "", bidirule.ErrInvalid},
		{"\u26D6", "", bidirule.ErrInvalid},
		{"\u26FF", "", bidirule.ErrInvalid},
		{"\uFB00", "ff", nil}, // Side effect of case folding.
		{"\u1680", "", bidirule.ErrInvalid},
		{" ", "", bidirule.ErrInvalid},
		{"  ", "", bidirule.ErrInvalid},
		{"\u01C5", "", errDisallowedRune},
		{"\u16EE", "", errDisallowedRune},   // Nl RUNIC ARLAUG SYMBOL
		{"\u0488", "", bidirule.ErrInvalid}, // Me COMBINING CYRILLIC HUNDRED THOUSANDS SIGN
		{"\u212B", "\u00e5", nil},           // Angstrom sign, NFC -> U+00E5
		{"A\u030A", "å", nil},               // A + ring
		{"\u00C5", "å", nil},                // A with ring
		{"\u00E7", "ç", nil},                // c cedille
		{"\u0063\u0327", "ç", nil},          // c + cedille
		{"\u0158", "ř", nil},
		{"\u0052\u030C", "ř", nil},

		{"\u1E61", "\u1E61", nil}, // LATIN SMALL LETTER S WITH DOT ABOVE
		// U+1e9B: case folded.
		{"ẛ", "\u1E61", nil}, // LATIN SMALL LETTER LONG S WITH DOT ABOVE

		// Confusable characters ARE allowed and should NOT be mapped.
		{"\u0410", "\u0430", nil}, // CYRILLIC CAPITAL LETTER A

		// Full width should be mapped to the canonical decomposition.
		{"ＡＢ", "ab", nil},
		{"שc", "", bidirule.ErrInvalid}, // bidi rule

	}},
	{"UsernameCasePreserved", UsernameCasePreserved, []testCase{
		{"ABC", "ABC", nil},
		{"ＡＢ", "AB", nil},
		{"שc", "", bidirule.ErrInvalid}, // bidi rule
		{"\uFB00", "", errDisallowedRune},
		{"\u212B", "\u00c5", nil},    // Angstrom sign, NFC -> U+00E5
		{"ẛ", "", errDisallowedRune}, // LATIN SMALL LETTER LONG S WITH DOT ABOVE
	}},
}

func TestString(t *testing.T) {
	doTests(t, func(t *testing.T, p *Profile, tc testCase) {
		if e, err := p.String(tc.input); tc.err != err || e != tc.output {
			t.Errorf("got %+q (err: %v); want %+q (err: %v)", e, err, tc.output, tc.err)
		}
	})
}

func TestBytes(t *testing.T) {
	doTests(t, func(t *testing.T, p *Profile, tc testCase) {
		if e, err := p.Bytes([]byte(tc.input)); tc.err != err || string(e) != tc.output {
			t.Errorf("got %+q (err: %v); want %+q (err: %v)", string(e), err, tc.output, tc.err)
		}
	})
}

func TestAppend(t *testing.T) {
	doTests(t, func(t *testing.T, p *Profile, tc testCase) {
		if e, err := p.Append(nil, []byte(tc.input)); tc.err != err || string(e) != tc.output {
			t.Errorf("got %+q (err: %v); want %+q (err: %v)", string(e), err, tc.output, tc.err)
		}
	})
}
