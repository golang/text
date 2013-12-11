// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package norm contains types and functions for normalizing Unicode strings.
package norm

import "unicode/utf8"

// A Form denotes a canonical representation of Unicode code points.
// The Unicode-defined normalization and equivalence forms are:
//
//   NFC   Unicode Normalization Form C
//   NFD   Unicode Normalization Form D
//   NFKC  Unicode Normalization Form KC
//   NFKD  Unicode Normalization Form KD
//
// For a Form f, this documentation uses the notation f(x) to mean
// the bytes or string x converted to the given form.
// A position n in x is called a boundary if conversion to the form can
// proceed independently on both sides:
//   f(x) == append(f(x[0:n]), f(x[n:])...)
//
// References: http://unicode.org/reports/tr15/ and
// http://unicode.org/notes/tn5/.
type Form int

const (
	NFC Form = iota
	NFD
	NFKC
	NFKD
)

// Bytes returns f(b). May return b if f(b) = b.
func (f Form) Bytes(b []byte) []byte {
	src := inputBytes(b)
	ft := formTable[f]
	n, ok := ft.quickSpan(src, 0, len(b), true)
	if ok {
		return b
	}
	out := make([]byte, n, len(b))
	copy(out, b[0:n])
	rb := reorderBuffer{f: *ft, src: src, nsrc: len(b)}
	return doAppend(&rb, out, n)
}

// String returns f(s).
func (f Form) String(s string) string {
	src := inputString(s)
	ft := formTable[f]
	n, ok := ft.quickSpan(src, 0, len(s), true)
	if ok {
		return s
	}
	out := make([]byte, n, len(s))
	copy(out, s[0:n])
	rb := reorderBuffer{f: *ft, src: src, nsrc: len(s)}
	return string(doAppend(&rb, out, n))
}

// IsNormal returns true if b == f(b).
func (f Form) IsNormal(b []byte) bool {
	src := inputBytes(b)
	ft := formTable[f]
	bp, ok := ft.quickSpan(src, 0, len(b), true)
	if ok {
		return true
	}
	rb := reorderBuffer{f: *ft, src: src, nsrc: len(b)}
	rb.setFlusher(nil, cmpNormalBytes)
	for bp < len(b) {
		rb.out = b[bp:]
		if bp = decomposeSegment(&rb, bp, true); bp < 0 {
			return false
		}
		bp, _ = rb.f.quickSpan(rb.src, bp, len(b), true)
	}
	return true
}

func cmpNormalBytes(rb *reorderBuffer) bool {
	b := rb.out
	for i := 0; i < rb.nrune; i++ {
		info := rb.rune[i]
		if int(info.size) > len(b) {
			return false
		}
		p := info.pos
		pe := p + info.size
		for ; p < pe; p++ {
			if b[0] != rb.byte[p] {
				return false
			}
			b = b[1:]
		}
	}
	return true
}

// IsNormalString returns true if s == f(s).
func (f Form) IsNormalString(s string) bool {
	src := inputString(s)
	ft := formTable[f]
	bp, ok := ft.quickSpan(src, 0, len(s), true)
	if ok {
		return true
	}
	rb := reorderBuffer{f: *ft, src: src, nsrc: len(s)}
	rb.setFlusher(nil, func(rb *reorderBuffer) bool {
		for i := 0; i < rb.nrune; i++ {
			info := rb.rune[i]
			if bp+int(info.size) > len(s) {
				return false
			}
			p := info.pos
			pe := p + info.size
			for ; p < pe; p++ {
				if s[bp] != rb.byte[p] {
					return false
				}
				bp++
			}
		}
		return true
	})
	for bp < len(s) {
		if e := decomposeSegment(&rb, bp, true); e < 0 {
			return false
		}
		bp, _ = rb.f.quickSpan(rb.src, bp, len(s), true)
	}
	return true
}

// patchTail fixes a case where a rune may be incorrectly normalized
// if it is followed by illegal continuation bytes. It returns the
// patched buffer and whether there were trailing continuation bytes.
func patchTail(rb *reorderBuffer) bool {
	info, p := lastRuneStart(&rb.f, rb.out)
	if p == -1 || info.size == 0 {
		return false
	}
	end := p + int(info.size)
	extra := len(rb.out) - end
	if extra > 0 {
		// Potentially allocating memory. However, this only
		// happens with ill-formed UTF-8.
		x := make([]byte, 0)
		x = append(x, rb.out[len(rb.out)-extra:]...)
		rb.out = rb.out[:end]
		decomposeToLastBoundary(rb)
		rb.doFlush()
		rb.out = append(rb.out, x...)
		return true
	}
	return false
}

func appendQuick(rb *reorderBuffer, i int) int {
	if rb.nsrc == i {
		return i
	}
	end, _ := rb.f.quickSpan(rb.src, i, rb.nsrc, true)
	rb.out = rb.src.appendSlice(rb.out, i, end)
	return end
}

// Append returns f(append(out, b...)).
// The buffer out must be nil, empty, or equal to f(out).
func (f Form) Append(out []byte, src ...byte) []byte {
	return f.doAppend(out, inputBytes(src), len(src))
}

func (f Form) doAppend(out []byte, src input, n int) []byte {
	if n == 0 {
		return out
	}
	ft := formTable[f]
	// Attempt to do a quickSpan first so we can avoid initializing the reorderBuffer.
	if len(out) == 0 {
		p, _ := ft.quickSpan(src, 0, n, true)
		out = src.appendSlice(out, 0, p)
		if p == n {
			return out
		}
		rb := reorderBuffer{f: *ft, src: src, nsrc: n, out: out, flushF: appendFlush}
		return doAppendInner(&rb, p)
	}
	rb := reorderBuffer{f: *ft, src: src, nsrc: n}
	return doAppend(&rb, out, 0)
}

func doAppend(rb *reorderBuffer, out []byte, p int) []byte {
	rb.setFlusher(out, appendFlush)
	src, n := rb.src, rb.nsrc
	doMerge := len(out) > 0
	if q := src.skipNonStarter(p); q > p {
		// Move leading non-starters to destination.
		rb.out = src.appendSlice(rb.out, p, q)
		if patchTail(rb) {
			doMerge = false // no need to merge, ends with illegal UTF-8
		} else {
			decomposeToLastBoundary(rb) // force decomposition
		}
		p = q
	}
	fd := &rb.f
	if doMerge {
		var info Properties
		if p < n {
			info = fd.info(src, p)
			if p == 0 && !info.BoundaryBefore() {
				decomposeToLastBoundary(rb)
			}
		}
		if info.size == 0 || info.BoundaryBefore() {
			rb.doFlush()
			if info.size == 0 {
				// Append incomplete UTF-8 encoding.
				return src.appendSlice(rb.out, p, n)
			}
		}
	}
	if rb.nrune == 0 {
		p = appendQuick(rb, p)
	}
	return doAppendInner(rb, p)
}

func doAppendInner(rb *reorderBuffer, p int) []byte {
	for n := rb.nsrc; p < n; {
		// do we need to check for short src?
		p = decomposeSegment(rb, p, true)
		p = appendQuick(rb, p)
	}
	return rb.out
}

// AppendString returns f(append(out, []byte(s))).
// The buffer out must be nil, empty, or equal to f(out).
func (f Form) AppendString(out []byte, src string) []byte {
	return f.doAppend(out, inputString(src), len(src))
}

// QuickSpan returns a boundary n such that b[0:n] == f(b[0:n]).
// It is not guaranteed to return the largest such n.
func (f Form) QuickSpan(b []byte) int {
	n, _ := formTable[f].quickSpan(inputBytes(b), 0, len(b), true)
	return n
}

// quickSpan returns a boundary n such that src[0:n] == f(src[0:n]) and
// whether any non-normalized parts were found. If atEOF is false, n will
// not point past the last segment if this segment might be become
// non-normalized by appending other runes.
func (f *formInfo) quickSpan(src input, i, end int, atEOF bool) (n int, ok bool) {
	var lastCC uint8
	var nc int
	lastSegStart := i
	for n = end; i < n; {
		if j := src.skipASCII(i, n); i != j {
			i = j
			lastSegStart = i - 1
			lastCC = 0
			nc = 0
			continue
		}
		info := f.info(src, i)
		if info.size == 0 {
			if atEOF {
				// include incomplete runes
				return n, true
			}
			return lastSegStart, true
		}
		cc := info.ccc
		if f.composing {
			if !info.isYesC() {
				break
			}
		} else {
			if !info.isYesD() {
				break
			}
		}
		if cc == 0 {
			lastSegStart = i
			nc = 0
		} else if nc >= maxCombiningChars {
			lastSegStart = i
			nc = 1
		} else {
			if lastCC > cc {
				return lastSegStart, false
			}
			nc++
		}
		lastCC = cc
		i += int(info.size)
	}
	if i == n {
		if !atEOF {
			n = lastSegStart
		}
		return n, true
	}
	if f.composing {
		return lastSegStart, false
	}
	return i, false
}

// QuickSpanString returns a boundary n such that b[0:n] == f(s[0:n]).
// It is not guaranteed to return the largest such n.
func (f Form) QuickSpanString(s string) int {
	n, _ := formTable[f].quickSpan(inputString(s), 0, len(s), true)
	return n
}

// FirstBoundary returns the position i of the first boundary in b
// or -1 if b contains no boundary.
func (f Form) FirstBoundary(b []byte) int {
	return f.firstBoundary(inputBytes(b), len(b))
}

func (f Form) firstBoundary(src input, nsrc int) int {
	i := src.skipNonStarter(0)
	if i >= nsrc {
		return -1
	}
	fd := formTable[f]
	info := fd.info(src, i)
	for n := 0; info.size != 0 && !info.BoundaryBefore(); {
		i += int(info.size)
		if n++; n >= maxCombiningChars {
			return i
		}
		if i >= nsrc {
			if !info.BoundaryAfter() {
				return -1
			}
			return nsrc
		}
		info = fd.info(src, i)
	}
	if info.size == 0 {
		return -1
	}
	return i
}

// FirstBoundaryInString returns the position i of the first boundary in s
// or -1 if s contains no boundary.
func (f Form) FirstBoundaryInString(s string) int {
	return f.firstBoundary(inputString(s), len(s))
}

// LastBoundary returns the position i of the last boundary in b
// or -1 if b contains no boundary.
func (f Form) LastBoundary(b []byte) int {
	return lastBoundary(formTable[f], b)
}

func lastBoundary(fd *formInfo, b []byte) int {
	i := len(b)
	info, p := lastRuneStart(fd, b)
	if p == -1 {
		return -1
	}
	if info.size == 0 { // ends with incomplete rune
		if p == 0 { // starts with incomplete rune
			return -1
		}
		i = p
		info, p = lastRuneStart(fd, b[:i])
		if p == -1 { // incomplete UTF-8 encoding or non-starter bytes without a starter
			return i
		}
	}
	if p+int(info.size) != i { // trailing non-starter bytes: illegal UTF-8
		return i
	}
	if info.BoundaryAfter() {
		return i
	}
	i = p
	for n := 0; i >= 0 && !info.BoundaryBefore(); {
		info, p = lastRuneStart(fd, b[:i])
		if n++; n >= maxCombiningChars {
			return len(b)
		}
		if p+int(info.size) != i {
			if p == -1 { // no boundary found
				return -1
			}
			return i // boundary after an illegal UTF-8 encoding
		}
		i = p
	}
	return i
}

// decomposeSegment scans the first segment in src into rb.
// It returns the number of bytes consumed from src or iShortDst
// or iShortSrc.
// TODO(mpvl): consider inserting U+034f (Combining Grapheme Joiner)
// when we detect a sequence of 30+ non-starter chars.
func decomposeSegment(rb *reorderBuffer, sp int, atEOF bool) int {
	// Force one character to be consumed.
	info := rb.f.info(rb.src, sp)
	if info.size == 0 {
		return 0
	}
	for {
		if err := rb.insert(rb.src, sp, info); err != 0 {
			if err != iOverflow {
				return int(err)
			}
			break
		}
		sp += int(info.size)
		if sp >= rb.nsrc {
			if !atEOF && !info.BoundaryAfter() {
				return int(iShortSrc)
			}
			break
		}
		info = rb.f.info(rb.src, sp)
		if info.size == 0 {
			if !atEOF {
				return int(iShortSrc)
			}
			break
		}
		if info.BoundaryBefore() {
			break
		}
	}
	if !rb.doFlush() {
		return int(iShortDst)
	}
	return sp
}

// lastRuneStart returns the runeInfo and position of the last
// rune in buf or the zero runeInfo and -1 if no rune was found.
func lastRuneStart(fd *formInfo, buf []byte) (Properties, int) {
	p := len(buf) - 1
	for ; p >= 0 && !utf8.RuneStart(buf[p]); p-- {
	}
	if p < 0 {
		return Properties{}, -1
	}
	return fd.info(inputBytes(buf), p), p
}

// decomposeToLastBoundary finds an open segment at the end of the buffer
// and scans it into rb. Returns the buffer minus the last segment.
func decomposeToLastBoundary(rb *reorderBuffer) {
	fd := &rb.f
	info, i := lastRuneStart(fd, rb.out)
	if int(info.size) != len(rb.out)-i {
		// illegal trailing continuation bytes
		return
	}
	if info.BoundaryAfter() {
		return
	}
	var add [maxBackRunes]Properties // stores runeInfo in reverse order
	add[0] = info
	padd := 1
	p := len(rb.out) - int(info.size)
	for ; p >= 0 && padd < maxBackRunes && !info.BoundaryBefore(); p -= int(info.size) {
		info, i = lastRuneStart(fd, rb.out[:p])
		if int(info.size) != p-i {
			break
		}
		add[padd] = info
		padd++
	}
	// Copy bytes for insertion as we may need to overwrite rb.out.
	var buf [maxBackRunes * utf8.UTFMax]byte
	cp := buf[:copy(buf[:], rb.out[p:])]
	rb.out = rb.out[:p]
	for padd--; padd >= 0; padd-- {
		info = add[padd]
		if rb.insert(inputBytes(cp), 0, info) == iOverflow {
			rb.doFlush()
		}
		cp = cp[info.size:]
	}
}
