// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package currency

import "testing"

func TestParseISO(t *testing.T) {
	tests := []struct {
		in  string
		out string
		ok  bool
	}{
		{"USD", "USD", true},
		{"xxx", "XXX", true},
		{"xts", "XTS", true},
		{"XX", "XXX", false},
		{"XXXX", "XXX", false},
		{"", "XXX", false},       // not well-formed
		{"UUU", "XXX", false},    // unknown
		{"\u22A9", "XXX", false}, // non-ASCII, printable

		{"aaa", "XXX", false},
		{"zzz", "XXX", false},
		{"000", "XXX", false},
		{"999", "XXX", false},
		{"---", "XXX", false},
		{"\x00\x00\x00", "XXX", false},
		{"\xff\xff\xff", "XXX", false},
	}
	for i, tc := range tests {
		if x, err := ParseISO(tc.in); x.String() != tc.out || err == nil != tc.ok {
			t.Errorf("%d:%s: was %s, %v; want %s, %v", i, tc.in, x, err == nil, tc.out, tc.ok)
		}
	}
}

var (
	czk = MustParseISO("CZK")
	zwr = MustParseISO("ZWR")
)

func TestTable(t *testing.T) {
	for i := 4; i < len(currency); i += 4 {
		if a, b := currency[i-4:i-1], currency[i:i+3]; a >= b {
			t.Errorf("currency unordered at element %d: %s >= %s", i, a, b)
		}
	}
	// First currency has index 1, last is numCurrencies.
	if c := currency.Elem(1)[:3]; c != "ADP" {
		t.Errorf("first was %c; want ADP", c)
	}
	if c := currency.Elem(numCurrencies)[:3]; c != "ZWR" {
		t.Errorf("last was %c; want ZWR", c)
	}
}

func TestKindRounding(t *testing.T) {
	tests := []struct {
		kind  Kind
		cur   Currency
		scale int
		inc   int
	}{
		{Standard, USD, 2, 1},
		{Standard, CHF, 2, 1},
		{Cash, CHF, 2, 5},
		{Standard, TWD, 2, 1},
		{Cash, TWD, 0, 1},
		{Standard, czk, 2, 1},
		{Cash, czk, 0, 1},
		{Standard, zwr, 2, 1},
		{Cash, zwr, 0, 1},
	}
	for i, tc := range tests {
		if scale, inc := tc.kind.Rounding(tc.cur); scale != tc.scale && inc != tc.inc {
			t.Errorf("%d: got %d, %d; want %d, %d", i, scale, inc, tc.scale, tc.inc)
		}
	}
}

func BenchmarkString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		USD.String()
	}
}
