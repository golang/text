// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package language_test

import (
	"code.google.com/p/go.text/language"
	"fmt"
)

func ExampleCanonType() {
	p := func(id string) {
		fmt.Printf("BCP47(%s) -> %s\n", id, language.BCP47.Make(id))
		fmt.Printf("Macro(%s) -> %s\n", id, language.Macro.Make(id))
		fmt.Printf("All(%s) -> %s\n", id, language.All.Make(id))
	}
	p("en-Latn")
	p("sh")
	p("zh-cmn")
	p("bjd")
	p("iw-Latn-fonipa-u-cu-usd")
	// Output:
	// BCP47(en-Latn) -> en
	// Macro(en-Latn) -> en-Latn
	// All(en-Latn) -> en
	// BCP47(sh) -> sh
	// Macro(sh) -> sh
	// All(sh) -> sr-Latn
	// BCP47(zh-cmn) -> cmn
	// Macro(zh-cmn) -> zh
	// All(zh-cmn) -> zh
	// BCP47(bjd) -> drl
	// Macro(bjd) -> bjd
	// All(bjd) -> drl
	// BCP47(iw-Latn-fonipa-u-cu-usd) -> he-Latn-fonipa-u-cu-usd
	// Macro(iw-Latn-fonipa-u-cu-usd) -> iw-Latn-fonipa-u-cu-usd
	// All(iw-Latn-fonipa-u-cu-usd) -> he-Latn-fonipa-u-cu-usd
}

func ExampleTag_Base() {
	fmt.Println(language.Make("und").Base())
	fmt.Println(language.Make("und-US").Base())
	fmt.Println(language.Make("und-NL").Base())
	fmt.Println(language.Make("und-419").Base())
	fmt.Println(language.Make("und-ZZ").Base())
	// Output:
	// en Low
	// en High
	// nl High
	// en Low
	// en Low
}

func ExampleTag_Script() {
	en := language.Make("en")
	sr := language.Make("sr")
	sr_Latn := language.Make("sr_Latn")
	fmt.Println(en.Script())
	fmt.Println(sr.Script())
	// Was a script explicitly specified?
	_, c := sr.Script()
	fmt.Println(c == language.Exact)
	_, c = sr_Latn.Script()
	fmt.Println(c == language.Exact)
	// Output:
	// Latn High
	// Cyrl Low
	// false
	// true
}

func ExampleTag_Region() {
	ru := language.Make("ru")
	en := language.Make("en")
	fmt.Println(ru.Region())
	fmt.Println(en.Region())
	// Output:
	// RU Low
	// US Low
}

func ExampleCompose() {
	nl, _ := language.ParseBase("nl")
	us, _ := language.ParseRegion("US")
	de := language.Make("de-1901-u-co-phonebk")
	jp := language.Make("ja-JP")
	fi := language.Make("fi-x-ing")

	u, _ := language.ParseExtension("u-nu-arabic")
	x, _ := language.ParseExtension("x-piglatin")

	// Combine a base language and region.
	fmt.Println(language.Compose(nl, us))
	// Combine a base language and extension.
	fmt.Println(language.Compose(nl, x))
	// Replace the region.
	fmt.Println(language.Compose(jp, us))
	// Combine several tags.
	fmt.Println(language.Compose(us, nl, u))

	// Replace the base language of a tag.
	fmt.Println(language.Compose(de, nl))
	fmt.Println(language.Compose(de, nl, u))
	// Remove the base language.
	fmt.Println(language.Compose(de, language.Base{}))
	// Remove all variants.
	fmt.Println(language.Compose(de, []language.Variant{}))
	// Remove all extensions.
	fmt.Println(language.Compose(de, []language.Extension{}))
	fmt.Println(language.Compose(fi, []language.Extension{}))
	// Remove all variants and extensions.
	fmt.Println(language.Compose(de.Raw()))

	// An error is gobbled or returned if non-nil.
	fmt.Println(language.Compose(language.ParseRegion("ZA")))
	fmt.Println(language.Compose(language.ParseRegion("HH")))

	// Compose uses the same Default canonicalization as Make.
	fmt.Println(language.Compose(language.Raw.Parse("en-Latn-UK")))

	// Call compose on a different CanonType for different results.
	fmt.Println(language.All.Compose(language.Raw.Parse("en-Latn-UK")))

	// Output:
	// nl-US <nil>
	// nl-x-piglatin <nil>
	// ja-US <nil>
	// nl-US-u-nu-arabic <nil>
	// nl-1901-u-co-phonebk <nil>
	// nl-1901-u-nu-arabic <nil>
	// und-1901-u-co-phonebk <nil>
	// de-u-co-phonebk <nil>
	// de-1901 <nil>
	// fi <nil>
	// de <nil>
	// und-ZA <nil>
	// und language: subtag "HH" is well-formed but unknown
	// en-Latn-GB <nil>
	// en-GB <nil>
}

func ExampleParse_errors() {
	for _, s := range []string{"Foo", "Bar", "Foobar"} {
		_, err := language.Parse(s)
		if err != nil {
			if inv, ok := err.(language.ValueError); ok {
				fmt.Println(inv.Subtag())
			} else {
				fmt.Println(s)
			}
		}
	}
	for _, s := range []string{"en", "aa-Uuuu", "AC", "ac-u"} {
		_, err := language.Parse(s)
		switch e := err.(type) {
		case language.ValueError:
			fmt.Printf("%s: culprit %q\n", s, e.Subtag())
		case nil:
			// No error.
		default:
			// A syntax error.
			fmt.Printf("%s: ill-formed\n", s)
		}
	}
	// Output:
	// foo
	// Foobar
	// aa-Uuuu: culprit "Uuuu"
	// AC: culprit "ac"
	// ac-u: ill-formed
}
