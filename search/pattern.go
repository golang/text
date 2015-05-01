// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package search

import (
	"golang.org/x/text/collate/colltab"
)

// TODO: handle variable primary weights?

type source struct {
	m *Matcher
	b []byte
	s string
}

func (s *source) appendNext(a []colltab.Elem, p int) ([]colltab.Elem, int) {
	if s.b != nil {
		return s.m.w.AppendNext(a, s.b[p:])
	}
	return s.m.w.AppendNextString(a, s.s[p:])
}

func (s *source) len() int {
	if s.b != nil {
		return len(s.b)
	}
	return len(s.s)
}

// compile compiles and returns a pattern that can be used for faster searching.
func compile(src source) *Pattern {
	p := &Pattern{m: src.m}

	// Convert the full input to collation elements.
	for m, i, nSrc := 0, 0, src.len(); i < nSrc; i += m {
		p.ce, m = src.appendNext(p.ce, i)
	}

	// Remove empty elements.
	k := 0
	for _, e := range p.ce {
		if !isIgnorable(p.m, e) {
			p.ce[k] = e
			k++
		}
	}
	p.ce = p.ce[:k]

	return p
}

func isIgnorable(m *Matcher, e colltab.Elem) bool {
	if e.Primary() > 0 {
		return false
	}
	if e.Secondary() > 0 {
		if !m.ignoreDiacritics {
			return false
		}
		// Primary value is 0 and ignoreDiacritics is true. In this case we
		// ignore the tertiary element, as it only pertains to the modifier.
		return true
	}
	// TODO: further distinguish once we have the new implementation.
	if !(m.ignoreWidth || m.ignoreCase) && e.Tertiary() > 0 {
		return false
	}
	// TODO: we ignore the Quaternary level for now.
	return true
}

type searcher struct {
	source
	*Pattern
	seg []colltab.Elem // current segment of collation elements
}

// TODO: Use a Boyer-Moore-like algorithm (probably Sunday) for searching.

func (s *searcher) forwardSearch() (start, end int) {
	// Pick a large enough buffer such that we likely do not need to allocate
	// and small enough to not cause too much overhead initializing.
	var buf [8]colltab.Elem

	for m, i, nText := 0, 0, s.len(); i < nText; i += m {
		s.seg, m = s.appendNext(buf[:0], i)
		if end := s.searchOnce(i, m); end != -1 {
			return i, end
		}
	}
	return -1, -1
}

func (s *searcher) anchoredForwardSearch() (start, end int) {
	var buf [8]colltab.Elem
	s.seg, end = s.appendNext(buf[:0], 0)
	if end := s.searchOnce(0, end); end != -1 {
		return 0, end
	}
	return -1, -1
}

// next advances to the next weight in a pattern. f must return one of the
// weights of a collation element. next will advance to the first non-zero
// weight and return this weight and true if it exists, or 0, false otherwise.
func (p *Pattern) next(i *int, f func(colltab.Elem) int) (weight int, ok bool) {
	for *i < len(p.ce) {
		v := f(p.ce[*i])
		*i++
		if v != 0 {
			return v, true
		}
	}
	return 0, false
}

func tertiary(e colltab.Elem) int {
	return int(e.Tertiary())
}

// searchOnce tries to match the pattern s.p at the text position i. s.buf needs
// to be filled with collation elements of the first segment, where n is the
// number of source bytes consumed for this segment. It will return the end
// position of the match or -1.
func (s *searcher) searchOnce(i, n int) (end int) {
	var pLevel [4]int

	// TODO: patch non-normalized strings (see collate.go).

	m := s.Pattern.m
	nSrc := s.len()
	for {
		k := 0
		for ; k < len(s.seg); k++ {
			if v := s.seg[k].Primary(); v > 0 {
				if w, ok := s.next(&pLevel[0], colltab.Elem.Primary); !ok || v != w {
					return -1
				}
			}

			if !m.ignoreDiacritics {
				if v := s.seg[k].Secondary(); v > 0 {
					if w, ok := s.next(&pLevel[1], colltab.Elem.Secondary); !ok || v != w {
						return -1
					}
				}
			} else if s.seg[k].Primary() == 0 {
				// We ignore tertiary values of collation elements of the
				// secondary level.
				continue
			}

			// TODO: distinguish between case and width. This will be easier to
			// implement after we moved to the new collation implementation.
			if !m.ignoreWidth && !m.ignoreCase {
				if v := s.seg[k].Tertiary(); v > 0 {
					if w, ok := s.next(&pLevel[2], tertiary); !ok || int(v) != w {
						return -1
					}
				}
			}
			// TODO: check quaternary weight
		}
		i += n

		// Check for completion.
		switch {
		// If any of these cases match, we are not at the end.
		case pLevel[0] < len(s.ce):
		case !m.ignoreDiacritics && pLevel[1] < len(s.ce):
		case !(m.ignoreWidth || m.ignoreCase) && pLevel[2] < len(s.ce):
		default:
			// At this point, both the segment and pattern has matched fully.
			// However, appendNext does not guarantee a segment is on a grapheme
			// boundary. The proper way to check this is whether the text is
			// followed by a modifier. We inspecting the properties of the
			// collation element as an alternative to using unicode/norm or
			// range tables.
			// TODO: verify correctness of this algorithm and verify space/time
			// trade-offs of the other approaches.
			for ; i < nSrc; i += n {
				// TODO: implement different behavior for WholeGrapheme(false).
				s.seg, n = s.appendNext(s.seg[:0], i)
				if s.seg[0].Primary() != 0 {
					break
				}
				if !m.ignoreDiacritics {
					return -1
				}
			}
			return i
		}

		if i >= nSrc {
			return -1
		}

		// Fill the buffer with the next batch of collation elements.
		s.seg, n = s.appendNext(s.seg[:0], i)
	}
}
