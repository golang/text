// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package locale_test

import (
	"code.google.com/p/go.text/locale"
	"fmt"
)

func ExampleID_Canonicalize() {
	p := func(id string) {
		loc, _ := locale.Parse(id)
		l, _ := loc.Canonicalize(locale.BCP47)
		fmt.Printf("BCP47(%s) -> %s\n", id, l)
		l, _ = loc.Canonicalize(locale.Macro)
		fmt.Printf("Macro(%s) -> %s\n", id, l)
		l, _ = loc.Canonicalize(locale.All)
		fmt.Printf("All(%s) -> %s\n", id, l)
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

func ExampleID_Language() {
	fmt.Println(locale.Make("und").Language())
	fmt.Println(locale.Make("und-US").Language())
	fmt.Println(locale.Make("und-NL").Language())
	fmt.Println(locale.Make("und-419").Language())
	fmt.Println(locale.Make("und-ZZ").Language())
	// Output:
	// en Low
	// en High
	// nl High
	// en Low
	// en Low
}

func ExampleID_Script() {
	en := locale.Make("en")
	sr := locale.Make("sr")
	sr_Latn := locale.Make("sr_Latn")
	fmt.Println(en.Script())
	fmt.Println(sr.Script())
	// Was a script explicitly specified?
	_, c := sr.Script()
	fmt.Println(c == locale.Exact)
	_, c = sr_Latn.Script()
	fmt.Println(c == locale.Exact)
	// Output:
	// Latn High
	// Cyrl Low
	// false
	// true
}

func ExampleID_Region() {
	ru := locale.Make("ru")
	en := locale.Make("en")
	fmt.Println(ru.Region())
	fmt.Println(en.Region())
	// Output:
	// RU Low
	// US Low
}

func ExampleID_Part() {
	loc := locale.Make("sr-RS")
	script := loc.Part(locale.ScriptPart)
	region := loc.Part(locale.RegionPart)
	fmt.Printf("%q %q", script, region)
	// Output: "" "RS"
}
