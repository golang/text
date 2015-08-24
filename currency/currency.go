// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package currency

//go:generate go run gen.go gen_common.go -output tables.go

import (
	"errors"

	"golang.org/x/text/internal/tag"
)

// TODO:
// - language-specific currency names.
// - Value type
// - currency formatting.
// - get currency from tag/region.
// - currency information per region
// - register currency code (there are no private use area)

// TODO: remove Currency type from package language.

// Kind determines the rounding and rendering properties of a currency value.
type Kind struct {
	rounding rounding
	// TODO: formatting type: standard, accounting. See CLDR.
}

type rounding byte

const (
	standard rounding = iota
	cash
)

var (
	// Standard defines standard rounding and formatting for currencies.
	Standard Kind = Kind{rounding: standard}

	// Cash defines rounding and formatting standards for cash transactions.
	Cash Kind = Kind{rounding: cash}

	// Accounting defines rounding and formatting standards for accounting.
	Accounting Kind = Kind{rounding: standard}
)

// Rounding reports the rounding characteristics for the given currency, where
// scale is the number of fractional decimals and increment is the number of
// units in terms of 10^(-scale) to which to round to.
func (k Kind) Rounding(c Currency) (scale, increment int) {
	info := currency.Elem(int(c.index))[3]
	switch k.rounding {
	case standard:
		info &= roundMask
	case cash:
		info >>= cashShift
	}
	return int(roundings[info].scale), int(roundings[info].increment)
}

// Currency is an ISO 4217 currency designator.
type Currency struct {
	index uint16
}

// String returns the ISO code of c.
func (c Currency) String() string {
	if c.index == 0 {
		return "XXX"
	}
	return currency.Elem(int(c.index))[:3]
}

var (
	errSyntax = errors.New("currency: tag is not well-formed")
	errValue  = errors.New("currency: tag is not a recognized currency")
)

// ParseISO parses a 3-letter ISO 4217 code. It returns an error if s not
// well-formed or not a recognized currency code.
func ParseISO(s string) (Currency, error) {
	var buf [4]byte // Take one byte more to detect oversize keys.
	key := buf[:copy(buf[:], s)]
	if !tag.FixCase("XXX", key) {
		return Currency{}, errSyntax
	}
	if i := currency.Index(key); i >= 0 {
		return Currency{uint16(i)}, nil
	}
	return Currency{}, errValue
}

// MustParseISO is like ParseISO, but panics if the given currency
// cannot be parsed. It simplifies safe initialization of Currency values.
func MustParseISO(s string) Currency {
	c, err := ParseISO(s)
	if err != nil {
		panic(err)
	}
	return c
}

var (
	// Undefined and testing.
	XXX Currency = Currency{xxx}
	XTS Currency = Currency{xts}

	// G10 currencies https://en.wikipedia.org/wiki/G10_currencies.
	USD Currency = Currency{usd}
	EUR Currency = Currency{eur}
	JPY Currency = Currency{jpy}
	GBP Currency = Currency{gbp}
	CHF Currency = Currency{chf}
	AUD Currency = Currency{aud}
	NZD Currency = Currency{nzd}
	CAD Currency = Currency{cad}
	SEK Currency = Currency{sek}
	NOK Currency = Currency{nok}

	// Additional common currencies as defined by CLDR.
	BRL Currency = Currency{brl}
	CNY Currency = Currency{cny}
	DKK Currency = Currency{dkk}
	INR Currency = Currency{inr}
	RUB Currency = Currency{rub}
	HKD Currency = Currency{hkd}
	IDR Currency = Currency{idr}
	KRW Currency = Currency{krw}
	MXN Currency = Currency{mxn}
	PLN Currency = Currency{pln}
	SAR Currency = Currency{sar}
	THB Currency = Currency{thb}
	TRY Currency = Currency{try}
	TWD Currency = Currency{twd}
	ZAR Currency = Currency{zar}

	// Precious metals.
	XAG Currency = Currency{xag}
	XAU Currency = Currency{xau}
	XPT Currency = Currency{xpt}
	XPD Currency = Currency{xpd}
)
