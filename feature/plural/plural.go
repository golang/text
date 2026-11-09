// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run gen.go gen_common.go

// Package plural provides utilities for handling linguistic plurals in text.
//
// The definitions in this package are based on the plural rule handling defined
// in CLDR. See
// https://unicode.org/reports/tr35/tr35-numbers.html#Language_Plural_Rules for
// details.
package plural

import (
	"golang.org/x/text/internal/language/compact"
	"golang.org/x/text/internal/number"
	"golang.org/x/text/language"
)

// Rules defines the plural rules for all languages for a certain plural type.
//
// This package is UNDER CONSTRUCTION and its API may change.
type Rules struct {
	rules          []pluralCheck
	index          []byte
	langToIndex    []byte
	inclusionMasks []uint64
}

var (
	// Cardinal defines the plural rules for numbers indicating quantities.
	Cardinal *Rules = cardinal

	// Ordinal defines the plural rules for numbers indicating position
	// (first, second, etc.).
	Ordinal *Rules = ordinal

	ordinal = &Rules{
		ordinalRules,
		ordinalIndex,
		ordinalLangToIndex,
		ordinalInclusionMasks[:],
	}

	cardinal = &Rules{
		cardinalRules,
		cardinalIndex,
		cardinalLangToIndex,
		cardinalInclusionMasks[:],
	}
)

// getIntApprox converts the digits in slice digits[start:end] to an integer
// according to the following rules:
//   - Let i be asInt(digits[start:end]), where out-of-range digits are assumed
//     to be zero.
//   - Result n is big if i / 10^nMod > 1.
//   - Otherwise the result is i % 10^nMod.
//
// For example, if digits is {1, 2, 3} and start:end is 0:5, then the result
// for various values of nMod is:
//   - when nMod == 2, n == big
//   - when nMod == 3, n == big
//   - when nMod == 4, n == big
//   - when nMod == 5, n == 12300
//   - when nMod == 6, n == 12300
//   - when nMod == 7, n == 12300
func getIntApprox(digits []byte, start, end, nMod, big int) (n int) {
	// Leading 0 digits just result in 0.
	p := start
	if p < 0 {
		p = 0
	}
	// Range only over the part for which we have digits.
	mid := end
	if mid >= len(digits) {
		mid = len(digits)
	}
	// Check digits more significant that nMod.
	if q := end - nMod; q > 0 {
		if q > mid {
			q = mid
		}
		for ; p < q; p++ {
			if digits[p] != 0 {
				return big
			}
		}
	}
	for ; p < mid; p++ {
		n = 10*n + int(digits[p])
	}
	// Multiply for trailing zeros.
	for ; p < end; p++ {
		n *= 10
	}
	return n
}

// MatchDigits computes the plural form for the given language and the given
// decimal floating point digits. The digits are stored in big-endian order and
// are of value byte(0) - byte(9). The floating point position is indicated by
// exp and the number of visible decimals is scale. All leading and trailing
// zeros may be omitted from digits.
//
// The following table contains examples of possible arguments to represent
// the given numbers.
//
//	decimal    digits              exp    scale
//	123        []byte{1, 2, 3}     3      0
//	123.4      []byte{1, 2, 3, 4}  3      1
//	123.40     []byte{1, 2, 3, 4}  3      2
//	100000     []byte{1}           6      0
//	100000.00  []byte{1}           6      3
func (p *Rules) MatchDigits(t language.Tag, digits []byte, exp, scale int) Form {
	index := tagToID(t)

	// Differentiate up to including mod 1000000 for the integer part.
	n := getPluralInteger(digits, exp)

	// Differentiate up to including mod 100 for the fractional part.
	f := getIntApprox(digits, exp, exp+scale, 2, 100)

	return matchPlural(p, index, n, f, scale, fractionalWithoutTrailingZeros(digits, exp, scale))
}

// fractionalWithoutTrailingZeros returns t modulo 100: the fractional digits
// without trailing zeros. Only modulo 10 and 100 comparisons use t.
// getPluralInteger retains the last six integer digits of large numbers
// while marking values >= 1000000. This distinguishes multiples of a
// million from nearby integers without losing the existing small-number
// comparison semantics used by matchPlural.
func getPluralInteger(digits []byte, exp int) int {
	const million = 1000000
	if exp <= 6 {
		return getIntApprox(digits, 0, exp, 6, million)
	}
	end := exp
	if end > len(digits) {
		end = len(digits)
	}
	mod := 0
	big := false
	for i := 0; i < end; i++ {
		if i < exp-6 && digits[i] != 0 {
			big = true
		}
		mod = (mod*10 + int(digits[i])) % million
	}
	if exp-end >= 6 {
		mod = 0
	} else {
		for i := end; i < exp; i++ {
			mod = (mod * 10) % million
		}
	}
	if big {
		mod += million
	}
	return mod
}

func fractionalWithoutTrailingZeros(digits []byte, exp, scale int) int {
	end := exp + scale
	if end > len(digits) {
		end = len(digits)
	}
	for end > exp && end > 0 && digits[end-1] == 0 {
		end--
	}
	start := exp
	if start < 0 {
		start = 0
	}
	t := 0
	for i := start; i < end; i++ {
		t = (t*10 + int(digits[i])) % maxMod
	}
	return t
}

func (p *Rules) matchDisplayDigits(t language.Tag, d *number.Digits) (Form, int) {
	n := getIntApprox(d.Digits, 0, int(d.Exp), 6, 1000000)
	return p.MatchDigits(t, d.Digits, int(d.Exp), d.NumFracDigits()), n
}

func validForms(p *Rules, t language.Tag) (forms []Form) {
	offset := p.langToIndex[tagToID(t)]
	rules := p.rules[p.index[offset]:p.index[offset+1]]

	forms = append(forms, Other)
	last := Other
	for _, r := range rules {
		if cat := Form(r.cat & formMask); cat != andNext && last != cat {
			forms = append(forms, cat)
			last = cat
		}
	}
	return forms
}

func (p *Rules) matchComponents(t language.Tag, n, f, scale int) Form {
	truncated := f
	for truncated > 0 && truncated%10 == 0 {
		truncated /= 10
	}
	return matchPlural(p, tagToID(t), n, f, scale, truncated)
}

// MatchPlural returns the plural form for the given language and plural
// operands (as defined in
// https://unicode.org/reports/tr35/tr35-numbers.html#Language_Plural_Rules):
//
//	where
//		n  absolute value of the source number (integer and decimals)
//	input
//		i  integer digits of n.
//		v  number of visible fraction digits in n, with trailing zeros.
//		w  number of visible fraction digits in n, without trailing zeros.
//		f  visible fractional digits in n, with trailing zeros (f = t * 10^(v-w))
//		t  visible fractional digits in n, without trailing zeros.
//
// If any of the operand values is too large to fit in an int, it is okay to
// pass the value modulo 10,000,000.
func (p *Rules) MatchPlural(lang language.Tag, i, v, w, f, t int) Form {
	return matchPlural(p, tagToID(lang), i, f, v, t)
}

func matchPlural(p *Rules, index compact.ID, n, f, v, t int) Form {
	nMask := p.inclusionMasks[n%maxMod]
	// Compute the fMask inline in the rules below, as it is relatively rare.
	// fMask := p.inclusionMasks[f%maxMod]
	vMask := p.inclusionMasks[v%maxMod]

	// Do the matching
	offset := p.langToIndex[index]
	rules := p.rules[p.index[offset]:p.index[offset+1]]
	for i := 0; i < len(rules); i++ {
		rule := rules[i]
		setBit := uint64(1 << rule.setID)
		var skip bool
		switch op := opID(rule.cat >> opShift); op {
		case opI: // i = x
			skip = n >= numN || nMask&setBit == 0

		case opI | opNotEqual: // i != x
			skip = n < numN && nMask&setBit != 0

		case opI | opMod: // i % m = x
			skip = nMask&setBit == 0

		case opI | opMod | opNotEqual: // i % m != x
			skip = nMask&setBit != 0

		case opN: // n = x
			skip = f != 0 || n >= numN || nMask&setBit == 0

		case opN | opNotEqual: // n != x
			skip = f == 0 && n < numN && nMask&setBit != 0

		case opN | opMod: // n % m = x
			skip = f != 0 || nMask&setBit == 0

		case opN | opMod | opNotEqual: // n % m != x
			skip = f == 0 && nMask&setBit != 0

		case opF: // f = x
			skip = f >= numN || p.inclusionMasks[f%maxMod]&setBit == 0

		case opF | opNotEqual: // f != x
			skip = f < numN && p.inclusionMasks[f%maxMod]&setBit != 0

		case opF | opMod: // f % m = x
			skip = p.inclusionMasks[f%maxMod]&setBit == 0

		case opF | opMod | opNotEqual: // f % m != x
			skip = p.inclusionMasks[f%maxMod]&setBit != 0

		case opV: // v = x
			skip = v < numN && vMask&setBit == 0

		case opV | opNotEqual: // v != x
			skip = v < numN && vMask&setBit != 0

		case opW: // w == 0
			skip = f != 0

		case opW | opNotEqual: // w != 0
			skip = f == 0

		case opW | opMod: // t % 10 or t % 100 = x
			skip = p.inclusionMasks[t%maxMod]&setBit == 0

		case opW | opMod | opNotEqual: // t % 10 or t % 100 != x
			skip = p.inclusionMasks[t%maxMod]&setBit != 0

		// Hard-wired rules that cannot be handled by our algorithm.

		case opBretonM:
			skip = f != 0 || n == 0 || n%1000000 != 0

		case opBretonM | opNotEqual: // CLDR 49 e = 0, i multiple of a million, v = 0
			skip = v != 0 || n == 0 || n%1000000 != 0

		case opAzerbaijan00s:
			// 100,200,300,400,500,600,700,800,900
			skip = n == 0 || n >= 1000 || n%100 != 0

		case opAzerbaijan00s | opMod: // Cornish: n % 1000 = 0 and selected n % 100000
			r := n % 100000
			skip = f != 0 || n%1000 != 0 || !(r >= 1000 && r <= 20000 || r == 40000 || r == 60000 || r == 80000)

		case opAzerbaijan00s | opNotEqual: // Cornish: n != 0 and n % 1000000 = 100000
			skip = f != 0 || n == 0 || n%1000000 != 100000

		case opItalian800:
			skip = (f != 0 || n >= numN || nMask&setBit == 0) && n != 800

		case opItalian800 | opNotEqual: // Ligurian and Sicilian: 800..899
			skip = (f != 0 || n >= numN || nMask&setBit == 0) && (n < 800 || n > 899)
		}
		if skip {
			// advance over AND entries.
			for ; i < len(rules) && rules[i].cat&formMask == andNext; i++ {
			}
			continue
		}
		// return if we have a final entry.
		if cat := rule.cat & formMask; cat != andNext {
			return Form(cat)
		}
	}
	return Other
}

func tagToID(t language.Tag) compact.ID {
	id, _ := compact.RegionalID(compact.Tag(t))
	return id
}
