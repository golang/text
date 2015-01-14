// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package colltab_test

// This file contains tests which need to import package collate, which causes
// an import cycle when done within package colltab itself.

import (
	"bytes"
	"testing"
	"unicode"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

func TestNonDigits(t *testing.T) {
	c := collate.New(language.English, collate.Loose, collate.Numeric)

	// Verify that all non-digit numbers sort outside of the number range.
	for r, hi := rune(unicode.N.R16[0].Lo), rune(unicode.N.R32[0].Hi); r <= hi; r++ {
		if unicode.In(r, unicode.Nd) {
			continue
		}
		if a := string(r); c.CompareString(a, "0") != -1 && c.CompareString(a, "999999") != 1 {
			t.Errorf("%+q non-digit number is collated as digit", a)
		}
	}
}

func TestNumericCompare(t *testing.T) {
	c := collate.New(language.English, collate.Loose, collate.Numeric)

	// Iterate over all digits.
	for _, r16 := range unicode.Nd.R16 {
		testDigitCompare(t, c, rune(r16.Lo), rune(r16.Hi))
	}
	for _, r32 := range unicode.Nd.R32 {
		testDigitCompare(t, c, rune(r32.Lo), rune(r32.Hi))
	}
}

// This implementation of collation uses an out of date Unicode table, as the
// current implemenation uses CLDR 24, which uses Unicode 6.3.0. This will
// be resolved once we move to CLDR version 25 (or later), which introduces
// significant changes.
// We skip digits that are not known as digits by the Weighter.
//
// TODO: remove once we upgrade collation to newer tables.
var notSupported = map[rune]bool{
	0x0de6:     true,
	0xa9f0:     true,
	0x000112f0: true,
	0x000114d0: true,
	0x00011650: true,
	0x000118e0: true,
	0x00016a60: true,
	0x00016b50: true,
}

func testDigitCompare(t *testing.T, c *collate.Collator, zero, nine rune) {
	if notSupported[zero] {
		return
	}
	n := int(nine - zero + 1)
	if n%10 != 0 {
		t.Fatalf("len([%+q, %+q]) = %d; want a multiple of 10", zero, nine, n)
	}
	for _, tt := range []struct {
		prefix string
		b      [11]string
	}{
		{
			prefix: "",
			b: [11]string{
				"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10",
			},
		},
		{
			prefix: "1",
			b: [11]string{
				"10", "11", "12", "13", "14", "15", "16", "17", "18", "19", "20",
			},
		},
		{
			prefix: "0",
			b: [11]string{
				"00", "01", "02", "03", "04", "05", "06", "07", "08", "09", "10",
			},
		},
		{
			prefix: "00",
			b: [11]string{
				"000", "001", "002", "003", "004", "005", "006", "007", "008", "009", "010",
			},
		},
		{
			prefix: "9",
			b: [11]string{
				"90", "91", "92", "93", "94", "95", "96", "97", "98", "99", "100",
			},
		},
	} {
		for k := 0; k <= n; k++ {
			i := k % 10
			a := tt.prefix + string(zero+rune(i))
			for j, b := range tt.b {
				want := 0
				switch {
				case i < j:
					want = -1
				case i > j:
					want = 1
				}
				got := c.CompareString(a, b)
				if got != want {
					t.Errorf("Compare(%+q, %+q) = %d; want %d", a, b, got, want)
					return
				}
			}
		}
	}
}

func BenchmarkNumericWeighter(b *testing.B) {
	c := collate.New(language.English, collate.Numeric)
	input := bytes.Repeat([]byte("Testing, testing 123..."), 100)
	b.SetBytes(int64(2 * len(input)))
	for i := 0; i < b.N; i++ {
		c.Compare(input, input)
	}
}
