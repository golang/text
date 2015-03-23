// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package search provides language-specific search and string matching.
//
// Natural language matching can be intricate. For example, Danish will insist
// "Århus" and "Aarhus" are the same name and Turkish will match I to ı (note
// the lack of a dot) in a case-insensitive match. This package handles such
// language-specific details.
//
// Text passed to any of the calls in this message does not need to be
// normalized.
package search

import (
	"golang.org/x/text/collate/colltab"
	"golang.org/x/text/language"
)

// An Option configures a Matcher.
type Option func(*Matcher)

var (
	// WholeWord restricts matches to complete words. The default is to match at
	// the character level.
	WholeWord Option = nil

	// Exact requires that two strings are their exact equivalent. For example
	// å would not match aa in Danish. It overrides any of the ignore options.
	Exact Option = nil

	// Loose causes case, diacritics and width to be ignored.
	Loose Option = nil //

	// IgnoreCase enables case-insensitive search.
	IgnoreCase Option = nil

	// IgnoreDiacritics causes diacritics to be ignored ("ö" == "o").
	IgnoreDiacritics Option = nil

	// IgnoreWidth equates fullwidth with halfwidth variants.
	IgnoreWidth Option = nil
)

// New returns a new Matcher for the given language and options.
func New(t language.Tag, opts ...Option) *Matcher {
	panic("TODO: implement")
}

// A Matcher implements language-specific string matching.
type Matcher struct {
	w colltab.Weighter
}

// An IndexOption specifies how the Index methods of Pattern or Matcher should
// match the input.
type IndexOption byte

const (
	// Anchor restricts the search to the start (or end for Backwards) of the
	// text.
	Anchor IndexOption = iota

	// Backwards starts the search from the end of the text.
	Backwards
)

// Design note (TODO remove):
// We use IndexOption, instead of having Index, IndexString, IndexLast,
// IndexLastString, IndexPrefix, IndexPrefixString, ....
// (Note: HasPrefix would have reduced utility compared to those in the  strings
// and bytes packages as the matched prefix in the searched strings may be of
// different lengths, so we need to return an additional index.)
// Advantage:
// - Avoid combinatorial explosion of method calls (now 2 Index variants,
//   instead of 8, or 16 if we have All variants).
// - Compared to an API where these options are set on Matcher or Pattern, it
//   will be clearer when Index*() is invoked with certain options.
// - Small API and still normal Index() call for the by far most common case.
// Disadvantage:
// - Slightly different from analogous packages in the core library (even though
//   there things are not entirely consistent anyway.)
// - Little bit of overhead on each Index call (one branch for the common case.)

// Index reports the start and end position of the first occurrence of pat in b
// or -1, -1 if pat is not present.
func (m *Matcher) Index(b, pat []byte, opts ...IndexOption) (start, end int) {
	// TODO: implement optimized version that does not use a pattern.
	return m.Compile(pat).Index(b, opts...)
}

// IndexString reports the start and end position of the first occurrence of pat
// in s or -1, -1 if pat is not present.
func (m *Matcher) IndexString(s, pat string, opts ...IndexOption) (start, end int) {
	// TODO: implement optimized version that does not use a pattern.
	return m.CompileString(pat).IndexString(s, opts...)
}

// Equal reports whether a and b are equivalent.
func (m *Matcher) Equal(a, b []byte) bool {
	_, end := m.Index(a, b, Anchor)
	return end == len(a)
}

// EqualString reports whether a and b are equivalent.
func (m *Matcher) EqualString(a, b string) bool {
	_, end := m.IndexString(a, b, Anchor)
	return end == len(a)
}

// Compile compiles and returns a pattern that can be used for faster searching.
func (m *Matcher) Compile(b []byte) *Pattern {
	panic("TODO: implement")
}

// CompileString compiles and returns a pattern that can be used for faster
// searching.
func (m *Matcher) CompileString(str string) *Pattern {
	panic("TODO: implement")
}

// A Pattern is a compiled search string. It is safe for concurrent use.
type Pattern struct {
}

// Design note (TODO remove):
// The cost of retrieving collation elements for each rune, which is used for
// search as well, is not trivial. Also, algorithms like Boyer-Moore and
// Sunday require some additional precomputing.

// Index reports the start and end position of the first occurrence of p in b
// or -1, -1 if p is not present.
func (p *Pattern) Index(b []byte, opts ...IndexOption) (start, end int) {
	panic("TODO: implement")
}

// IndexString reports the start and end position of the first occurrence of p
// in s or -1, -1 if p is not present.
func (p *Pattern) IndexString(s string, opts ...IndexOption) (start, end int) {
	panic("TODO: implement")
}

// Supported lists the languages for which search differs from its parent.
var Supported language.Coverage // TODO: implement.

// TODO:
// - Maybe IndexAll methods (probably not necessary).
// - Some way to match patterns in a Reader (a bit tricky).
// - Some fold transformer that folds text to comparable text, based on the
//   search options. This is a common technique, though very different from the
//   collation-based design of this package. It has a somewhat different use
//   case, so probably makes sense to support both. Should probably be in a
//   different package, though, as it uses completely different kind of tables
//   (based on norm, cases, width and range tables.)
