// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// Generator for currency-related data.

package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/text/cldr"
	"golang.org/x/text/internal/gen"
	"golang.org/x/text/internal/tag"
)

var (
	test = flag.Bool("test", false,
		"test existing tables; can be used to compare web data with package data.")
	outputFile = flag.String("output", "tables.go", "output file")

	draft = flag.String("draft",
		"contributed",
		`Minimal draft requirements (approved, contributed, provisional, unconfirmed).`)
)

func main() {
	gen.Init()

	rewriteCommon()

	// Read the CLDR zip file.
	r := gen.OpenCLDRCoreZip()
	defer r.Close()

	d := &cldr.Decoder{}
	d.SetDirFilter("supplemental")
	data, err := d.DecodeZip(r)
	if err != nil {
		log.Fatalf("DecodeZip: %v", err)
	}

	w := gen.NewCodeWriter()
	defer w.WriteGoFile(*outputFile, "currency")

	fmt.Fprintln(w, `import "golang.org/x/text/internal/tag"`)

	gen.WriteCLDRVersion(w)
	genCurrencies(w, data.Supplemental())
}

func rewriteCommon() {
	// Generate common.go
	src, err := ioutil.ReadFile("gen_common.go")
	if err != nil {
		log.Fatal(err)
	}
	const toDelete = "// +build ignore\n\npackage main\n\n"
	i := bytes.Index(src, []byte(toDelete))
	if i < 0 {
		log.Fatalf("could not find %q in gen_common.go", toDelete)
	}
	w := &bytes.Buffer{}
	w.Write(src[i+len(toDelete):])
	gen.WriteGoFile("common.go", "currency", w.Bytes())
}

var constants = []string{
	// Undefined and testing.
	"XXX", "XTS",
	// G11 currencies https://en.wikipedia.org/wiki/G10_currencies.
	"USD", "EUR", "JPY", "GBP", "CHF", "AUD", "NZD", "CAD", "SEK", "NOK", "DKK",
	// Precious metals.
	"XAG", "XAU", "XPT", "XPD",

	// Additional common currencies as defined by CLDR.
	"BRL", "CNY", "INR", "RUB", "HKD", "IDR", "KRW", "MXN", "PLN", "SAR",
	"THB", "TRY", "TWD", "ZAR",
}

func genCurrencies(w *gen.CodeWriter, data *cldr.SupplementalData) {
	// 3-letter ISO currency codes
	// Start with dummy to let index start at 1.
	currencies := []string{"\x00\x00\x00\x00"}

	// currency codes
	for _, reg := range data.CurrencyData.Region {
		for _, cur := range reg.Currency {
			currencies = append(currencies, cur.Iso4217)
		}
	}

	sort.Strings(currencies)
	// Unique the elements.
	k := 0
	for i := 1; i < len(currencies); i++ {
		if currencies[k] != currencies[i] {
			currencies[k+1] = currencies[i]
			k++
		}
	}
	currencies = currencies[:k+1]

	// Close with dummy for simpler and faster searching.
	currencies = append(currencies, "\xff\xff\xff\xff")

	// Write currency values.
	fmt.Fprintln(w, "const (")
	for _, c := range constants {
		index := sort.SearchStrings(currencies, c)
		fmt.Fprintf(w, "\t%s = %d\n", strings.ToLower(c), index)
	}
	fmt.Fprint(w, ")")

	// Compute currency-related data that we merge into the table.
	for _, info := range data.CurrencyData.Fractions[0].Info {
		if info.Iso4217 == "DEFAULT" {
			continue
		}
		standard := getRoundingIndex(info.Digits, info.Rounding)
		cash := getRoundingIndex(info.CashDigits, info.CashRounding)

		index := sort.SearchStrings(currencies, info.Iso4217)
		currencies[index] += mkCurrencyInfo(standard, cash)
	}

	// Set default values for entries that weren't touched.
	for i, c := range currencies {
		if len(c) == 3 {
			currencies[i] += mkCurrencyInfo(0, 0)
		}
	}

	w.WriteComment(`
	currency holds an alphabetically sorted list of canonical 3-letter currency
	identifiers. Each identifier is followed by a byte of type currencyInfo,
	defined in gen_common.go.`)
	w.WriteConst("currency", tag.Index(strings.Join(currencies, "")))

	// Hack alert: gofmt indents a trailing comment after an indented string.
	// Ensure that the next thing written is not a comment.
	w.WriteConst("numCurrencies", len(currencies)-2)
}

func mkCurrencyInfo(standard, cash int) string {
	return string([]byte{byte(cash<<cashShift | standard)})
}

func getRoundingIndex(digits, rounding string) int {
	round := roundings[0] // default

	if digits != "" {
		round.scale = parseUint8(digits)
	}
	if rounding != "" && rounding != "0" { // 0 means 1 here in CLDR
		round.increment = parseUint8(rounding)
	}

	// Will panic if the entry doesn't exist:
	for i, r := range roundings {
		if r == round {
			return i
		}
	}
	log.Fatalf("Rounding entry %#v does not exist.", round)
	panic("unreachable")
}

func parseUint8(str string) uint8 {
	x, err := strconv.ParseUint(str, 10, 8)
	if err != nil {
		// Show line number of where this function was called.
		log.New(os.Stderr, "", log.Lshortfile).Output(2, err.Error())
		os.Exit(1)
	}
	return uint8(x)
}
