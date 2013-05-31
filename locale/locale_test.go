// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package locale

import (
	"reflect"
	"testing"
)

func TestIDSize(t *testing.T) {
	id := ID{}
	typ := reflect.TypeOf(id)
	if typ.Size() > 16 {
		t.Errorf("size of ID was %d; want 16", typ.Size())
	}
}

func TestIsRoot(t *testing.T) {
	for i, tt := range parseTests() {
		loc, _ := Parse(tt.in)
		undef := tt.lang == "und" && tt.script == "" && tt.region == "" && tt.ext == ""
		if loc.IsRoot() != undef {
			t.Errorf("%d: was %v; want %v", i, loc.IsRoot(), undef)
		}
	}
}

/*
func TestParent(t *testing.T) {
	tests := []struct {
		in, out string
	}{
		{"und", "und"},
		{"de-1994", "de"},
		{"de-CH-1994", "de-CH"},
		{"de-Cyrl-CH-1994", "de-Cyrl-CH"},
		{"zh", "und"},
		{"zh-HK-u-cu-usd", "zh"},
		{"zh-Hans-HK-u-cu-usd", "zh-Hans"},
		{"zh-u-cu-usd", "und"},
		{"zh_Hans", "zh"},
		{"zh_Hant", "und"},
		{"vai", "und"},
		{"vai_Latn", "und"},
		{"nl_Cyrl", "nl"},
		{"nl", "und"},
		{"en_US", "en"},
		{"en_150", "en-GB"},
		{"en-SG", "en-GB"},
		{"en_GB", "en"},
	}
	for i, tt := range tests {
		test, _ := Parse(tt.in)
		gold, _ := Parse(tt.out)
		if p := test.Parent(); p.String() != gold.String() {
			t.Errorf("%d:parent(%q): found %s; want %s", i, tt.in, p.String(), tt.out)
		}
	}
}

func TestWritten(t *testing.T) {
	tests := []struct {
		in, out string
	}{
		{"und", "und"},
		{"zh-Hans", "zh"},
		{"zh-Hant", "zh-Hant"},
		{"vai", "vai"},
		{"vai-Latn", "vai-Latn"},
		{"nl-Cyrl", "nl-Cyrl"},
		{"en-US", "en"},
		{"en-150", "en"},
		{"en-SG", "en"},
		{"en-GB", "en"},
	}
	for i, tt := range tests {
		test, _ := Parse(tt.in)
		gold, _ := Parse(tt.out)
		if test.Written() != gold {
			t.Errorf("%d:parent(%q): found %s; want %s", i, tt.in, test.String(), tt.out)
		}
	}
}
*/
