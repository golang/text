// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package precis

import "errors"

// This file contains tables and code related to context rules.

type catBitmap uint16

const (
	// These bits, once set depending on the current value, are never unset.
	bJapanese catBitmap = 1 << iota
	bArabicIndicDigit
	bExtendedArabicIndicDigit

	// These bits are set on each iteration depending on the current value.
	bJoinStart
	bJoinMid
	bJoinEnd
	bVirama
	bLatinSmallL
	bGreek
	bHebrew

	// These bits indicated which of the permanent bits need to be set at the
	// end of the checks.
	bMustHaveJapn

	permanent = bJapanese | bArabicIndicDigit | bExtendedArabicIndicDigit | bMustHaveJapn
)

const finalShift = 10

var errContext = errors.New("precis: contextual rule violated")

func init() {
	// Programmatically set these required bits as, manually setting them seems
	// too error prone.
	for i := range categoryTransitions {
		categoryTransitions[i].keep |= permanent
	}
}

var categoryTransitions = []struct {
	keep catBitmap // mask selecting which bits to keep from the previous state
	set  catBitmap // mask for which bits to set for this transition

	// The rule function cannot take a *context as an argument, as it would
	// cause the context to escape, adding significant overhead.
	rule func(beforeBits catBitmap) error
}{
	joiningL:          {set: bJoinStart},
	joiningD:          {set: bJoinStart | bJoinEnd},
	joiningT:          {keep: bJoinStart, set: bJoinMid},
	joiningR:          {set: bJoinEnd},
	viramaModifier:    {set: bVirama},
	viramaJoinT:       {set: bVirama | bJoinMid},
	latinSmallL:       {set: bLatinSmallL},
	greek:             {set: bGreek},
	greekJoinT:        {set: bGreek | bJoinMid},
	hebrew:            {set: bHebrew},
	hebrewJoinT:       {set: bHebrew | bJoinMid},
	japanese:          {set: bJapanese},
	katakanaMiddleDot: {set: bMustHaveJapn},

	zeroWidthNonJoiner: {
		rule: func(before catBitmap) error {
			if before&bVirama != 0 {
				return nil
			}
			if before&bJoinStart == 0 {
				return errContext
			}
			// TODO: check future.
			return nil
		},
	},
	zeroWidthJoiner: {
		rule: func(before catBitmap) error {
			if before&bVirama == 0 {
				return errContext
			}
			return nil
		},
	},
	middleDot: {
		rule: func(before catBitmap) error {
			if before&bLatinSmallL == 0 {
				return errContext
			}
			// TODO: check future.
			return nil
		},
	},
	greekLowerNumeralSign: {
		set: bGreek,
		rule: func(before catBitmap) error {
			// TODO: check future.
			return nil
		},
	},
	hebrewPreceding: {
		set: bHebrew,
		rule: func(before catBitmap) error {
			if before&bHebrew == 0 {
				return errContext
			}
			return nil
		},
	},
	arabicIndicDigit: {
		set: bArabicIndicDigit,
		rule: func(before catBitmap) error {
			if before&bExtendedArabicIndicDigit != 0 {
				return errContext
			}
			return nil
		},
	},
	extendedArabicIndicDigit: {
		set: bExtendedArabicIndicDigit,
		rule: func(before catBitmap) error {
			if before&bArabicIndicDigit != 0 {
				return errContext
			}
			return nil
		},
	},
}
