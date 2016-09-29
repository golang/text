// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package precis

import (
	"fmt"
	"testing"

	"golang.org/x/text/internal/testtext"
)

type compareTestCase struct {
	a      string
	b      string
	result bool
}

var compareTestCases = []struct {
	name  string
	p     *Profile
	cases []compareTestCase
}{
	{"Nickname", Nickname, []compareTestCase{
		{"a", "b", false},
		{"  Swan  of   Avon   ", "swan of avon", true},
		{"Foo", "foo", true},
		{"foo", "foo", true},
		{"Foo Bar", "foo bar", true},
		{"foo bar", "foo bar", true},
		{"\u03A3", "\u03C3", true},
		{"\u03A3", "\u03C2", false},
		{"\u03C3", "\u03C2", false},
		{"Richard \u2163", "richard iv", true},
		{"Å", "å", true},
		{"ﬀ", "ff", true}, // because of NFKC
		{"ß", "sS", false},
	}},
}

func TestCompare(t *testing.T) {
	for _, g := range compareTestCases {
		for i, tc := range g.cases {
			name := fmt.Sprintf("%s:%d:%+q", g.name, i, tc.a)
			testtext.Run(t, name, func(t *testing.T) {
				if result := g.p.Compare(tc.a, tc.b); result != tc.result {
					t.Errorf("got %v; want %v", result, tc.result)
				}
			})
		}
	}
}
