// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package language

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTagSize(t *testing.T) {
	id := Tag{}
	typ := reflect.TypeOf(id)
	if typ.Size() > 16 {
		t.Errorf("size of Tag was %d; want 16", typ.Size())
	}
}

func TestIsRoot(t *testing.T) {
	loc := Tag{}
	if !loc.IsRoot() {
		t.Errorf("unspecified should be root.")
	}
	for i, tt := range parseTests() {
		loc, _ := Parse(tt.in)
		undef := tt.lang == "und" && tt.script == "" && tt.region == "" && tt.ext == ""
		if loc.IsRoot() != undef {
			fmt.Printf("%d: %#v %s %s\n", i, loc, tt.lang, undef)
			t.Errorf("%d: was %v; want %v", i, loc.IsRoot(), undef)
		}
	}
}

func TestMakeString(t *testing.T) {
	tests := []struct{ in, out string }{
		{"und", "und"},
		{"und", "und-CW"},
		{"nl", "nl-NL"},
		{"de-1901", "nl-1901"},
		{"de-1901", "de-Arab-1901"},
		{"x-a-b", "de-Arab-x-a-b"},
		{"x-a-b", "x-a-b"},
	}
	for i, tt := range tests {
		id, _ := Parse(tt.in)
		mod, _ := Parse(tt.out)
		id.setTagsFrom(mod)
		for j := 0; j < 2; j++ {
			id.remakeString()
			if str := id.String(); str != tt.out {
				t.Errorf("%d:%d: found %s; want %s", i, j, id.String(), tt.out)
			}
		}
	}
}

func TestBase(t *testing.T) {
	tests := []struct {
		loc, lang string
		conf      Confidence
	}{
		{"und", "en", Low},
		{"x-abc", "und", No},
		{"en", "en", Exact},
		{"und-Cyrl", "ru", High},
		// If a region is not included, the official language should be English.
		{"und-US", "en", High},
		// TODO: not-explicitly listed scripts should probably be und, No
		// Modify addTags to return info on how the match was derived.
		// {"und-Aghb", "und", No},
	}
	for i, tt := range tests {
		loc, _ := Parse(tt.loc)
		loc, _ = loc.Canonicalize(0)
		lang, conf := loc.Base()
		if lang.String() != tt.lang {
			t.Errorf("%d: language was %s; want %s", i, lang, tt.lang)
		}
		if conf != tt.conf {
			t.Errorf("%d: confidence was %d; want %d", i, conf, tt.conf)
		}
	}
}

func TestParseBase(t *testing.T) {
	tests := []struct {
		in  string
		out string
		ok  bool
	}{
		{"en", "en", true},
		{"EN", "en", true},
		{"nld", "nl", true},
		{"aaj", "und", false}, // unknown
		{"qaa", "qaa", true},
		{"a", "und", false},
		{"", "und", false},
		{"aaaa", "und", false},
	}
	for i, tt := range tests {
		x, err := ParseBase(tt.in)
		if x.String() != tt.out || err == nil != tt.ok {
			t.Errorf("%d:%s: was %s, %v; want %s, %v", i, tt.in, x, err == nil, tt.out, tt.ok)
		}
		tag, _ := Parse(tt.out)
		if err == nil && !tag.equalTags(x.Tag()) {
			t.Errorf("%d:%s: tag was %s; want %s", i, tt.in, x.Tag(), tt.out)
		}
	}
}

func TestScript(t *testing.T) {
	tests := []struct {
		loc, scr string
		conf     Confidence
	}{
		{"und", "Latn", Low},
		{"en-Latn", "Latn", Exact},
		{"en", "Latn", High},
		{"sr", "Cyrl", Low},
		{"cmn", "Hans", Low},
		{"ru", "Cyrl", High},
		{"yue", "Zyyy", No},
		{"x-abc", "Zyyy", Low},
	}
	for i, tt := range tests {
		loc, _ := Parse(tt.loc)
		loc, _ = loc.Canonicalize(0)
		sc, conf := loc.Script()
		if sc.String() != tt.scr {
			t.Errorf("%d: script was %s; want %s", i, sc, tt.scr)
		}
		if conf != tt.conf {
			t.Errorf("%d: confidence was %d; want %d", i, conf, tt.conf)
		}
	}
}

func TestParseScript(t *testing.T) {
	tests := []struct {
		in  string
		out string
		ok  bool
	}{
		{"Latn", "Latn", true},
		{"zzzz", "Zzzz", true},
		{"Latm", "Zyyy", false},
		{"Zzz", "Zyyy", false},
		{"", "Zyyy", false},
		{"Zzzxx", "Zyyy", false},
	}
	for i, tt := range tests {
		x, err := ParseScript(tt.in)
		if x.String() != tt.out || err == nil != tt.ok {
			t.Errorf("%d:%s: was %s, %v; want %s, %v", i, tt.in, x, err == nil, tt.out, tt.ok)
		}
		tag, _ := Parse("und-" + tt.in)
		if err == nil && !tag.equalTags(x.Tag()) {
			t.Errorf("%d:%s: tag was %s; want %s", i, tt.in, x.Tag(), tt.out)
		}
	}
}

func TestRegion(t *testing.T) {
	tests := []struct {
		loc, reg string
		conf     Confidence
	}{
		{"und", "US", Low},
		{"en", "US", Low},
		{"zh-Hant", "TW", Low},
		{"en-US", "US", Exact},
		{"cmn", "CN", Low},
		{"ru", "RU", Low},
		{"yue", "ZZ", No},
		{"x-abc", "ZZ", Low},
	}
	for i, tt := range tests {
		loc, _ := Parse(tt.loc)
		loc, _ = loc.Canonicalize(0)
		reg, conf := loc.Region()
		if reg.String() != tt.reg {
			t.Errorf("%d: region was %s; want %s", i, reg, tt.reg)
		}
		if conf != tt.conf {
			t.Errorf("%d: confidence was %d; want %d", i, conf, tt.conf)
		}
	}
}

func TestEncodeM49(t *testing.T) {
	tests := []struct {
		m49  int
		code string
		ok   bool
	}{
		{1, "001", true},
		{840, "US", true},
		{899, "ZZ", false},
	}
	for i, tt := range tests {
		if r, err := EncodeM49(tt.m49); r.String() != tt.code || err == nil != tt.ok {
			t.Errorf("%d:%d: was %s, %v; want %s, %v", i, tt.m49, r, err == nil, tt.code, tt.ok)
		}
	}
}

func TestParseRegion(t *testing.T) {
	tests := []struct {
		in  string
		out string
		ok  bool
	}{
		{"001", "001", true},
		{"840", "US", true},
		{"899", "ZZ", false},
		{"USA", "US", true},
		{"US", "US", true},
		{"BC", "ZZ", false},
		{"C", "ZZ", false},
		{"CCCC", "ZZ", false},
		{"01", "ZZ", false},
	}
	for i, tt := range tests {
		r, err := ParseRegion(tt.in)
		if r.String() != tt.out || err == nil != tt.ok {
			t.Errorf("%d:%s: was %s, %v; want %s, %v", i, tt.in, r, err == nil, tt.out, tt.ok)
		}
		tag, _ := Parse("und-" + tt.out)
		if err == nil && !tag.equalTags(r.Tag()) {
			t.Errorf("%d:%s: tag was %s; want %s", i, tt.in, r.Tag(), tag)
		}
	}
}

func TestIsCountry(t *testing.T) {
	tests := []struct {
		reg     string
		country bool
	}{
		{"US", true},
		{"001", false},
		{"958", false},
		{"419", false},
		{"203", true},
		{"020", true},
		{"900", false},
		{"999", false},
		{"QO", false},
		{"EU", false},
		{"AA", false},
	}
	for i, tt := range tests {
		reg, _ := getRegionID([]byte(tt.reg))
		r := Region{reg}
		if r.IsCountry() != tt.country {
			t.Errorf("%d: IsCountry(%s) was %v; want %v", i, tt.reg, r.IsCountry(), tt.country)
		}
	}
}

func TestParseCurrency(t *testing.T) {
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
		{"", "XXX", false},
		{"UUU", "XXX", false}, // unknown
	}
	for i, tt := range tests {
		if x, err := ParseCurrency(tt.in); x.String() != tt.out || err == nil != tt.ok {
			t.Errorf("%d:%s: was %s, %v; want %s, %v", i, tt.in, x, err == nil, tt.out, tt.ok)
		}
	}
}

func TestCanonicalize(t *testing.T) {
	// TODO: do a full test using CLDR data in a separate regression test.
	tests := []struct {
		in, out string
		option  CanonType
	}{
		{"en-Latn", "en", SuppressScript},
		{"sr-Cyrl", "sr-Cyrl", SuppressScript},
		{"sh", "sr-Latn", Legacy},
		{"sh-HR", "sr-Latn-HR", Legacy},
		{"sh-Cyrl-HR", "sr-Cyrl-HR", Legacy},
		{"tl", "fil", Legacy},
		{"no", "no", Legacy},
		{"no", "nb", Legacy | CLDR},
		{"cmn", "cmn", Legacy},
		{"cmn", "zh", Macro},
		{"yue", "yue", Macro},
		{"nb", "no", Macro},
		{"nb", "nb", Macro | CLDR},
		{"no", "no", Macro},
		{"no", "no", Macro | CLDR},
		{"iw", "he", Deprecated},
		{"iw", "he", Deprecated | CLDR},
		{"mo", "ro-MD", Deprecated},
		{"mo", "ro", Deprecated | CLDR},
	}
	for i, tt := range tests {
		in, _ := Parse(tt.in)
		in, _ = in.Canonicalize(tt.option)
		if in.String() != tt.out {
			t.Errorf("%d:%s: was %s; want %s", i, tt.in, in.String(), tt.out)
		}
	}
}

var (
	// Tags without error that don't need to be changed.
	benchBasic = []string{
		"en",
		"en-Latn",
		"en-GB",
		"za",
		"zh-Hant",
		"zh",
		"zh-HK",
		"ar-MK",
		"en-CA",
		"fr-CA",
		"fr-CH",
		"fr",
		"lv",
		"he-IT",
		"tlh",
		"ja",
		"ja-Jpan",
		"ja-Jpan-JP",
		"de-1996",
		"de-CH",
		"sr",
		"sr-Latn",
	}
	// Tags with extensions, not changes required.
	benchExt = []string{
		"x-a-b-c-d",
		"x-aa-bbbb-cccccccc-d",
		"en-x_cc-b-bbb-a-aaa",
		"en-c_cc-b-bbb-a-aaa-x-x",
		"en-u-co-phonebk",
		"en-Cyrl-u-co-phonebk",
		"en-US-u-co-phonebk-cu-xau",
		"en-nedix-u-co-phonebk",
		"en-t-t0-abcd",
		"en-t-nl-latn",
		"en-t-t0-abcd-x-a",
	}
	// Change, but not memory allocation required.
	benchSimpleChange = []string{
		"EN",
		"i-klingon",
		"en-latn",
		"zh-cmn-Hans-CN",
		"iw-NL",
	}
	// Change and memory allocation required.
	benchChangeAlloc = []string{
		"en-c_cc-b-bbb-a-aaa",
		"en-u-cu-xua-co-phonebk",
		"en-u-cu-xua-co-phonebk-a-cd",
		"en-u-def-abc-cu-xua-co-phonebk",
		"en-t-en-Cyrl-NL-1994",
		"en-t-en-Cyrl-NL-1994-t0-abc-def",
	}
	// Tags that result in errors.
	benchErr = []string{
		// IllFormed
		"x_A.-B-C_D",
		"en-u-cu-co-phonebk",
		"en-u-cu-xau-co",
		"en-t-nl-abcd",
		// Invalid
		"xx",
		"nl-Uuuu",
		"nl-QB",
	}
	benchChange = append(benchSimpleChange, benchChangeAlloc...)
	benchAll    = append(append(append(benchBasic, benchExt...), benchChange...), benchErr...)
)

func doParse(b *testing.B, tag []string) {
	for i := 0; i < b.N; i++ {
		// Use the modulo instead of looping over all tags so that we get a somewhat
		// meaningful ns/op.
		Parse(tag[i%len(tag)])
	}
}

func BenchmarkParse(b *testing.B) {
	doParse(b, benchAll)
}

func BenchmarkParseBasic(b *testing.B) {
	doParse(b, benchBasic)
}

func BenchmarkParseError(b *testing.B) {
	doParse(b, benchErr)
}

func BenchmarkParseSimpleChange(b *testing.B) {
	doParse(b, benchSimpleChange)
}

func BenchmarkParseChangeAlloc(b *testing.B) {
	doParse(b, benchChangeAlloc)
}
