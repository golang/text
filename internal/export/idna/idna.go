// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run gen.go gen_trieval.go gen_common.go

// http://www.unicode.org/reports/tr46

// Package idna implements IDNA2008 using the compatibility processing
// defined by UTS (Unicode Technical Standard) #46, which defines a standard to
// deal with the transition from IDNA2003.
//
// IDNA2008 (Internationalized Domain Names for Applications), is defined in RFC
// 5890, RFC 5891, RFC 5892, RFC 5893 and RFC 5894.
// UTS #46 is defined in http://www.unicode.org/reports/tr46.
// See http://unicode.org/cldr/utility/idna.jsp for a visualization of the
// differences between these two standards.
package idna // import "golang.org/x/text/internal/export/idna"

import (
	"errors"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// A Profile defines the configuration of a IDNA mapper.
type Profile struct {
	Transitional    bool
	IgnoreSTD3Rules bool
	// ErrHandler      func(error)
}

// String reports a string with a description of the profile for debugging
// purposes. The string format may change with different versions.
func (p *Profile) String() string {
	s := ""
	if p.Transitional {
		s = "Transitional"
	}
	s = "NonTraditional"
	if p.IgnoreSTD3Rules {
		s += ":NoSTD3Rules"
	}
	return s
}

var (
	// Resolve is the recommended profile for resolving domain names.
	// The configuration of this profile may change over time.
	Resolve = resolve

	// Transitional defines a profile that implements the Transitional mapping
	// as defined in UTS #46 with no additional constraints.
	Transitional = transitional

	// NonTransitional defines a profile that implements the Transitional
	// mapping as defined in UTS #46 with no additional constraints.
	NonTransitional = nonTransitional

	resolve         = &Profile{Transitional: true}
	transitional    = &Profile{Transitional: true}
	nonTransitional = &Profile{}

	// TODO: profiles
	// V2008: strict IDNA2008
	// Registrar: recommended for approving domain names.
)

var (
	// ErrDisallowed indicates a domain name contains a disallowed rune.
	ErrDisallowed = errors.New("idna: disallowed rune")
)

// process implements the algorithm described in section 4 of UTS #46,
// see http://www.unicode.org/reports/tr46.
func (p *Profile) process(s string, toASCII bool) (string, error) {
	var (
		b    []byte
		err  error
		k, i int
	)
	for i < len(s) {
		v, sz := trie.lookupString(s[i:])
		start := i
		i += sz
		// Copy bytes not copied so far.
		switch p.simplify(info(v).category()) {
		case valid:
			continue
		case disallowed:
			if err == nil {
				err = ErrDisallowed
			}
			continue
		case mapped, deviation:
			b = append(b, s[k:start]...)
			b = info(v).appendMapping(b, s[start:i])
		case ignored:
			b = append(b, s[k:start]...)
			// drop the rune
		case unknown:
			b = append(b, s[k:start]...)
			b = append(b, "\ufffd"...)
		}
		k = i
	}
	if k == 0 {
		// No changes so far.
		s = norm.NFC.String(s)
	} else {
		b = append(b, s[k:]...)
		if norm.NFC.QuickSpan(b) != len(b) {
			b = norm.NFC.Bytes(b)
		}
		// TODO: the punycode converters requires strings as input.
		s = string(b)
	}
	// TODO(perf): don't split.
	labels := strings.Split(s, ".")
	// Remove leading empty labels
	for len(labels) > 0 && labels[0] == "" {
		labels = labels[1:]
	}
	for i, label := range labels {
		transitional := p.Transitional
		if strings.HasPrefix(label, acePrefix) {
			u, err := decode(label[len(acePrefix):])
			if err != nil {
				// TODO: really return here? Probably not!
				return "", err
			}
			labels[i] = u
			transitional = false
		}
		if err == nil {
			err = validate(labels[i], transitional)
		}
	}
	// TODO(perf): do quick check for ascii
	if toASCII {
		for i, label := range labels {
			if ascii(label) {
				continue
			}
			a, err := encode(acePrefix, label)
			if err != nil {
				// TODO: really return here? Probably not!
				return "", err
			}
			labels[i] = a
		}
	}
	return strings.Join(labels, "."), err
}

// acePrefix is the ASCII Compatible Encoding prefix.
const acePrefix = "xn--"

func (p *Profile) simplify(cat category) category {
	switch cat {
	case disallowedSTD3Mapped:
		if !p.IgnoreSTD3Rules {
			cat = disallowed
		} else {
			cat = mapped
		}
	case disallowedSTD3Valid:
		if !p.IgnoreSTD3Rules {
			cat = disallowed
		} else {
			cat = valid
		}
	case deviation:
		if !p.Transitional {
			cat = valid
		}
	case validNV8, validXV8:
		// TODO: handle V2008
		cat = valid
	}
	return cat
}

func validate(s string, transitional bool) error {
	return nil
}

func (p *Profile) ToASCII(s string) (string, error) {
	return p.process(s, true)
}

func (p *Profile) ToUnicode(s string) (string, error) {
	return NonTransitional.process(s, false)
}

func ascii(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= utf8.RuneSelf {
			return false
		}
	}
	return true
}
