// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package language implements BCP 47 language tags and related functionality.
//
// The Tag type, which is used to represent language tags, is agnostic to the
// meaning of its subtags. Tags are not fully canonicalized to preserve
// information that may be valuable in certain contexts. As a consequence, two
// different tags may represent identical languages in certain contexts.
//
// To determine equivalence between tags, a user should typically use a Matcher
// that is aware of the intricacies of equivalence within the given context.
// The default Matcher implementation provided in this package takes into
// account things such as deprecated subtags, legacy tags, and mutual
// intelligibility between scripts and languages.
//
// See http://tools.ietf.org/html/bcp47 for more details.
//
// NOTE: This package is still under development. Parts of it are not yet
// implemented, and the API is subject to change.
package language

import (
	"errors"
	"fmt"
	"strings"
)

var (
	// Und represents the undertermined language. It is also the root language tag.
	Und   Tag = und
	En    Tag = en    // Default language tag for English.
	En_US Tag = en_US // Default language tag for American English.
	De    Tag = de    // Default language tag for German.
	// TODO: list of most common language tags.
)

var (
	de    = Tag{lang: lang_de}
	en    = Tag{lang: lang_en}
	en_US = Tag{lang: lang_en, region: regUS}
	und   = Tag{}
)

// Tag represents a BCP 47 language tag. It is used to specifify
// an instance of a specific language or locale.
// All language tag values are guaranteed to be well-formed.
type Tag struct {
	// In most cases, just lang, region and script will be needed.  In such cases
	// str may be nil.
	lang     langID
	region   regionID
	script   scriptID
	pVariant byte   // offset in str, includes preceding '-'
	pExt     uint16 // offset of first extension, includes preceding '-'
	str      *string
}

// Make calls Parse and Canonicalize and returns the resulting Tag.
// Any errors are ignored and a sensible default is returned.
// In most cases, language tags should be created using this method.
func Make(id string) Tag {
	loc, _ := Parse(id)
	loc, _ = loc.Canonicalize(Default)
	return loc
}

// equalTags compares language, script and region subtags only.
func (t Tag) equalTags(a Tag) bool {
	return t.lang == a.lang && t.script == a.script && t.region == a.region
}

// IsRoot returns true if t is equal to language "und".
func (t Tag) IsRoot() bool {
	if t.str != nil {
		n := len(*t.str)
		if int(t.pVariant) < n {
			return false
		}
		t.str = nil
	}
	return t.equalTags(und)
}

// private reports whether the Tag consists solely of a private use tag.
func (t Tag) private() bool {
	return t.str != nil && t.pVariant == 0
}

// CanonType can be used to enable or disable various types of canonicalization.
type CanonType int

const (
	// Replace deprecated values with their preferred ones.
	Deprecated CanonType = 1 << iota
	// Remove redundant scripts.
	SuppressScript
	// Normalize legacy encodings, as defined by CLDR.
	Legacy
	// Map the dominant language of a macro language group to the macro language subtag.
	// For example cmn -> zh.
	Macro
	// The CLDR flag should be used if full compatibility with CLDR is required.  There are
	// a few cases where language.Tag may differ from CLDR.
	CLDR
	// All canonicalizations prescribed by BCP 47.
	BCP47   = Deprecated | SuppressScript
	All     = BCP47 | Legacy | Macro
	Default = All

	// TODO: LikelyScript, LikelyRegion: supress similar to ICU.
)

// canonicalize returns the canonicalized equivalent of the tag and
// whether there was any change.
func (t Tag) canonicalize(c CanonType) (Tag, bool) {
	changed := false
	if c&SuppressScript != 0 {
		if t.lang < langNoIndexOffset && uint8(t.script) == suppressScript[t.lang] {
			t.script = 0
			changed = true
		}
	}
	if c&Legacy != 0 {
		// We hard code this set as it is very small, unlikely to change and requires some
		// handling that does not fit elsewhere.
		switch t.lang {
		case lang_no:
			if c&CLDR != 0 {
				t.lang = lang_nb
				changed = true
			}
		case lang_tl:
			t.lang = lang_fil
			changed = true
		case lang_sh:
			if t.script == 0 {
				t.script = scrLatn
			}
			t.lang = lang_sr
			changed = true
		}
	}
	if c&Deprecated != 0 {
		l := normLang(langOldMap[:], t.lang)
		if l != t.lang {
			// CLDR maps "mo" to "ro". This mapping loses the piece of information
			// that "mo" very likely implies the region "MD". This may be important
			// for applications that insist on making a difference between these
			// two language codes.
			if t.lang == lang_mo && t.region == 0 && c&CLDR == 0 {
				t.region = regMD
			}
			changed = true
			t.lang = l
		}
		if t.script == scrQaai {
			changed = true
			t.script = scrZinh
		}
		if r := normRegion(t.region); r != 0 {
			changed = true
			t.region = r
		}
	}
	if c&Macro != 0 {
		// We deviate here from CLDR. The mapping "nb" -> "no" qualifies as a typical
		// Macro language mapping.  However, for legacy reasons, CLDR maps "no",
		// the macro language code for Norwegian, to the dominant variant "nb".
		// This change is currently under consideration for CLDR as well.
		// See http://unicode.org/cldr/trac/ticket/2698 and also
		// http://unicode.org/cldr/trac/ticket/1790 for some of the practical
		// implications.
		// TODO: this check could be removed if CLDR adopts this change.
		if c&CLDR == 0 || t.lang != lang_nb {
			l := normLang(langMacroMap[:], t.lang)
			if l != t.lang {
				changed = true
				t.lang = l
			}
		}
	}
	return t, changed
}

// Canonicalize returns the canonicalized equivalent of the tag.
func (t Tag) Canonicalize(c CanonType) (Tag, error) {
	t, changed := t.canonicalize(c)
	if changed && t.str != nil {
		t.remakeString()
	}
	return t, nil
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

// remakeString is used to update t.str in case lang, script or region changed.
// It is assumed that pExt and pVariant still point to the start of the
// respective parts, if applicable.
// remakeString can also be used to compute the string for Tag for which str
// is not defined.
func (t *Tag) remakeString() {
	extra := ""
	if t.str != nil && int(t.pVariant) < len(*t.str) {
		extra = (*t.str)[t.pVariant:]
		if t.pVariant > 0 {
			extra = extra[1:]
		}
		if t.equalTags(und) && strings.HasPrefix(extra, "x-") {
			t.str = &extra
			t.pVariant = 0
			t.pExt = 0
			return
		}
	}
	var buf [128]byte // avoid memory allocation for the vast majority of tags.
	b := buf[:t.genCoreBytes(buf[:])]
	if extra != "" {
		diff := uint8(len(b)) - t.pVariant
		b = append(b, '-')
		b = append(b, extra...)
		t.pVariant += diff
		t.pExt += uint16(diff)
	} else {
		t.pVariant = uint8(len(b))
		t.pExt = uint16(len(b))
	}
	s := string(b)
	t.str = &s
}

func (t *Tag) genCoreBytes(buf []byte) int {
	n := t.lang.stringToBuf(buf[:])
	if t.script != 0 {
		n += copy(buf[n:], "-")
		n += copy(buf[n:], t.script.String())
	}
	if t.region != 0 {
		n += copy(buf[n:], "-")
		n += copy(buf[n:], t.region.String())
	}
	return n
}

// String returns the canonical string representation of the language tag.
func (t Tag) String() string {
	if t.str == nil {
		if t.script == 0 && t.region == 0 {
			return t.lang.String()
		}
		buf := [16]byte{}
		return string(buf[:t.genCoreBytes(buf[:])])
	}
	return *t.str
}

// Base returns the base language of the language tag. If the base language is
// unspecified, an attempt will be made to infer it from the context.
// It uses a variant of CLDR's Add Likely Subtags algorithm. This is subject to change.
func (t Tag) Base() (Base, Confidence) {
	if t.lang != 0 {
		return Base{t.lang}, Exact
	}
	c := High
	if t.script == 0 && !(Region{t.region}).IsCountry() {
		c = Low
	}
	if tag, err := addTags(t); err == nil && tag.lang != 0 {
		return Base{tag.lang}, c
	}
	return Base{0}, No
}

// Script infers the script for the language tag. If it was not explictly given, it will infer
// a most likely candidate.
// If more than one script is commonly used for a language, the most likely one
// is returned with a low confidence indication. For example, it returns (Cyrl, Low)
// for Serbian.
// If a script cannot be inferred (Zzzz, No) is returned. We do not use Zyyy (undertermined)
// as one would suspect from the IANA registry for BCP 47. In a Unicode context Zyyy marks
// common characters (like 1, 2, 3, '.', etc.) and is therefore more like multiple scripts.
// See http://www.unicode.org/reports/tr24/#Values for more details. Zzzz is also used for
// unknown value in CLDR.  (Zzzz, Exact) is returned if Zzzz was explicitly specified.
// Note that an inferred script is never guaranteed to be the correct one. Latin is
// almost exclusively used for Afrikaans, but Arabic has been used for some texts
// in the past.  Also, the script that is commonly used may change over time.
// It uses a variant of CLDR's Add Likely Subtags algorithm. This is subject to change.
func (t Tag) Script() (Script, Confidence) {
	if t.script != 0 {
		return Script{t.script}, Exact
	}
	if t.lang < langNoIndexOffset {
		if sc := suppressScript[t.lang]; sc != 0 {
			return Script{scriptID(sc)}, High
		}
	}
	sc, c := Script{scrZzzz}, No
	if tag, err := addTags(t); err == nil {
		sc, c = Script{tag.script}, Low
	}
	t, _ = t.Canonicalize(Deprecated | Macro)
	if tag, err := addTags(t); err == nil {
		sc, c = Script{tag.script}, Low
	}
	return sc, c
}

// Region returns the region for the language tag. If it was not explicitly given, it will
// infer a most likely candidate from the context.
// It uses a variant of CLDR's Add Likely Subtags algorithm. This is subject to change.
func (t Tag) Region() (Region, Confidence) {
	if t.region != 0 {
		return Region{t.region}, Exact
	}
	if t, err := addTags(t); err == nil {
		return Region{t.region}, Low // TODO: differentiate between high and low.
	}
	t, _ = t.Canonicalize(Deprecated | Macro)
	if tag, err := addTags(t); err == nil {
		return Region{tag.region}, Low
	}
	return Region{regZZ}, No // TODO: return world instead of undetermined?
}

// Variant returns the variants specified explicitly for this language tag.
// or nil if no variant was specified.
func (t Tag) Variant() []Variant {
	// TODO: implement
	return nil
}

// TypeForKey returns the type associated with the given key, where key and type
// are of the allowed values defined for the Unicode locale extension ('u') in
// http://www.unicode.org/reports/tr35/#Unicode_Language_and_Locale_Identifiers.
// TypeForKey will traverse the inheritance chain to get the correct value.
func (t Tag) TypeForKey(key string) string {
	if start, end, _ := t.findTypeForKey(key); end != start {
		return (*t.str)[start:end]
	}
	return ""
}

var (
	errPrivateUse       = errors.New("cannot set a key on a private use tag")
	errInvalidArguments = errors.New("invalid key or type")
)

// SetTypeForKey returns a new Tag with the key set to type, where key and type
// are of the allowed values defined for the Unicode locale extension ('u') in
// http://www.unicode.org/reports/tr35/#Unicode_Language_and_Locale_Identifiers.
func (t Tag) SetTypeForKey(key, value string) (Tag, error) {
	if t.private() {
		return t, errPrivateUse
	}
	if len(key) != 2 || len(value) < 3 || len(value) > 8 {
		return t, errInvalidArguments
	}
	var (
		buf    [26]byte // enough to hold a core tag and simple -u extension
		uStart int      // start of the -u extension.
	)

	// Generate the tag string if needed.
	if t.str == nil {
		uStart = t.genCoreBytes(buf[:])
		buf[uStart] = '-'
		t.pVariant, t.pExt = byte(uStart), uint16(uStart)
		uStart++
	}

	// Create new key-type pair and parse it to verify.
	b := buf[uStart:]
	copy(b, "u-")
	copy(b[2:], key)
	b[4] = '-'
	b = b[:5+copy(b[5:], value)]
	scan := makeScanner(b)
	if parseExtensions(&scan); scan.err != nil {
		return t, scan.err
	}

	// Assemble the replacement string.
	s := ""
	if t.str == nil {
		s = string(buf[:uStart+len(b)])
	} else {
		s = *t.str
		start, end, hasExt := t.findTypeForKey(key)
		if start == end {
			if hasExt {
				b = b[2:]
			}
			s = fmt.Sprintf("%s-%s%s", s[:start], b, s[end:])
		} else {
			s = fmt.Sprintf("%s%s%s", s[:start], value, s[end:])
		}
	}
	t.str = &s
	return t, nil
}

// findKeyAndType returns the start and end position for the type corresponding
// to key or the point at which to insert the key-value pair if the type
// wasn't found. The hasExt return value reports whether an -u extension was present.
// Note: the extensions are typically very small and are likely to contain
// only one key-type pair.
func (t Tag) findTypeForKey(key string) (start, end int, hasExt bool) {
	p := int(t.pExt)
	if t.str == nil || len(key) != 2 || p == 0 || p == len(*t.str) {
		return p, p, false
	}
	s := *t.str

	// Find the correct extension.
	for p++; s[p] != 'u'; p++ {
		if s[p] > 'u' {
			p--
			return p, p, false
		}
		if p = nextExtension(s, p); p == len(s) {
			return len(s), len(s), false
		}
	}
	// Proceed to the hyphen following the extension name.
	p++

	// curKey is the key currently being processed.
	curKey := ""

	// Iterate over keys until we get the end of a section.
	for {
		// p points to the hyphen preceeding the current token.
		if p3 := p + 3; s[p3] == '-' {
			// Found a key.
			// Check whether we just processed the key that was requested.
			if curKey == key {
				return start, p, true
			}
			// Set to the next key and continue scanning type tokens.
			curKey = s[p+1 : p3]
			if curKey > key {
				return p, p, true
			}
			// Start of the type token sequence.
			start = p + 4
			// A type is at least 3 characters long.
			p += 7 // 4 + 3
		} else {
			// Attribute or type, which is at least 3 characters long.
			p += 4
		}
		// p points past the third character of a type or attribute.
		max := p + 5 // maximum length of token plus hyphen.
		if len(s) < max {
			max = len(s)
		}
		for ; p < max && s[p] != '-'; p++ {
		}
		// Bail if we have exhausted all tokens or if the next token starts
		// a new extension.
		if p == len(s) || s[p+2] == '-' {
			if curKey == key {
				return start, p, true
			}
			return p, p, true
		}
	}
}

// Base is an ISO 639 language code, used for encoding the base language
// of a language tag.
type Base struct {
	langID
}

// ParseBase parses a 2- or 3-letter ISO 639 code.
// It returns a ValueError if s is a well-formed but unknown language identifier
// or another error if another error occurred.
func ParseBase(s string) (Base, error) {
	if n := len(s); n < 2 || 3 < n {
		return Base{}, errSyntax
	}
	var buf [3]byte
	l, err := getLangID(buf[:copy(buf[:], s)])
	return Base{l}, err
}

// Tag returns a Tag with this base language as its only subtag.
func (b Base) Tag() Tag {
	return Tag{lang: b.langID}
}

// Script is a 4-letter ISO 15924 code for representing scripts.
// It is idiomatically represented in title case.
type Script struct {
	scriptID
}

// ParseScript parses a 4-letter ISO 15924 code.
// It returns a ValueError if s is a well-formed but unknown script identifier
// or another error if another error occurred.
func ParseScript(s string) (Script, error) {
	if len(s) != 4 {
		return Script{}, errSyntax
	}
	var buf [4]byte
	sc, err := getScriptID(script, buf[:copy(buf[:], s)])
	return Script{sc}, err
}

// Tag returns a Tag with the undetermined language and this script as its only subtags.
func (s Script) Tag() Tag {
	return Tag{script: s.scriptID}
}

// Region is an ISO 3166-1 or UN M.49 code for representing countries and regions.
type Region struct {
	regionID
}

// EncodeM49 returns the Region for the given UN M.49 code.
// It returns an error if r is not a valid code.
func EncodeM49(r int) (Region, error) {
	rid, err := getRegionM49(r)
	return Region{rid}, err
}

// ParseRegion parses a 2- or 3-letter ISO 3166-1 or a UN M.49 code.
// It returns a ValueError if s is a well-formed but unknown region identifier
// or another error if another error occurred.
func ParseRegion(s string) (Region, error) {
	if n := len(s); n < 2 || 3 < n {
		return Region{}, errSyntax
	}
	var buf [3]byte
	r, err := getRegionID(buf[:copy(buf[:], s)])
	return Region{r}, err
}

// Tag returns a Tag with the undetermined language and this region as its only subtags.
func (r Region) Tag() Tag {
	return Tag{region: r.regionID}
}

// IsCountry returns whether this region is a country or autonomous area.
func (r Region) IsCountry() bool {
	if r.regionID < isoRegionOffset || r.IsPrivateUse() {
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

// ParseCurrency parses a 3-letter ISO 4217 code.
// It returns a ValueError if s is a well-formed but unknown currency identifier
// or another error if another error occurred.
func ParseCurrency(s string) (Currency, error) {
	if len(s) != 3 {
		return Currency{}, errSyntax
	}
	var buf [3]byte
	c, err := getCurrencyID(currency, buf[:copy(buf[:], s)])
	return Currency{c}, err
}
