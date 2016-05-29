// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Unicode table generator.
// Data read from the web.

// +build ignore

package main

import (
	"flag"
	"log"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/internal/gen"
	"golang.org/x/text/internal/triegen"
	"golang.org/x/text/internal/ucd"
	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/unicode/rangetable"
)

var outputFile = flag.String("output", "tables.go", "output file for generated tables; default tables.go")

var assigned, disallowedRunes *unicode.RangeTable

func main() {
	gen.Init()

	// Load data
	runes := []rune{}
	// PrecisIgnorableProperties: https://tools.ietf.org/html/rfc7564#section-9.13
	ucd.Parse(gen.OpenUCDFile("DerivedCoreProperties.txt"), func(p *ucd.Parser) {
		if p.String(1) == "Default_Ignorable_Code_Point" {
			runes = append(runes, p.Rune(0))
		}
	})
	ucd.Parse(gen.OpenUCDFile("PropList.txt"), func(p *ucd.Parser) {
		switch p.String(1) {
		case "Noncharacter_Code_Point":
			runes = append(runes, p.Rune(0))
		}
	})
	// OldHangulJamo: https://tools.ietf.org/html/rfc5892#section-2.9
	ucd.Parse(gen.OpenUCDFile("HangulSyllableType.txt"), func(p *ucd.Parser) {
		switch p.String(1) {
		case "L", "V", "T":
			runes = append(runes, p.Rune(0))
		}
	})

	disallowedRunes = rangetable.New(runes...)
	assigned = rangetable.Assigned(unicode.Version)

	writeTables()
	gen.Repackage("gen_trieval.go", "trieval.go", "precis")
}

// The Exceptions class as defined in RFC 5892
// https://tools.ietf.org/html/rfc5892#section-2.6
var exceptions = map[uint32]property{
	0x00DF: pValid,
	0x03C2: pValid,
	0x06FD: pValid,
	0x06FE: pValid,
	0x0F0B: pValid,
	0x3007: pValid,
	0x00B7: contextO,
	0x0375: contextO,
	0x05F3: contextO,
	0x05F4: contextO,
	0x30FB: contextO,
	0x0660: contextO,
	0x0661: contextO,
	0x0662: contextO,
	0x0663: contextO,
	0x0664: contextO,
	0x0665: contextO,
	0x0666: contextO,
	0x0667: contextO,
	0x0668: contextO,
	0x0669: contextO,
	0x06F0: contextO,
	0x06F1: contextO,
	0x06F2: contextO,
	0x06F3: contextO,
	0x06F4: contextO,
	0x06F5: contextO,
	0x06F6: contextO,
	0x06F7: contextO,
	0x06F8: contextO,
	0x06F9: contextO,
	0x0640: disallowed,
	0x07FA: disallowed,
	0x302E: disallowed,
	0x302F: disallowed,
	0x3031: disallowed,
	0x3032: disallowed,
	0x3033: disallowed,
	0x3034: disallowed,
	0x3035: disallowed,
	0x303B: disallowed,
}

// LetterDigits: https://tools.ietf.org/html/rfc5892#section-2.1
// r in {Ll, Lu, Lo, Nd, Lm, Mn, Mc}.
func isLetterDigits(r rune) bool {
	return unicode.In(r,
		unicode.Ll, unicode.Lu, unicode.Lm, unicode.Lo, // Letters
		unicode.Mn, unicode.Mc, // Modifiers
		unicode.Nd, // Digits
	)
}

func isIdDisAndFreePVal(r rune) bool {
	return unicode.In(r,
		// OtherLetterDigits: https://tools.ietf.org/html/rfc7564#section-9.18
		// r in in {Lt, Nl, No, Me}
		unicode.Lt, unicode.Nl, unicode.No, // Other letters / numbers
		unicode.Me, // Modifiers

		// Spaces: https://tools.ietf.org/html/rfc7564#section-9.14
		// r in in {Zs}
		unicode.Zs,

		// Symbols: https://tools.ietf.org/html/rfc7564#section-9.15
		// r in {Sm, Sc, Sk, So}
		unicode.Sm, unicode.Sc, unicode.Sk, unicode.So,

		// Punctuation: https://tools.ietf.org/html/rfc7564#section-9.16
		// r in {Pc, Pd, Ps, Pe, Pi, Pf, Po}
		unicode.Pc, unicode.Pd, unicode.Ps, unicode.Pe,
		unicode.Pi, unicode.Pf, unicode.Po,
	)
}

// HasCompat: https://tools.ietf.org/html/rfc7564#section-9.17
func hasCompat(r rune) bool {
	return !norm.NFKC.IsNormalString(string(r))
}

// From https://tools.ietf.org/html/rfc5892:
//
// If .cp. .in. Exceptions Then Exceptions(cp);
//   Else If .cp. .in. BackwardCompatible Then BackwardCompatible(cp);
//   Else If .cp. .in. Unassigned Then UNASSIGNED;
//   Else If .cp. .in. ASCII7 Then PVALID;
//   Else If .cp. .in. JoinControl Then CONTEXTJ;
//   Else If .cp. .in. OldHangulJamo Then DISALLOWED;
//   Else If .cp. .in. PrecisIgnorableProperties Then DISALLOWED;
//   Else If .cp. .in. Controls Then DISALLOWED;
//   Else If .cp. .in. HasCompat Then ID_DIS or FREE_PVAL;
//   Else If .cp. .in. LetterDigits Then PVALID;
//   Else If .cp. .in. OtherLetterDigits Then ID_DIS or FREE_PVAL;
//   Else If .cp. .in. Spaces Then ID_DIS or FREE_PVAL;
//   Else If .cp. .in. Symbols Then ID_DIS or FREE_PVAL;
//   Else If .cp. .in. Punctuation Then ID_DIS or FREE_PVAL;
//   Else DISALLOWED;

func writeTables() {
	propTrie := triegen.NewTrie("derivedProperties")
	w := gen.NewCodeWriter()
	defer w.WriteGoFile(*outputFile, "precis")
	gen.WriteUnicodeVersion(w)

	// Iterate over all the runes...
	for i := uint32(0); i < unicode.MaxRune; i++ {
		r := rune(i)

		if !utf8.ValidRune(r) {
			continue
		}

		p, ok := exceptions[i]
		switch {
		case ok:
		case !unicode.In(r, assigned):
			p = unassigned
		case r >= 0x0021 && r <= 0x007e: // Is ASCII 7
			p = pValid
		case r == 0x200C || r == 0x200D: // Is join control
			p = contextJ
		case unicode.In(r, disallowedRunes, unicode.Cc):
			p = disallowed
		case hasCompat(r):
			p = idDisOrFreePVal
		case isLetterDigits(r):
			p = pValid
		case isIdDisAndFreePVal(r):
			p = idDisOrFreePVal
		default:
			p = disallowed
		}
		propTrie.Insert(r, uint64(p))
	}
	sz, err := propTrie.Gen(w)
	if err != nil {
		log.Fatal(err)
	}
	w.Size += sz
}
