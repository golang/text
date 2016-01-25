// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package precis

import (
	"testing"
)

func TestEnforce(t *testing.T) {
	var data = []struct {
		prof          Profile
		input, output string
		isErr         bool
	}{
		// Nickname profile
		{Nickname, "  Swan  of   Avon   ", "Swan of Avon", false},
		{Nickname, "", "", true},
		{Nickname, " ", "", true},
		{Nickname, "  ", "", true},
		{Nickname, "a\u00A0a\u1680a\u2000a\u2001a\u2002a\u2003a\u2004a\u2005a\u2006a\u2007a\u2008a\u2009a\u200Aa\u202Fa\u205Fa\u3000a", "a a a a a a a a a a a a a a a a a", false},
		{Nickname, "Foo", "Foo", false},
		{Nickname, "foo", "foo", false},
		{Nickname, "Foo Bar", "Foo Bar", false},
		{Nickname, "foo bar", "foo bar", false},
		{Nickname, "\u03C3", "\u03C3", false},
		// TODO: Figure out why this is failing.
		// {Nickname, "\u03C2", "\u03C3", false},
		{Nickname, "\u265A", "♚", false},
		{Nickname, "Richard \u2163", "Richard IV", false},
		{Nickname, "\u212B", "Å", false},

		// UsernameCaseMapped profile
		{UsernameCaseMapped, "juliet@example.com", "juliet@example.com", false},
		{UsernameCaseMapped, "fussball", "fussball", false},
		{UsernameCaseMapped, "fu\u00DFball", "fussball", false},
		{UsernameCaseMapped, "\u03C0", "\u03C0", false},
		{UsernameCaseMapped, "\u03A3", "\u03C3", false},
		{UsernameCaseMapped, "\u03C3", "\u03C3", false},
		{UsernameCaseMapped, "\u03C2", "\u03C3", false},
		{UsernameCaseMapped, "\u0049", "\u0069", false},
		{UsernameCaseMapped, "\u0049", "\u0069", false},
		// TODO: Should this be disallowed?
		// {UsernameCaseMapped, "\u03D2", "\u03C5", false},
		{UsernameCaseMapped, "\u03B0", "\u03B0", false},
		{UsernameCaseMapped, "foo bar", "", true},
		{UsernameCaseMapped, "♚", "", true},
		{UsernameCaseMapped, "\u007E", "\u007E", false},
		{UsernameCaseMapped, "a", "a", false},
		{UsernameCaseMapped, "!", "!", false},
		{UsernameCaseMapped, "²", "", true},
		// TODO: Should this work?
		// {UsernameCaseMapped, "", "", true},
		{UsernameCaseMapped, "\t", "", true},
		{UsernameCaseMapped, "\n", "", true},
		{UsernameCaseMapped, "\u26D6", "", true},
		{UsernameCaseMapped, "\u26FF", "", true},
		{UsernameCaseMapped, "\uFB00", "", true},
		{UsernameCaseMapped, "\u1680", "", true},
		{UsernameCaseMapped, " ", "", true},
		{UsernameCaseMapped, "  ", "", true},
		{UsernameCaseMapped, "\u01C5", "", true},
		{UsernameCaseMapped, "\u16EE", "", true}, // Nl RUNIC ARLAUG SYMBOL
		{UsernameCaseMapped, "\u0488", "", true}, // Me COMBINING CYRILLIC HUNDRED THOUSANDS SIGN
		// TODO: Should this be disallowed and/or case mapped?
		// {UsernameCaseMapped, "\u212B", "å", false}, // angstrom sign
		{UsernameCaseMapped, "A\u030A", "å", false},      // A + ring
		{UsernameCaseMapped, "\u00C5", "å", false},       // A with ring
		{UsernameCaseMapped, "\u00E7", "ç", false},       // c cedille
		{UsernameCaseMapped, "\u0063\u0327", "ç", false}, // c + cedille
		{UsernameCaseMapped, "\u0158", "ř", false},
		{UsernameCaseMapped, "\u0052\u030C", "ř", false},

		{UsernameCaseMapped, "\u1E61", "\u1E61", false}, // LATIN SMALL LETTER S WITH DOT ABOVE
		// TODO: Why is this disallowed?
		// {UsernameCaseMapped, "ẛ", "\u1E61", false}, // LATIN SMALL LETTER LONG S WITH DOT ABOVE

		// Confusable characters ARE allowed and should NOT be mapped.
		{UsernameCaseMapped, "\u0410", "\u0430", false}, // CYRILLIC CAPITAL LETTER A

		// Full width should be mapped to the narrow or canonical decomposition… no
		// idea which, but either way in this case it should be the same:
		{UsernameCaseMapped, "ＡＢ", "ab", false},

		{UsernameCasePreserved, "ABC", "ABC", false},
		{UsernameCasePreserved, "ＡＢ", "AB", false},
	}

	for _, d := range data {
		if e, err := d.prof.String(d.input); (d.isErr && err == nil) ||
			!d.isErr && (err != nil || e != d.output) {
			t.Log("Expected '"+d.output+"'", "but got", "'"+e+"'", "with error:", err)
			t.Fail()
		}
	}
}
