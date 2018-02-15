// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package language

import (
	"errors"
	"strconv"
	"strings"

	"golang.org/x/text/language/internal"
)

// ValueError is returned by any of the parsing functions when the
// input is well-formed but the respective subtag is not recognized
// as a valid value.
type ValueError interface {
	error

	// Subtag returns the subtag for which the error occurred.
	Subtag() string
}

// Parse parses the given BCP 47 string and returns a valid Tag. If parsing
// failed it returns an error and any part of the tag that could be parsed.
// If parsing succeeded but an unknown value was found, it returns
// ValueError. The Tag returned in this case is just stripped of the unknown
// value. All other values are preserved. It accepts tags in the BCP 47 format
// and extensions to this standard defined in
// http://www.unicode.org/reports/tr35/#Unicode_Language_and_Locale_Identifiers.
// The resulting tag is canonicalized using the default canonicalization type.
func Parse(s string) (t Tag, err error) {
	return Default.Parse(s)
}

// Parse parses the given BCP 47 string and returns a valid Tag. If parsing
// failed it returns an error and any part of the tag that could be parsed.
// If parsing succeeded but an unknown value was found, it returns
// ValueError. The Tag returned in this case is just stripped of the unknown
// value. All other values are preserved. It accepts tags in the BCP 47 format
// and extensions to this standard defined in
// http://www.unicode.org/reports/tr35/#Unicode_Language_and_Locale_Identifiers.
// The resulting tag is canonicalized using the the canonicalization type c.
func (c CanonType) Parse(s string) (t Tag, err error) {
	tt, err := language.Parse(s)
	if err != nil {
		return Tag{tt}, err
	}
	tt, changed := canonicalize(c, tt)
	if changed {
		tt.RemakeString()
	}
	return Tag{tt}, err
}

// Compose creates a Tag from individual parts, which may be of type Tag, Base,
// Script, Region, Variant, []Variant, Extension, []Extension or error. If a
// Base, Script or Region or slice of type Variant or Extension is passed more
// than once, the latter will overwrite the former. Variants and Extensions are
// accumulated, but if two extensions of the same type are passed, the latter
// will replace the former. A Tag overwrites all former values and typically
// only makes sense as the first argument. The resulting tag is returned after
// canonicalizing using the Default CanonType. If one or more errors are
// encountered, one of the errors is returned.
func Compose(part ...interface{}) (t Tag, err error) {
	return Default.Compose(part...)
}

// Compose creates a Tag from individual parts, which may be of type Tag, Base,
// Script, Region, Variant, []Variant, Extension, []Extension or error. If a
// Base, Script or Region or slice of type Variant or Extension is passed more
// than once, the latter will overwrite the former. Variants and Extensions are
// accumulated, but if two extensions of the same type are passed, the latter
// will replace the former. A Tag overwrites all former values and typically
// only makes sense as the first argument. The resulting tag is returned after
// canonicalizing using CanonType c. If one or more errors are encountered,
// one of the errors is returned.
func (c CanonType) Compose(part ...interface{}) (t Tag, err error) {
	var b language.Builder
	if err = update(&b, part...); err != nil {
		return und, err
	}
	b.Tag, _ = canonicalize(c, b.Tag)
	return Tag{b.Make()}, err
}

var errInvalidArgument = errors.New("invalid Extension or Variant")

func update(b *language.Builder, part ...interface{}) (err error) {
	replace := func(l *[]string, s string, eq func(a, b string) bool) bool {
		if s == "" {
			b.Err = errInvalidArgument
			return true
		}
		for i, v := range *l {
			if eq(v, s) {
				(*l)[i] = s
				return true
			}
		}
		return false
	}
	for _, x := range part {
		switch v := x.(type) {
		case Tag:
			b.Tag.LangID = v.lang()
			b.Tag.RegionID = v.region()
			b.Tag.ScriptID = v.script()
			// TODO: optimize
			b.Variant = b.Variant[:0]
			for _, vr := range v.Variants() {
				b.Variant = append(b.Variant, vr.String())
			}
			b.Ext, b.Private = b.Ext[:0], ""
			for _, e := range v.Extensions() {
				b.AddExt(e.String())
			}
		case Base:
			b.Tag.LangID = v.langID
		case Script:
			b.Tag.ScriptID = v.scriptID
		case Region:
			b.Tag.RegionID = v.regionID
		case Variant:
			if !replace(&b.Variant, v.variant, func(a, b string) bool { return a == b }) {
				b.Variant = append(b.Variant, v.variant)
			}
		case Extension:
			if !replace(&b.Ext, v.s, func(a, b string) bool { return a[0] == b[0] }) {
				b.AddExt(v.s)
			}
		case []Variant:
			b.Variant = nil
			for _, x := range v {
				update(b, x)
			}
		case []Extension:
			b.Ext, b.Private = nil, ""
			for _, e := range v {
				update(b, e)
			}
		// TODO: support parsing of raw strings based on morphology or just extensions?
		case error:
			err = v
		}
	}
	return
}

var errInvalidWeight = errors.New("ParseAcceptLanguage: invalid weight")

// ParseAcceptLanguage parses the contents of an Accept-Language header as
// defined in http://www.ietf.org/rfc/rfc2616.txt and returns a list of Tags and
// a list of corresponding quality weights. It is more permissive than RFC 2616
// and may return non-nil slices even if the input is not valid.
// The Tags will be sorted by highest weight first and then by first occurrence.
// Tags with a weight of zero will be dropped. An error will be returned if the
// input could not be parsed.
func ParseAcceptLanguage(s string) (tag []Tag, q []float32, err error) {
	var entry string
	for s != "" {
		if entry, s = split(s, ','); entry == "" {
			continue
		}

		entry, weight := split(entry, ';')

		// Scan the language.
		t, err := Parse(entry)
		if err != nil {
			id, ok := acceptFallback[entry]
			if !ok {
				return nil, nil, err
			}
			t = Tag{tag: language.Tag{LangID: id}}
		}

		// Scan the optional weight.
		w := 1.0
		if weight != "" {
			weight = consume(weight, 'q')
			weight = consume(weight, '=')
			// consume returns the empty string when a token could not be
			// consumed, resulting in an error for ParseFloat.
			if w, err = strconv.ParseFloat(weight, 32); err != nil {
				return nil, nil, errInvalidWeight
			}
			// Drop tags with a quality weight of 0.
			if w <= 0 {
				continue
			}
		}

		tag = append(tag, t)
		q = append(q, float32(w))
	}
	sortStable(&tagSort{tag, q})
	return tag, q, nil
}

// consume removes a leading token c from s and returns the result or the empty
// string if there is no such token.
func consume(s string, c byte) string {
	if s == "" || s[0] != c {
		return ""
	}
	return strings.TrimSpace(s[1:])
}

func split(s string, c byte) (head, tail string) {
	if i := strings.IndexByte(s, c); i >= 0 {
		return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+1:])
	}
	return strings.TrimSpace(s), ""
}

// Add hack mapping to deal with a small number of cases that that occur
// in Accept-Language (with reasonable frequency).
var acceptFallback = map[string]language.Language{
	"english": _en,
	"deutsch": _de,
	"italian": _it,
	"french":  _fr,
	"*":       _mul, // defined in the spec to match all languages.
}

type tagSort struct {
	tag []Tag
	q   []float32
}

func (s *tagSort) Len() int {
	return len(s.q)
}

func (s *tagSort) Less(i, j int) bool {
	return s.q[i] > s.q[j]
}

func (s *tagSort) Swap(i, j int) {
	s.tag[i], s.tag[j] = s.tag[j], s.tag[i]
	s.q[i], s.q[j] = s.q[j], s.q[i]
}
