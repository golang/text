// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package display

import (
	"reflect"
	"testing"
	"unicode"

	"golang.org/x/text/language"
)

// TODO: test that tables are properly dropped by the linker for various use
// cases.

var (
	firstLang2aa  = language.MustParseBase("aa")
	lastLang2zu   = language.MustParseBase("zu")
	firstLang3ace = language.MustParseBase("ace")
	lastLang3zza  = language.MustParseBase("zza")
	firstTagAr001 = language.MustParse("ar-001")
	lastTagZhHant = language.MustParse("zh-Hant")
)

func TestSupported(t *testing.T) {
	supportedTags := Supported.Tags()
	if len(supportedTags) != numSupported {
		t.Errorf("number of supported was %d; want %d", len(supportedTags), numSupported)
	}

	tags := make(map[language.Tag]bool)
	namers := make(map[Namer]bool)
	// isNil verifies that the namer is unique and returns whether it is nil.
	isNil := func(n Namer) bool {
		if n != nil {
			if namers[n] {
				t.Errorf("%s: duplicate namer", n)
			}
			namers[n] = true
		}
		return n == nil
	}

	for _, tag := range supportedTags {
		if isNil(Languages(tag)) && isNil(Regions(tag)) && isNil(Scripts(tag)) {
			t.Errorf("%s: supported, but no data available", tag)
		}
		if tags[tag] {
			t.Errorf("%s: included in Supported.Tags more than once", tag)
		}
		tags[tag] = true
	}
}

func TestCoverage(t *testing.T) {
	en := language.English
	tests := []struct {
		n Namer
		x interface{}
	}{
		{Languages(en), Values.Tags()},
		{Scripts(en), Values.Scripts()},
		{Regions(en), Values.Regions()},
	}
	for i, tt := range tests {
		uniq := make(map[string]interface{})

		v := reflect.ValueOf(tt.x)
		for j := 0; j < v.Len(); j++ {
			x := v.Index(j).Interface()
			s := tt.n.Name(x)
			if s == "" {
				t.Errorf("%d:%d:%s: missing content", i, j, x)
			} else if uniq[s] != nil {
				t.Errorf("%d:%d:%s: identical return value %q for %v and %v", i, j, x, s, x, uniq[s])
			}
			uniq[s] = x
		}
	}
}

// TestUpdate tests whether dictionary entries for certain languages need to be
// updated. For some languages, some of the headers may be empty or they may be
// identical to the parent. This code detects if such entries need to be updated
// after a table update.
func TestUpdate(t *testing.T) {
	tests := []struct {
		d   *Dictionary
		tag string
	}{
		{ModernStandardArabic, "ar-001"},
		{AmericanEnglish, "en-US"},
		{EuropeanSpanish, "es-ES"},
		{BrazilianPortuguese, "pt-BR"},
		{SimplifiedChinese, "zh-Hans"},
	}

	for _, tt := range tests {
		_, i, _ := matcher.Match(language.MustParse(tt.tag))
		if !reflect.DeepEqual(tt.d.lang, langHeaders[i]) {
			t.Errorf("%s: lang table update needed", tt.tag)
		}
		if !reflect.DeepEqual(tt.d.script, scriptHeaders[i]) {
			t.Errorf("%s: script table update needed", tt.tag)
		}
		if !reflect.DeepEqual(tt.d.region, regionHeaders[i]) {
			t.Errorf("%s: region table update needed", tt.tag)
		}
	}
}

func TestIndex(t *testing.T) {
	notIn := []string{"aa", "xx", "zz", "aaa", "xxx", "zzz", "Aaaa", "Xxxx", "Zzzz"}
	tests := []tagIndex{
		{
			"",
			"",
			"",
		},
		{
			"bb",
			"",
			"",
		},
		{
			"",
			"bbb",
			"",
		},
		{
			"",
			"",
			"Bbbb",
		},
		{
			"bb",
			"bbb",
			"Bbbb",
		},
		{
			"bbccddyy",
			"bbbcccdddyyy",
			"BbbbCcccDdddYyyy",
		},
	}
	for i, tt := range tests {
		// Create the test set from the tagIndex.
		cnt := 0
		for sz := 2; sz <= 4; sz++ {
			a := tt[sz-2]
			for j := 0; j < len(a); j += sz {
				s := a[j : j+sz]
				if idx := tt.index(s); idx != cnt {
					t.Errorf("%d:%s: index was %d; want %d", i, s, idx, cnt)
				}
				cnt++
			}
		}
		if n := tt.len(); n != cnt {
			t.Errorf("%d: len was %d; want %d", i, n, cnt)
		}
		for _, x := range notIn {
			if idx := tt.index(x); idx != -1 {
				t.Errorf("%d:%s: index was %d; want -1", i, x, idx)
			}
		}
	}
}

func TestTag(t *testing.T) {
	tests := []struct {
		dict string
		tag  string
		name string
	}{
		{"nl", "nl", "Nederlands"},
		{"nl", "nl-BE", "Vlaams"},
		{"en", "en", "English"},
		{"en", "en-GB", "British English"},
		{"en", "en-US", "American English"}, // American English in CLDR 24+
		{"ru", "ru", "русский"},
		{"ru", "ru-RU", "русский (Россия)"},
		{"ru", "ru-Cyrl", "русский (Кириллица)"},
		{"en", lastLang2zu.String(), "Zulu"},
		{"en", firstLang2aa.String(), "Afar"},
		{"en", lastLang3zza.String(), "Zaza"},
		{"en", firstLang3ace.String(), "Achinese"},
		{"en", firstTagAr001.String(), "Modern Standard Arabic"},
		{"en", lastTagZhHant.String(), "Traditional Chinese"},
		{"en", "aaa", ""},
		{"en", "zzj", ""},
		// If full tag doesn't match, try without script or retion.
		{"en", "aa-Hans", "Afar (Simplified Han)"},
		{"en", "af-Arab", "Afrikaans (Arabic)"},
		{"en", "zu-Cyrl", "Zulu (Cyrillic)"},
		{"en", "aa-GB", "Afar (United Kingdom)"},
		{"en", "af-NA", "Afrikaans (Namibia)"},
		{"en", "zu-BR", "Zulu (Brazil)"},
		// Correct inheritance and language selection.
		{"zh", "zh-TW", "中文 (台湾)"},
		{"zh", "zh-Hant-TW", "繁体中文 (台湾)"},
		{"zh-Hant", "zh-TW", "中文 (台灣)"},
		{"zh-Hant", "zh-Hant-TW", "繁體中文 (台灣)"},
		// Some rather arbitrary interpretations for Serbian. This is arguably
		// correct and consistent with the way zh-[Hant-]TW is handled. It will
		// also give results more in line with the expectations if users
		// explicitly use "sh".
		{"sr-Latn", "sr-ME", "Srpski (Crna Gora)"},
		{"sr-Latn", "sr-Latn-ME", "Srpskohrvatski (Crna Gora)"},
		// Double script and region
		{"nl", "en-Cyrl-BE", "Engels (Cyrillisch, België)"},
		// Canonical equivalents.
		{"ro", "ro-MD", "moldovenească"},
		{"ro", "mo", "moldovenească"},
	}
	for i, tt := range tests {
		d := Tags(language.MustParse(tt.dict))
		if n := d.Name(language.Raw.MustParse(tt.tag)); n != tt.name {
			t.Errorf("%d:%s:%s: was %q; want %q", i, tt.dict, tt.tag, n, tt.name)
		}
	}
}

func TestLanguage(t *testing.T) {
	tests := []struct {
		dict string
		tag  string
		name string
	}{
		{"nl", "nl", "Nederlands"},
		{"nl", "nl-BE", "Vlaams"},
		{"en", "pt", "Portuguese"},
		{"en", "pt-PT", "European Portuguese"},
		{"en", "pt-BR", "Brazilian Portuguese"},
		{"en", "en", "English"},
		{"en", "en-GB", "British English"},
		{"en", "en-US", "American English"}, // American English in CLDR 24+
		{"en", lastLang2zu.String(), "Zulu"},
		{"en", firstLang2aa.String(), "Afar"},
		{"en", lastLang3zza.String(), "Zaza"},
		{"en", firstLang3ace.String(), "Achinese"},
		{"en", firstTagAr001.String(), "Modern Standard Arabic"},
		{"en", lastTagZhHant.String(), "Traditional Chinese"},
		{"en", "aaa", ""},
		{"en", "zzj", ""},
		// If full tag doesn't match, try without script or region.
		{"en", "aa-Hans", "Afar"},
		{"en", "af-Arab", "Afrikaans"},
		{"en", "zu-Cyrl", "Zulu"},
		{"en", "aa-GB", "Afar"},
		{"en", "af-NA", "Afrikaans"},
		{"en", "zu-BR", "Zulu"},
		{"agq", "zh-Hant", ""},
		// Canonical equivalents.
		{"ro", "ro-MD", "moldovenească"},
		{"ro", "mo", "moldovenească"},
		{"en", "sh", "Serbo-Croatian"},
		{"en", "sr-Latn", "Serbo-Croatian"},
		{"en", "sr", "Serbian"},
		{"en", "sr-ME", "Serbian"},
		{"en", "sr-Latn-ME", "Serbo-Croatian"}, // See comments in TestTag.
	}
	for i, tt := range tests {
		d := Languages(language.Raw.MustParse(tt.dict))
		if n := d.Name(language.Raw.MustParse(tt.tag)); n != tt.name {
			t.Errorf("%d:%s:%s: was %q; want %q", i, tt.dict, tt.tag, n, tt.name)
		}
		if len(tt.tag) <= 3 {
			if n := d.Name(language.MustParseBase(tt.tag)); n != tt.name {
				t.Errorf("%d:%s:base(%s): was %q; want %q", i, tt.dict, tt.tag, n, tt.name)
			}
		}
	}
}

func TestScript(t *testing.T) {
	tests := []struct {
		dict string
		scr  string
		name string
	}{
		{"nl", "Arab", "Arabisch"},
		{"en", "Arab", "Arabic"},
		{"en", "Zzzz", "Unknown Script"},
		{"zh-Hant", "Hang", "韓文字"},
		{"zh-Hant-HK", "Hang", "韓文字母"},
		{"zh", "Arab", "阿拉伯文"},
		{"zh-Hans-HK", "Arab", "阿拉伯文"}, // same as zh
		{"zh-Hant", "Arab", "阿拉伯文"},
		{"zh-Hant-HK", "Arab", "阿拉伯文"}, // same as zh
		// Canonicalized form
		{"en", "Qaai", "Inherited"},    // deprecated script, now is Zinh
		{"en", "sh", "Unknown Script"}, // sh canonicalizes to sr-Latn
		{"en", "en", "Unknown Script"},
		// Don't introduce scripts with canonicalization.
		{"en", "sh", "Unknown Script"}, // sh canonicalizes to sr-Latn
	}
	for i, tt := range tests {
		d := Scripts(language.MustParse(tt.dict))
		var x interface{}
		if unicode.IsUpper(rune(tt.scr[0])) {
			x = language.MustParseScript(tt.scr)
			tag, _ := language.Raw.Compose(x)
			if n := d.Name(tag); n != tt.name {
				t.Errorf("%d:%s:%s: was %q; want %q", i, tt.dict, tt.scr, n, tt.name)
			}
		} else {
			x = language.Raw.MustParse(tt.scr)
		}
		if n := d.Name(x); n != tt.name {
			t.Errorf("%d:%s:%s: was %q; want %q", i, tt.dict, tt.scr, n, tt.name)
		}
	}
}

func TestRegion(t *testing.T) {
	tests := []struct {
		dict string
		reg  string
		name string
	}{
		{"nl", "NL", "Nederland"},
		{"en", "US", "United States"},
		{"en", "ZZ", "Unknown Region"},
		{"en", "UM", "U.S. Outlying Islands"},
		{"en-GB", "UM", "U.S. Minor Outlying Islands"},
		{"en-GB", "NL", "Netherlands"},
		// Canonical equivalents
		{"en", "UK", "United Kingdom"},
		// No region
		{"en", "pt", "Unknown Region"},
		{"en", "und", "Unknown Region"},
		// Don't introduce regions with canonicalization.
		{"en", "mo", "Unknown Region"},
	}
	for i, tt := range tests {
		d := Regions(language.MustParse(tt.dict))
		var x interface{}
		if unicode.IsUpper(rune(tt.reg[0])) {
			// Region
			x = language.MustParseRegion(tt.reg)
			tag, _ := language.Raw.Compose(x)
			if n := d.Name(tag); n != tt.name {
				t.Errorf("%d:%s:%s: was %q; want %q", i, tt.dict, tt.reg, n, tt.name)
			}
		} else {
			// Tag
			x = language.Raw.MustParse(tt.reg)
		}
		if n := d.Name(x); n != tt.name {
			t.Errorf("%d:%s:%s: was %q; want %q", i, tt.dict, tt.reg, n, tt.name)
		}
	}
}

func TestSelf(t *testing.T) {
	tests := []struct {
		tag  string
		name string
	}{
		{"nl", "Nederlands"},
		{"nl-BE", "Vlaams"},
		{"en-GB", "British English"},
		{lastLang2zu.String(), "isiZulu"},
		{firstLang2aa.String(), ""},  // not defined
		{lastLang3zza.String(), ""},  // not defined
		{firstLang3ace.String(), ""}, // not defined
		{firstTagAr001.String(), "العربية الرسمية الحديثة"},
		{"ar", "العربية"},
		{lastTagZhHant.String(), "繁體中文"},
		{"aaa", ""},
		{"zzj", ""},
		// Drop entries that are not in the requested script, even if there is
		// an entry for the language.
		{"aa-Hans", ""},
		{"af-Arab", ""},
		{"zu-Cyrl", ""},
		// Append the country name in the language of the matching language.
		{"af-NA", "Afrikaans"},
		{"zh", "中文"},
		// zh-TW should match zh-Hant instead of zh!
		{"zh-TW", "繁體中文"},
		{"zh-Hant", "繁體中文"},
		{"zh-Hans", "简体中文"},
		{"zh-Hant-TW", "繁體中文"},
		{"zh-Hans-TW", "简体中文"},
		// Take the entry for sr which has the matching script.
		{"sr", "Српски"},
		// TODO: sr-ME should show up as Serbian or Montenegrin, not Serbo-
		// Croatian. This is an artifact of the current algorithm, which is the
		// way it is to have the preferred behavior for other languages such as
		// Chinese. We can hardwire this case in the table generator or package
		// code, but we first check if CLDR can be updated.
		// {"sr-ME", "Srpski"}, // Is Srpskohrvatski
		{"sr-Latn-ME", "Srpskohrvatski"},
		{"sr-Cyrl-ME", "Српски"},
		{"sr-NL", "Српски"},
		// Canonical equivalents.
		{"ro-MD", "moldovenească"},
		{"mo", "moldovenească"},
		// NOTE: kk is defined, but in Cyrillic script. For China, Arab is the
		// dominant script. We do not have data for kk-Arab and we chose to not
		// fall back in such cases.
		{"kk-CN", ""},
	}
	for i, tt := range tests {
		d := Self
		if n := d.Name(language.Raw.MustParse(tt.tag)); n != tt.name {
			t.Errorf("%d:%s: was %q; want %q", i, tt.tag, n, tt.name)
		}
	}
}
