// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// NOTE: This package is still under development. Parts of it are not yet implemented,
// and the API is subject to change.
//
// The locale package provides a type to represent BCP 47 locale identifiers.
// It supports various canonicalizations defined in CLDR.
package locale

import "strings"

var (
	// Und represents the undefined langauge. It is also the root locale.
	Und   = und
	En    = en    // Default Locale for English.
	En_US = en_US // Default locale for American English.
	De    = de    // Default locale for German.
	// TODO: list of most common language identifiers.
)

var (
	Supported Set // All supported locales.
	Common    Set // A selection of common locales.
)

var (
	de    = ID{lang: lang_de}
	en    = ID{lang: lang_en}
	en_US = ID{lang: lang_en, region: regUS}
	und   = ID{}
)

// ID represents a BCP 47 locale identifier. It can be used to
// select an instance for a specific locale. All Locale values are guaranteed
// to be well-formed.
type ID struct {
	// In most cases, just lang, region and script will be needed.  In such cases
	// str may be nil.
	lang     langID
	region   regionID
	script   scriptID
	pVariant byte   // offset in str, includes preceding '-'
	pExt     uint16 // offset of first extension, includes preceding '-'
	str      *string
}

// Make calls Parse and Canonicalize and returns the resulting ID.
// Any errors are ignored and a sensible default is returned.
// In most cases, locale IDs should be created using this method.
func Make(id string) ID {
	loc, _ := Parse(id)
	loc, _ = loc.Canonicalize(All)
	return loc
}

// equalTags compares language, script and region identifiers only.
func (loc ID) equalTags(id ID) bool {
	return loc.lang == id.lang && loc.script == id.script && loc.region == id.region
}

// IsRoot returns true if loc is equal to locale "und".
func (loc ID) IsRoot() bool {
	if loc.str != nil {
		n := len(*loc.str)
		if int(loc.pVariant) < n {
			return false
		}
		loc.str = nil
	}
	return loc.equalTags(und)
}

// private reports whether the ID consists solely of a private use tag.
func (loc ID) private() bool {
	return loc.str != nil && loc.pVariant == 0
}

// CanonType is can be used to enable or disable various types of canonicalization.
type CanonType int

const (
	// Replace deprecated values with their preferred ones.
	Deprecated CanonType = 1 << iota
	// Remove redundant scripts.
	SuppressScript
	// Map the dominant language of macro language group to the macro language identifier.
	// For example cmn -> zh.
	Macro
	// All canonicalizations prescribed by BCP 47.
	BCP47 = Deprecated | SuppressScript
	All   = BCP47 | Macro

	// TODO: LikelyScript, LikelyRegion: supress similar to ICU.
)

// Canonicalize replaces the identifier with its canonical equivalent.
func (loc ID) Canonicalize(t CanonType) (ID, error) {
	changed := false
	if t&SuppressScript != 0 {
		if loc.lang < langNoIndexOffset && uint8(loc.script) == suppressScript[loc.lang] {
			loc.script = 0
			changed = true
		}
	}
	if t&Deprecated != 0 {
		l := normLang(langOldMap[:], loc.lang)
		if l != loc.lang {
			changed = true
		}
		loc.lang = l
	}
	if t&Macro != 0 {
		l := normLang(langMacroMap[:], loc.lang)
		if l != loc.lang {
			changed = true
		}
		loc.lang = l
	}
	if changed && loc.str != nil {
		loc.remakeString()
	}
	return loc, nil
}

// Confidence indicates the level of certainty for a given return value.
// For example, Serbian may be written in cyrillic or latin script.
// The confidence level indicates whether a value was explicitly specified,
// whether it is typically the only possible value, or whether there is
// an ambiguity.
type Confidence int

const (
	No    Confidence = iota // full confidence that there was no match
	Low                     // most likely value picked out of a set of alternatives
	High                    // value is generally assumed to be the correct match
	Exact                   // exact match or explicitly specified value
)

var confName = []string{"No", "Low", "High", "Exact"}

func (c Confidence) String() string {
	return confName[c]
}

// remakeString is used to update loc.str in case lang, script or region changed.
// It is assumed that pExt and pVariant still point to the start of the
// respective parts, if applicable.
// remakeString can also be used to compute the string for IDs for which str
// is not defined.
func (loc *ID) remakeString() {
	extra := ""
	if loc.str != nil && int(loc.pVariant) < len(*loc.str) {
		extra = (*loc.str)[loc.pVariant:]
		if loc.pVariant > 0 {
			extra = extra[1:]
		}
	}
	buf := [128]byte{}
	isUnd := loc.lang == 0
	n := loc.lang.stringToBuf(buf[:])
	if loc.script != 0 {
		n += copy(buf[n:], "-")
		n += copy(buf[n:], loc.script.String())
		isUnd = false
	}
	if loc.region != 0 {
		n += copy(buf[n:], "-")
		n += copy(buf[n:], loc.region.String())
		isUnd = false
	}
	b := buf[:n]
	if extra != "" {
		if isUnd && strings.HasPrefix(extra, "x-") {
			loc.str = &extra
			loc.pVariant = 0
			loc.pExt = 0
			return
		} else {
			diff := uint8(n) - loc.pVariant
			b = append(b, '-')
			b = append(b, extra...)
			loc.pVariant += diff
			loc.pExt += uint16(diff)
		}
	} else {
		loc.pVariant = uint8(len(b))
		loc.pExt = uint16(len(b))
	}
	s := string(b)
	loc.str = &s
}

// String returns the canonical string representation of the locale.
func (loc ID) String() string {
	if loc.str == nil {
		loc.remakeString()
	}
	return *loc.str
}

// Language returns the language for the locale. If the language is unspecified,
// an attempt will be made to infer it from the context.
// It uses a variant of CLDR's Add Likely Subtags algorithm. This is subject to change.
func (loc ID) Language() (Language, Confidence) {
	if loc.lang != 0 {
		return Language{loc.lang}, Exact
	}
	c := High
	if loc.script == 0 && !(Region{loc.region}).IsCountry() {
		c = Low
	}
	if id, err := addTags(loc); err == nil && id.lang != 0 {
		return Language{id.lang}, c
	}
	return Language{0}, No
}

// Script infers the script for the locale.  If it was not explictly given, it will infer
// a most likely candidate from the parent locales.
// If more than one script is commonly used for a language, the most likely one
// is returned with a low confidence indication. For example, it returns (Cyrl, Low)
// for Serbian.
// Note that an inferred script is never guaranteed to be the correct one. Latn is
// almost exclusively used for Afrikaans, but Arabic has been used for some texts
// in the past.  Also, the script that is commonly used may change over time.
// It uses a variant of CLDR's Add Likely Subtags algorithm. This is subject to change.
func (loc ID) Script() (Script, Confidence) {
	if loc.script != 0 {
		return Script{loc.script}, Exact
	}
	if loc.lang < langNoIndexOffset {
		if sc := suppressScript[loc.lang]; sc != 0 {
			return Script{scriptID(sc)}, High
		}
	}
	sc, c := Script{scrZyyy}, No
	if id, err := addTags(loc); err == nil {
		sc, c = Script{id.script}, Low
	}
	loc, _ = loc.Canonicalize(Deprecated | Macro)
	if id, err := addTags(loc); err == nil {
		sc, c = Script{id.script}, Low
	}
	// Translate srcZzzz (uncoded) to srcZyyy (undetermined).
	if sc == (Script{scrZzzz}) {
		return Script{scrZyyy}, No
	}
	return sc, c
}

// Region returns the region for the locale.  If it was not explicitly given, it will
// infer a most likely candidate from the context.
// It uses a variant of CLDR's Add Likely Subtags algorithm. This is subject to change.
func (loc ID) Region() (Region, Confidence) {
	if loc.region != 0 {
		return Region{loc.region}, Exact
	}
	if id, err := addTags(loc); err == nil {
		return Region{id.region}, Low // TODO: differentiate between high and low.
	}
	loc, _ = loc.Canonicalize(Deprecated | Macro)
	if id, err := addTags(loc); err == nil {
		return Region{id.region}, Low
	}
	return Region{regZZ}, No // TODO: return world instead of undetermined?
}

// Variant returns the variant specified explicitly for this locale
// or nil if no variant was specified.
func (loc ID) Variant() Variant {
	// TODO: implement
	return Variant{""}
}

// TypeForKey returns the type associated with the given key, where key and type
// are of the allowed values defined for the Unicode locale extension ('u') in
// http://www.unicode.org/reports/tr35/#Unicode_Language_and_Locale_Identifiers.
// TypeForKey will traverse the inheritance chain to get the correct value.
func (loc ID) TypeForKey(key string) string {
	// TODO: implement
	return ""
}

// SetTypeForKey returns a new ID with the key set to type, where key and type
// are of the allowed values defined for the Unicode locale extension ('u') in
// http://www.unicode.org/reports/tr35/#Unicode_Language_and_Locale_Identifiers.
func (loc ID) SetTypeForKey(key, value string) ID {
	// TODO: implement
	return ID{}
}

// Language is an ISO 639 language identifier.
type Language struct {
	langID
}

// Script is a 4-letter ISO 15924 code for representing scripts.
// It is idiomatically represented in title case.
type Script struct {
	scriptID
}

// Region is an ISO 3166-1 or UN M.49 code for representing countries and regions.
type Region struct {
	regionID
}

// IsCountry returns whether this region is a country or autonomous area.
func (r Region) IsCountry() bool {
	const m49PrivateUseStart = 900
	if r.regionID < isoRegionOffset || r.m49() >= m49PrivateUseStart {
		return false
	}
	return true
}

// Variant represents a registered variant of a language as defined by BCP 47.
type Variant struct {
	// TODO: implement
	variant string
}

// String returns the string representation of the variant.
func (v Variant) String() string {
	// TODO: implement
	return v.variant
}

// Currency is an ISO 4217 currency designator.
type Currency struct {
	currencyID
}

// Set provides information about a set of locales.
type Set interface {
	Locales() []ID
	Languages() []Language
	Regions() []Region
	Scripts() []Script
	Currencies() []Currency
}
