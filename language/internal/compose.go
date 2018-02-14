// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package language

import (
	"sort"
	"strings"
)

type Builder struct {
	Tag Tag

	Private string // the x extension
	Ext     []string
	Variant []string

	Err error
}

func (b *Builder) Make() Tag {
	t := b.Tag

	if len(b.Ext) > 0 || len(b.Variant) > 0 {
		sort.Sort(sortVariants(b.Variant))
		sort.Strings(b.Ext)
		if b.Private != "" {
			b.Ext = append(b.Ext, b.Private)
		}
		n := maxCoreSize + tokenLen(b.Variant...) + tokenLen(b.Ext...)
		buf := make([]byte, n)
		p := t.genCoreBytes(buf)
		t.pVariant = byte(p)
		p += appendTokens(buf[p:], b.Variant...)
		t.pExt = uint16(p)
		p += appendTokens(buf[p:], b.Ext...)
		t.str = string(buf[:p])
	} else if b.Private != "" {
		t.str = b.Private
		t.RemakeString()
	}
	return t
}

func (b *Builder) SetTag(t Tag) {
	b.Tag.LangID = t.LangID
	b.Tag.RegionID = t.RegionID
	b.Tag.ScriptID = t.ScriptID
	// TODO: optimize
	b.Variant = b.Variant[:0]
	if variants := t.Variants(); variants != "" {
		for _, vr := range strings.Split(variants[1:], "-") {
			b.Variant = append(b.Variant, vr)
		}
	}
	b.Ext, b.Private = b.Ext[:0], ""
	for _, e := range t.Extensions() {
		b.AddExt(e)
	}
}

func (b *Builder) AddExt(e string) {
	if e == "" {
	} else if e[0] == 'x' {
		b.Private = e
	} else {
		b.Ext = append(b.Ext, e)
	}
}

func tokenLen(token ...string) (n int) {
	for _, t := range token {
		n += len(t) + 1
	}
	return
}

func appendTokens(b []byte, token ...string) int {
	p := 0
	for _, t := range token {
		b[p] = '-'
		copy(b[p+1:], t)
		p += 1 + len(t)
	}
	return p
}

type sortVariants []string

func (s sortVariants) Len() int {
	return len(s)
}

func (s sortVariants) Swap(i, j int) {
	s[j], s[i] = s[i], s[j]
}

func (s sortVariants) Less(i, j int) bool {
	return variantIndex[s[i]] < variantIndex[s[j]]
}
