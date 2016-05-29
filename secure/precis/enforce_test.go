// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package precis

import "testing"

type testCase struct {
	input, output string
	isErr         bool
}

var testCases = []struct {
	name  string
	p     *Profile
	cases []testCase
}{
	{"Nickname", Nickname, []testCase{
		{"  Swan  of   Avon   ", "Swan of Avon", false},
		{"", "", true},
		{" ", "", true},
		{"  ", "", true},
		{"a\u00A0a\u1680a\u2000a\u2001a\u2002a\u2003a\u2004a\u2005a\u2006a\u2007a\u2008a\u2009a\u200Aa\u202Fa\u205Fa\u3000a", "a a a a a a a a a a a a a a a a a", false},
		{"Foo", "Foo", false},
		{"foo", "foo", false},
		{"Foo Bar", "Foo Bar", false},
		{"foo bar", "foo bar", false},
		{"\u03C3", "\u03C3", false},
		// Greek final sigma is left as is (do not fold!)
		{"\u03C2", "\u03C2", false},
		{"\u265A", "♚", false},
		{"Richard \u2163", "Richard IV", false},
		{"\u212B", "Å", false},
		{"\uFB00", "ff", false}, // because of NFKC
		{"שa", "שa", false},     // no bidi rule
		{"동일조건변경허락", "동일조건변경허락", false},
	}},
	{"OpaqueString", OpaqueString, []testCase{
		{"  Swan  of   Avon   ", "  Swan  of   Avon   ", false},
		{"", "", true},
		{" ", " ", false},
		{"  ", "  ", false},
		{"a\u00A0a\u1680a\u2000a\u2001a\u2002a\u2003a\u2004a\u2005a\u2006a\u2007a\u2008a\u2009a\u200Aa\u202Fa\u205Fa\u3000a", "a a a a a a a a a a a a a a a a a", false},
		{"Foo", "Foo", false},
		{"foo", "foo", false},
		{"Foo Bar", "Foo Bar", false},
		{"foo bar", "foo bar", false},
		{"\u03C3", "\u03C3", false},
		{"Richard \u2163", "Richard \u2163", false},
		{"\u212B", "Å", false},
		{"Jack of \u2666s", "Jack of \u2666s", false},
		{"my cat is a \u0009by", "", true},
		{"·", "", true}, // Middle dot
		{"͵", "", true}, // Keraia
		{"׳", "", true},
		{"׳ה", "", true},
		{"a׳b", "", true},
		// TODO: Requires context rule to work.
		// {OpaqueString, "ש׳", "ש", false},  // U+05e9 U+05f3
		{"שa", "שa", false}, // no bidi rule

		// Katakana Middle Dot
		{"abc・def", "", true},
		// TODO: These require context rules to work.
		// {OpaqueString, "aヅc・def", "", false},
		// {OpaqueString, "abc・dぶf", "", false},
		// {OpaqueString, "⺐bc・def", "", false},

		// Arabic Indic Digit
		// TODO: These require context rules to work.
		// {OpaqueString, "١٢٣٤٥", "١٢٣٤٥", false},
		// {OpaqueString, "۱۲۳۴۵", "۱۲۳۴۵", false},
		{"١٢٣٤٥۶", "", true},
		{"۱۲۳۴۵٦", "", true},
	}},
	{"UsernameCaseMapped", UsernameCaseMapped, []testCase{
		// TODO: Should this work?
		// {UsernameCaseMapped, "", "", true},
		{"juliet@example.com", "juliet@example.com", false},
		{"fussball", "fussball", false},
		{"fu\u00DFball", "fussball", false},
		{"\u03C0", "\u03C0", false},
		{"\u03A3", "\u03C3", false},
		{"\u03C3", "\u03C3", false},
		{"\u03C2", "\u03C3", false},
		{"\u0049", "\u0069", false},
		{"\u0049", "\u0069", false},
		{"\u03D2", "\u03C5", true},
		{"\u03B0", "\u03B0", false},
		{"foo bar", "", true},
		{"♚", "", true},
		{"\u007E", "", true}, // disallowed by bidi rule
		{"a", "a", false},
		{"!", "", true}, // disallowed by bidi rule
		{"²", "", true},
		{"\t", "", true},
		{"\n", "", true},
		{"\u26D6", "", true},
		{"\u26FF", "", true},
		{"\uFB00", "ff", false}, // Side effect of case folding.
		{"\u1680", "", true},
		{" ", "", true},
		{"  ", "", true},
		{"\u01C5", "", true},
		{"\u16EE", "", true},         // Nl RUNIC ARLAUG SYMBOL
		{"\u0488", "", true},         // Me COMBINING CYRILLIC HUNDRED THOUSANDS SIGN
		{"\u212B", "\u00e5", false},  // Angstrom sign, NFC -> U+00E5
		{"A\u030A", "å", false},      // A + ring
		{"\u00C5", "å", false},       // A with ring
		{"\u00E7", "ç", false},       // c cedille
		{"\u0063\u0327", "ç", false}, // c + cedille
		{"\u0158", "ř", false},
		{"\u0052\u030C", "ř", false},

		{"\u1E61", "\u1E61", false}, // LATIN SMALL LETTER S WITH DOT ABOVE
		// U+1e9B: case folded.
		{"ẛ", "\u1E61", false}, // LATIN SMALL LETTER LONG S WITH DOT ABOVE

		// Confusable characters ARE allowed and should NOT be mapped.
		{"\u0410", "\u0430", false}, // CYRILLIC CAPITAL LETTER A

		// Full width should be mapped to the canonical decomposition.
		{"ＡＢ", "ab", false},
		{"שc", "שc", true}, // bidi rule

	}},
	{"UsernameCasePreserved", UsernameCasePreserved, []testCase{
		{"ABC", "ABC", false},
		{"ＡＢ", "AB", false},
		{"שc", "שc", true}, // bidi rule
		{"\uFB00", "", true},
		{"\u212B", "\u00c5", false}, // Angstrom sign, NFC -> U+00E5
		{"ẛ", "", true},             // LATIN SMALL LETTER LONG S WITH DOT ABOVE
	}},
}

func TestEnforce(t *testing.T) {
	doTests(t, func(t *testing.T, p *Profile, tc testCase) {
		if e, err := p.String(tc.input); (tc.isErr && err == nil) ||
			!tc.isErr && (err != nil || e != tc.output) {
			t.Errorf("got %+q (err: %v); want %+q (ok: %v)", e, err, tc.output, !tc.isErr)
		}
	})
}
