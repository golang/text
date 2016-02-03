// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package format

import (
	"flag"
	"log"
	"reflect"
	"testing"
	"unicode/utf8"

	"golang.org/x/text/internal"
	"golang.org/x/text/internal/gen"
	"golang.org/x/text/internal/testtext"
	"golang.org/x/text/language"
	"golang.org/x/text/unicode/cldr"
)

var draft = flag.String("draft",
	"contributed",
	`Minimal draft requirements (approved, contributed, provisional, unconfirmed).`)

func TestNumberSystems(t *testing.T) {
	testtext.SkipIfNotLong(t)

	r := gen.OpenCLDRCoreZip()
	defer r.Close()

	d := &cldr.Decoder{}
	d.SetDirFilter("supplemental")
	d.SetSectionFilter("numberingSystem")
	data, err := d.DecodeZip(r)
	if err != nil {
		t.Fatalf("DecodeZip: %v", err)
	}

	for _, ns := range data.Supplemental().NumberingSystems.NumberingSystem {
		n := numberSystemMap[ns.Id]
		if int(n) >= len(numSysData) {
			continue
		}
		d := numSysData[n]
		val := byte(0)
		for _, rWant := range ns.Digits {
			var x [utf8.UTFMax]byte
			copy(x[:], d.zero[:d.digitSize])
			x[d.digitSize-1] += val
			rGot, _ := utf8.DecodeRune(x[:])
			if rGot != rWant {
				t.Errorf("%s:%d: got %U; want %U", ns.Id, val, rGot, rWant)
			}
			val++
		}
	}
}

func TestSymbols(t *testing.T) {
	testtext.SkipIfNotLong(t)

	draft, err := cldr.ParseDraft(*draft)
	if err != nil {
		log.Fatalf("invalid draft level: %v", err)
	}

	r := gen.OpenCLDRCoreZip()
	defer r.Close()

	d := &cldr.Decoder{}
	d.SetDirFilter("main")
	d.SetSectionFilter("numbers")
	data, err := d.DecodeZip(r)
	if err != nil {
		t.Fatalf("DecodeZip: %v", err)
	}

	for _, lang := range data.Locales() {
		ldml := data.RawLDML(lang)
		if ldml.Numbers == nil {
			continue
		}
		langIndex, ok := language.CompactIndex(language.MustParse(lang))
		if !ok {
			t.Fatalf("No compact index for language %s", lang)
		}

		syms := cldr.MakeSlice(&ldml.Numbers.Symbols)
		syms.SelectDraft(draft)

		for _, sym := range ldml.Numbers.Symbols {
			if sym.NumberSystem == "" {
				continue
			}
			testCases := []struct {
				name string
				st   symbolType
				x    interface{}
			}{
				{"Decimal", symDecimal, sym.Decimal},
				{"Group", symGroup, sym.Group},
				{"List", symList, sym.List},
				{"PercentSign", symPercentSign, sym.PercentSign},
				{"PlusSign", symPlusSign, sym.PlusSign},
				{"MinusSign", symMinusSign, sym.MinusSign},
				{"Exponential", symExponential, sym.Exponential},
				{"SuperscriptingExponent", symSuperscriptingExponent, sym.SuperscriptingExponent},
				{"PerMille", symPerMille, sym.PerMille},
				{"Infinity", symInfinity, sym.Infinity},
				{"NaN", symNan, sym.Nan},
				{"TimeSeparator", symTimeSeparator, sym.TimeSeparator},
			}
			for _, tc := range testCases {
				// Extract the wanted value.
				v := reflect.ValueOf(tc.x)
				if v.Len() == 0 {
					return
				}
				if v.Len() > 1 {
					t.Fatalf("Multiple values of %q within single symbol not supported.", tc.name)
				}
				want := v.Index(0).MethodByName("Data").Call(nil)[0].String()

				// Extract the value from the table.
				ns := numberSystemMap[sym.NumberSystem]
				strIndex := -1
				for strIndex == -1 {
					index := langToDefaults[langIndex]
					if index&0x80 == 0 && ns == 0 {
						// The index directly points to the symbol data.
						strIndex = int(symIndex[index][tc.st])
						continue
					}
					// The index points to a list of symbol data indexes.
					for _, e := range langToAlt[index&^0x80:] {
						if int(e.compactTag) != langIndex {
							if langIndex == 0 {
								// Fall back to Latin.
								ns = 0
							} else {
								// Fall back to parent.
								langIndex = int(internal.Parent[langIndex])
							}
							break
						}
						if e.numberSystem == ns {
							strIndex = int(symIndex[e.symIndex][tc.st])
							break
						}
					}
				}
				got := symData.Elem(strIndex)
				if got != want {
					t.Errorf("%s:%s:%s: got %q; want %q", lang, sym.NumberSystem, tc.name, got, want)
				}
			}
		}
	}
}
