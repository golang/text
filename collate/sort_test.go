// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package collate_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

func ExampleCollator_strings() {
	c := collate.New(language.Und)
	strings := []string{
		"ad",
		"ab",
		"äb",
		"ac",
	}
	c.SortStrings(strings)
	fmt.Println(strings)
	// Output: [ab äb ac ad]
}

type sorter []string

func (s sorter) Len() int {
	return len(s)
}

func (s sorter) Swap(i, j int) {
	s[j], s[i] = s[i], s[j]
}

func (s sorter) Bytes(i int) []byte {
	return []byte(s[i])
}

func TestSort(t *testing.T) {
	c := collate.New(language.English)
	strings := []string{
		"bcd",
		"abc",
		"ddd",
	}
	c.Sort(sorter(strings))
	res := fmt.Sprint(strings)
	want := "[abc bcd ddd]"
	if res != want {
		t.Errorf("found %s; want %s", res, want)
	}
}

func TestSortStringsAndCompareString(t *testing.T) {
	for _, tt := range []struct {
		name string
		c    *collate.Collator
		want []string
	}{
		{
			name: "English default options",
			c:    collate.New(language.English),
			want: []string{
				"abc",
				"bcd",
				"ddd",
			},
		},
		{
			// From https://www.unicode.org/reports/tr10/#Variable_Weighting_Examples
			name: "Blanked",
			c:    collate.New(language.MustParse("en-us-u-ka-blanked")),
			want: []string{
				"death",
				"de luge",
				"de-luge",
				"deluge",
				"de-luge",
				"de Luge",
				"de-Luge",
				"deLuge",
				"de-Luge",
				"demark",
			},
		},
		{
			// From https://www.unicode.org/reports/tr10/#Variable_Weighting_Examples
			name: "Shifted",
			c:    collate.New(language.MustParse("en-us-u-ka-shifted")),
			want: []string{
				"death",
				"de luge",
				"de-luge",
				"de-luge",
				"deluge",
				"de Luge",
				"de-Luge",
				"de-Luge",
				"deLuge",
				"demark",
			},
		},
		{
			// From https://www.unicode.org/reports/tr10/#Variable_Weighting_Examples
			name: "Shift-Trimmed",
			c:    collate.New(language.MustParse("en-us-u-ka-posix-ks-level4")),
			want: []string{
				"death",
				"deluge",
				"de luge",
				"de-luge",
				"de-luge",
				"deLuge",
				"de Luge",
				"de-Luge",
				"de-Luge",
				"demark",
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			actual := make([]string, len(tt.want))
			copy(actual, tt.want)
			tt.c.SortStrings(actual)

			p := func(v []string) string { return strings.Join(v, ", ") }
			if p(tt.want) != p(actual) {
				t.Errorf("SortStrings want: '%v'\n Got: '%v'", p(tt.want), p(actual))
			}

			buf := collate.Buffer{}
			for i := 0; i < len(tt.want)-1; i++ {
				a, b := tt.want[i], tt.want[i+1]
				kA, kB := tt.c.KeyFromString(&buf, a), tt.c.KeyFromString(&buf, b)
				if bytes.Compare(kA, kB) > 0 {
					t.Errorf("KeyFromString for %v is bigger than for %v", a, b)
				}
			}

			for i := 0; i < len(tt.want)-1; i++ {
				a, b := tt.want[i], tt.want[i+1]
				cmp := tt.c.CompareString(a, b)
				if cmp > 0 {
					t.Errorf("CompareString for '%v' vs '%v' is 1 when should be -1 or 0", a, b)
				}
			}
		})
	}
}
