// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package language

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

type scanTest struct {
	ok  bool // true if scanning does not result in an error
	in  string
	tok []string // the expected tokens
}

var tests = []scanTest{
	{true, "", []string{}},
	{true, "1", []string{"1"}},
	{true, "en", []string{"en"}},
	{true, "root", []string{"root"}},
	{true, "maxchars", []string{"maxchars"}},
	{false, "bad/", []string{}},
	{false, "morethan8", []string{}},
	{false, "-", []string{}},
	{false, "----", []string{}},
	{false, "_", []string{}},
	{true, "en-US", []string{"en", "US"}},
	{true, "en_US", []string{"en", "US"}},
	{false, "en-US-", []string{"en", "US"}},
	{false, "en-US--", []string{"en", "US"}},
	{false, "en-US---", []string{"en", "US"}},
	{false, "en--US", []string{"en", "US"}},
	{false, "-en-US", []string{"en", "US"}},
	{false, "-en--US-", []string{"en", "US"}},
	{false, "-en--US-", []string{"en", "US"}},
	{false, "en-.-US", []string{"en", "US"}},
	{false, ".-en--US-.", []string{"en", "US"}},
	{false, "en-u.-US", []string{"en", "US"}},
	{true, "en-u1-US", []string{"en", "u1", "US"}},
	{true, "maxchar1_maxchar2-maxchar3", []string{"maxchar1", "maxchar2", "maxchar3"}},
	{false, "moreThan8-moreThan8-e", []string{"e"}},
}

func TestScan(t *testing.T) {
	for i, tt := range tests {
		scan := makeScannerString(tt.in)
		for j := 0; !scan.done; j++ {
			if j >= len(tt.tok) {
				t.Errorf("%d: extra token %q", i, scan.token)
			} else if cmp(tt.tok[j], scan.token) != 0 {
				t.Errorf("%d: token %d: found %q; want %q", i, j, scan.token, tt.tok[j])
				break
			}
			scan.scan()
		}
		if s := strings.Join(tt.tok, "-"); cmp(s, bytes.Replace(scan.b, b("_"), b("-"), -1)) != 0 {
			t.Errorf("%d: input: found %q; want %q", i, scan.b, s)
		}
		if (scan.err == nil) != tt.ok {
			t.Errorf("%d: ok: found %v; want %v", i, scan.err == nil, tt.ok)
		}
	}
}

func TestAcceptMinSize(t *testing.T) {
	for i, tt := range tests {
		// count number of successive tokens with a minimum size.
		for sz := 1; sz <= 8; sz++ {
			scan := makeScannerString(tt.in)
			scan.end, scan.next = 0, 0
			end := scan.acceptMinSize(sz)
			n := 0
			for i := 0; i < len(tt.tok) && len(tt.tok[i]) >= sz; i++ {
				n += len(tt.tok[i])
				if i > 0 {
					n++
				}
			}
			if end != n {
				t.Errorf("%d:%d: found len %d; want %d", i, sz, end, n)
			}
		}
	}
}

type parseTest struct {
	i                    int // the index of this test
	in                   string
	lang, script, region string
	variants, ext        string
	extList              []string // only used when more than one extension is present
	invalid              bool
	rewrite              bool // special rewrite not handled by parseTag
	changed              bool // string needed to be reformatted
}

func parseTests() []parseTest {
	var manyVars string
	for i := 0; i < 50; i++ {
		manyVars += fmt.Sprintf("-abc%02d", i)
	}
	tests := []parseTest{
		{in: "root", lang: "und", changed: true},
		{in: "und", lang: "und"},
		{in: "en", lang: "en"},
		{in: "xy", lang: "und", invalid: true, changed: true},
		{in: "en-ZY", lang: "en", invalid: true},
		{in: "gsw", lang: "gsw"},
		{in: "sr_Latn", lang: "sr", script: "Latn", changed: true},
		{in: "af-Arab", lang: "af", script: "Arab"},
		{in: "nl-BE", lang: "nl", region: "BE"},
		{in: "es-419", lang: "es", region: "419"},
		{in: "und-001", lang: "und", region: "001"},
		{in: "de-latn-be", lang: "de", script: "Latn", region: "BE", changed: true},
		// Variants
		{in: "de-1901", lang: "de", variants: "1901"},
		{in: "nl-fonupa-fonupa", lang: "nl", variants: "fonupa", invalid: true},
		// Accept with unsuppressed script.
		{in: "de-Latn-1901", lang: "de", script: "Latn", variants: "1901"},
		// Do not accept with incorrect script
		{in: "de-Cyrl-1901", lang: "de", script: "Cyrl", variants: "", invalid: true},
		// Specialized.
		{in: "sl-rozaj", lang: "sl", variants: "rozaj"},
		{in: "sl-rozaj-lipaw", lang: "sl", variants: "rozaj-lipaw"},
		{in: "sl-rozaj-biske", lang: "sl", variants: "rozaj-biske"},
		{in: "sl-rozaj-lipaw-1994", lang: "sl", variants: "rozaj-lipaw", invalid: true},
		{in: "sl-rozaj-biske-1994", lang: "sl", variants: "rozaj-biske-1994"},
		{in: "sl-rozaj-1994", lang: "sl", variants: "rozaj-1994"},
		// Maximum number of variants.
		{in: "sl-rozaj-biske-1994-alalc97-fonipa-fonupa-fonxsamp", lang: "sl", variants: "rozaj-biske-1994-alalc97-fonipa-fonupa-fonxsamp"},
		// Sorting.
		{in: "sl-1994-biske-rozaj", lang: "sl", variants: "rozaj-biske-1994", changed: true},
		{in: "sl-1994-biske-rozaj-1994-biske-rozaj", lang: "sl", variants: "rozaj-biske-1994", invalid: true, changed: true},
		{in: "sl-rozaj-biske-1994-alalc97-fonupa-fonipa-fonxsamp", lang: "sl", variants: "rozaj-biske-1994-alalc97-fonipa-fonupa-fonxsamp", changed: true},
		{in: "nl-fonxsamp-alalc97-fonipa-fonupa", lang: "nl", variants: "alalc97-fonipa-fonupa-fonxsamp", changed: true},

		//{in: "nl" + manyVars, lang: "nl", variants: manyVars[1:]},
		//{in: "nl" + manyVars + manyVars, lang: "nl", variants: manyVars[1:]},
		{in: "EN_CYRL", lang: "en", script: "Cyrl", changed: true},
		// private use and extensions
		{in: "x-a-b-c-d", ext: "x-a-b-c-d"},
		{in: "x_A.-B-C_D", ext: "x-b-c-d", invalid: true, changed: true},
		{in: "x-aa-bbbb-cccccccc-d", ext: "x-aa-bbbb-cccccccc-d"},
		{in: "en-c_cc-b-bbb-a-aaa", lang: "en", changed: true, extList: []string{"a-aaa", "b-bbb", "c-cc"}},
		{in: "en-x_cc-b-bbb-a-aaa", lang: "en", ext: "x-cc-b-bbb-a-aaa", changed: true},
		{in: "en-c_cc-b-bbb-a-aaa-x-x", lang: "en", changed: true, extList: []string{"a-aaa", "b-bbb", "c-cc", "x-x"}},
		{in: "en-u-co-phonebk", lang: "en", ext: "u-co-phonebk"},
		{in: "en-Cyrl-u-co-phonebk", lang: "en", script: "Cyrl", ext: "u-co-phonebk"},
		{in: "en-US-u-co-phonebk", lang: "en", region: "US", ext: "u-co-phonebk"},
		{in: "en-US-u-co-phonebk-cu-xau", lang: "en", region: "US", ext: "u-co-phonebk-cu-xau"},
		{in: "en-scotland-u-co-phonebk", lang: "en", variants: "scotland", ext: "u-co-phonebk"},
		{in: "en-u-cu-xua-co-phonebk", lang: "en", ext: "u-co-phonebk-cu-xua", changed: true},
		{in: "en-u-def-abc-cu-xua-co-phonebk", lang: "en", ext: "u-def-abc-co-phonebk-cu-xua", changed: true},
		{in: "en-u-def-abc", lang: "en", ext: "u-def-abc"},
		{in: "en-u-cu-xua-co-phonebk-a-cd", lang: "en", extList: []string{"a-cd", "u-co-phonebk-cu-xua"}, changed: true},
		// Invalid "u" extension. Drop invalid parts.
		{in: "en-u-cu-co-phonebk", lang: "en", extList: []string{"u-co-phonebk"}, invalid: true, changed: true},
		{in: "en-u-cu-xau-co", lang: "en", extList: []string{"u-cu-xau"}, invalid: true},
		{in: "en-t-en-Cyrl-NL-fonipa", lang: "en", ext: "t-en-cyrl-nl-fonipa", changed: true},
		{in: "en-t-en-Cyrl-NL-fonipa-t0-abc-def", lang: "en", ext: "t-en-cyrl-nl-fonipa-t0-abc-def", changed: true},
		{in: "en-t-t0-abcd", lang: "en", ext: "t-t0-abcd"},
		// Not necessary to have changed here.
		{in: "en-t-nl-abcd", lang: "en", ext: "t-nl", invalid: true},
		{in: "en-t-nl-latn", lang: "en", ext: "t-nl-latn"},
		{in: "en-t-t0-abcd-x-a", lang: "en", extList: []string{"t-t0-abcd", "x-a"}},
		// invalid
		{in: "", lang: "und", invalid: true, changed: true},
		{in: "-", lang: "und", invalid: true, changed: true},
		{in: "x", lang: "und", invalid: true, changed: true},
		{in: "x-", lang: "und", invalid: true, changed: true},
		{in: "x--", lang: "und", invalid: true, changed: true},
		{in: "a-a-b-c-d", lang: "und", invalid: true, changed: true},
		{in: "en-", lang: "en", invalid: true},
		{in: "enne-", lang: "und", invalid: true, changed: true},
		{in: "en.", lang: "und", invalid: true, changed: true},
		{in: "en.-latn", lang: "und", invalid: true, changed: true},
		{in: "en.-en", lang: "en", invalid: true},
		{in: "x-a-tooManyChars-c-d", ext: "x-a-c-d", invalid: true, changed: true},
		{in: "a-tooManyChars-c-d", lang: "und", invalid: true, changed: true},
		// TODO: check key-value validity
		// { in: "en-u-cu-xd", lang: "en", ext: "u-cu-xd", invalid: true },
		{in: "en-t-abcd", lang: "en", invalid: true},
		{in: "en-Latn-US-en", lang: "en", script: "Latn", region: "US", invalid: true},
		// rewrites
		{in: "zh-min", lang: "und", rewrite: true, changed: true},
		{in: "zh-min-nan", lang: "nan", changed: true},
		{in: "zh-yue", lang: "yue", changed: true},
		{in: "zh-xiang", lang: "hsn", rewrite: true, changed: true},
		{in: "zh-guoyu", lang: "zh", rewrite: true, changed: true},
		{in: "iw", lang: "iw", changed: false},
		{in: "sgn-BE-FR", lang: "sfb", rewrite: true, changed: true},
		{in: "i-klingon", lang: "tlh", rewrite: true, changed: true},
	}
	for i, tt := range tests {
		tests[i].i = i
		if tt.extList != nil {
			tests[i].ext = strings.Join(tt.extList, "-")
		}
		if tt.ext != "" && tt.extList == nil {
			tests[i].extList = []string{tt.ext}
		}
	}
	return tests
}

func TestParseExtensions(t *testing.T) {
	for i, tt := range parseTests() {
		if tt.ext == "" || tt.rewrite {
			continue
		}
		scan := makeScannerString(tt.in)
		if len(scan.b) > 1 && scan.b[1] != '-' {
			scan.end = nextExtension(string(scan.b), 0)
			scan.next = scan.end + 1
			scan.scan()
		}
		start := scan.start
		scan.toLower(start, len(scan.b))
		parseExtensions(&scan)
		ext := string(scan.b[start:])
		if ext != tt.ext {
			t.Errorf("%d(%s): ext was %v; want %v", i, tt.in, ext, tt.ext)
		}
		if changed := !strings.HasPrefix(tt.in[start:], ext); changed != tt.changed {
			t.Errorf("%d(%s): changed was %v; want %v", i, tt.in, changed, tt.changed)
		}
	}
}

// partChecks runs checks for each part by calling the function returned by f.
func partChecks(t *testing.T, f func(*parseTest) func(Part) string) {
	for i, tt := range parseTests() {
		get := f(&tt)
		if get == nil {
			continue
		}
		l, _ := getLangID(b(tt.lang))
		if s, g := get(LanguagePart), l.String(); s != g {
			t.Errorf("%d: lang was %q; want %q", i, s, g)
		}
		if s, g := get(ScriptPart), tt.script; s != g {
			t.Errorf("%d: script was %q; want %q", i, s, g)
		}
		if s, g := get(RegionPart), tt.region; s != g {
			t.Errorf("%d: region was %q; want %q", i, s, g)
		}
		if s, g := get(VariantPart), tt.variants; s != g {
			t.Errorf("%d: variants was %q; want %q", i, s, g)
		}
		for _, g := range tt.extList {
			if s := get(Extension(g[0])); s != g[2:] {
				t.Errorf("%d: extension '%c' was %q; want %q", i, g[0], s, g[2:])
			}
		}
		if s := get(Extension('q')); s != "" {
			t.Errorf(`%d: unused extension 'q' was %q; want ""`, s)
		}
	}
}

func TestParseTag(t *testing.T) {
	partChecks(t, func(tt *parseTest) func(Part) string {
		if strings.HasPrefix(tt.in, "x-") || tt.rewrite {
			return nil
		}
		scan := makeScannerString(tt.in)
		id, end := parseTag(&scan)
		s := string(scan.b[:end])
		if changed := !strings.HasPrefix(tt.in, s); changed != tt.changed && tt.ext == "" {
			t.Errorf("%d: changed was %v; want %v", tt.i, changed, tt.changed)
		}
		id.str = &s
		tt.ext = ""
		tt.extList = []string{}
		return func(p Part) string {
			return id.Part(p)
		}
	})
}

func TestParse(t *testing.T) {
	partChecks(t, func(tt *parseTest) func(Part) string {
		id, err := Parse(tt.in)
		ext := ""
		if id.str != nil {
			if strings.HasPrefix(*id.str, "x-") {
				ext = *id.str
			} else if int(id.pExt) < len(*id.str) && id.pExt > 0 {
				ext = (*id.str)[id.pExt+1:]
			}
		}
		if ext != tt.ext {
			t.Errorf("%d: ext was %q; want %q", tt.i, ext, tt.ext)
		}
		changed := id.str == nil || !strings.HasPrefix(tt.in, *id.str)
		if changed != tt.changed {
			t.Errorf("%d: changed was %v; want %v", tt.i, changed, tt.changed)
		}
		if (err != nil) != tt.invalid {
			t.Errorf("%d: invalid was %v; want %v. Error: %v", tt.i, err != nil, tt.invalid, err)
		}
		return func(p Part) string {
			return id.Part(p)
		}
	})
}

func TestErrors(t *testing.T) {
	mkInvalid := func(s string) error {
		return mkErrInvalid([]byte(s))
	}
	tests := []struct {
		in  string
		out error
	}{
		// invalid subtags.
		{"ac", mkInvalid("ac")},
		{"AC", mkInvalid("ac")},
		{"aa-Uuuu", mkInvalid("Uuuu")},
		{"aa-AB", mkInvalid("AB")},
		// ill-formed wins over invalid.
		{"ac-u", errSyntax},
		{"ac-u-ca", errSyntax},
		{"ac-u-ca-co-pinyin", errSyntax},
		{"noob", errSyntax},
	}
	for _, tt := range tests {
		_, err := Parse(tt.in)
		if err != tt.out {
			t.Errorf("%s: was %q; want %q", tt.in, err, tt.out)
		}
	}
}

func TestPart(t *testing.T) {
	partChecks(t, func(tt *parseTest) func(Part) string {
		id, _ := Parse(tt.in)
		return func(p Part) string {
			return id.Part(p)
		}
	})
}

func TestParts(t *testing.T) {
	partChecks(t, func(tt *parseTest) func(Part) string {
		id, _ := Parse(tt.in)
		m := id.Parts()
		return func(p Part) string {
			return m[p]
		}
	})
}

func TestCompose1(t *testing.T) {
	partChecks(t, func(tt *parseTest) func(Part) string {
		m := make(map[Part]string)
		set := func(p Part, s string) {
			if s != "" {
				m[p] = strings.ToUpper(s)
			}
		}
		set(LanguagePart, tt.lang)
		set(ScriptPart, tt.script)
		set(RegionPart, tt.region)
		if tt.variants != "" {
			m[VariantPart] = tt.variants + "-tooManyChars-inv@lid-" + tt.variants
		}
		for _, ext := range tt.extList {
			set(Extension(ext[0]), ext[2:])
		}
		id, err := Compose(m)
		if tt.variants != "" && err == nil {
			t.Errorf("%d: no error for invalid variant", tt.i)
		}
		return func(p Part) string {
			return id.Part(p)
		}
	})
}

func TestCompose2(t *testing.T) {
	partChecks(t, func(tt *parseTest) func(Part) string {
		m := make(map[Part]string)
		tag := tt.lang
		for _, s := range []string{tt.script, tt.region, tt.variants} {
			if s != "" {
				tag += "-" + s
			}
		}
		m[TagPart] = tag
		for _, ext := range tt.extList {
			m[Extension(ext[0])] = ext[2:] + "-tooManyChars"
		}
		id, err := Compose(m)
		if len(tt.extList) > 0 && err == nil {
			t.Errorf("%d: no error for invalid variant", tt.i)
		}
		return func(p Part) string {
			return id.Part(p)
		}
	})
}

func mk(s string) Tag {
	id, _ := Parse(s)
	return id
}

func TestParseAcceptLanguage(t *testing.T) {
	type res struct {
		t Tag
		q float32
	}
	en := []res{{mk("en"), 1.0}}
	tests := []struct {
		out []res
		in  string
		ok  bool
	}{
		{en, "en", true},
		{en, "   en", true},
		{en, "en   ", true},
		{en, "  en  ", true},
		{en, "en,", true},
		{en, ",en", true},
		{en, ",,,en,,,", true},
		{en, ",en;q=1", true},
		// We allow an empty input, contrary to spec.
		{nil, "", true},

		// errors
		{nil, ";", false},
		{nil, "$", false},
		{nil, "e;", false},
		{nil, "x;", false},
		{nil, "x", false},
		{nil, "ac", false}, // non-existing language
		{nil, "aa;", false},
		{nil, "aa;q", false},
		{nil, "aa;q=", false},
		{nil, "aa;q=.", false},

		// odd fallbacks
		{
			[]res{{mk("en"), 0.1}},
			" english ;q=.1",
			true,
		},
		{
			[]res{{mk("it"), 1.0}, {mk("de"), 1.0}, {mk("fr"), 1.0}},
			" italian, deutsch, french",
			true,
		},

		// lists
		{
			[]res{{mk("en"), 0.1}},
			"en;q=.1",
			true,
		},
		{
			[]res{{mk("mul"), 1.0}},
			"*",
			true,
		},
		{
			[]res{{mk("en"), 1.0}, {mk("de"), 1.0}},
			"en,de",
			true,
		},
		{
			[]res{{mk("en"), 1.0}, {mk("de"), .5}},
			"en,de;q=0.5",
			true,
		},
		{
			[]res{{mk("de"), 0.8}, {mk("en"), 0.5}},
			"  en ;   q    =   0.5    ,  , de;q=0.8",
			true,
		},
		{
			[]res{{mk("en"), 1.0}, {mk("de"), 1.0}, {mk("fr"), 1.0}, {mk("tlh"), 1.0}},
			"en,de,fr,i-klingon",
			true,
		},
		// sorting
		{
			[]res{{mk("tlh"), 0.4}, {mk("de"), 0.2}, {mk("fr"), 0.2}, {mk("en"), 0.1}},
			"en;q=0.1,de;q=0.2,fr;q=0.2,i-klingon;q=0.4",
			true,
		},
		// dropping
		{
			[]res{{mk("fr"), 0.2}, {mk("en"), 0.1}},
			"en;q=0.1,de;q=0,fr;q=0.2,i-klingon;q=0.0",
			true,
		},
	}
	for i, tt := range tests {
		tags, qs, e := ParseAcceptLanguage(tt.in)
		if e == nil != tt.ok {
			t.Errorf("%d:%s:err: was %v; want %v", i, tt.in, e == nil, tt.ok)
		}
		for j, tag := range tags {
			if out := tt.out[j]; !tag.equalTags(out.t) || qs[j] != out.q {
				t.Errorf("%d:%s: was %s, %1f; want %s, %1f", i, tt.in, tag, qs[j], out.t, out.q)
				break
			}
		}
	}
}
