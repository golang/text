// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package locale

import "errors"

type scriptRegionFlags uint8

const (
	isList = 1 << iota
	scriptInFrom
	regionInFrom
)

func (loc *ID) setUndefinedLang(id langID) {
	if loc.lang == 0 {
		loc.lang = id
	}
}

func (loc *ID) setUndefinedScript(id scriptID) {
	if loc.script == 0 {
		loc.script = id
	}
}

func (loc *ID) setUndefinedRegion(id regionID) {
	if loc.region == 0 {
		loc.region = id
	}
}

// MissingLikelyTagsData indicates no information was available
// to compute likely values of missing tags.
var MissingLikelyTagsData = errors.New("missing likely tags data")

// addLikelySubtags sets subtags to their most likely value, given the locale.
// In most cases this means setting fields for unknown values, but in some
// cases it may alter a value.  It returns a MissingLikelyTagsData error
// if the given locale cannot be expaned.
func (loc ID) addLikelySubtags() (ID, error) {
	// Hard-coded exception.  This is currently the only exception to the rule
	// that any defined value before expanding likely subtags remains the same.
	// maketables verifies this is indeed the only case.
	// We include this in addLikelySubtags instead of addTags to guarantee that
	// Minimize does not alter any of the tags.
	if loc.script == scrHani {
		loc.script = scrHans
	}
	id, err := addTags(loc)
	if err != nil {
		return loc, err
	} else if id.equalTags(loc) {
		return loc, nil
	}
	id.remakeString()
	return id, nil
}

func addTags(loc ID) (ID, error) {
	// We leave private use identifiers alone.
	if loc.private() {
		return loc, nil
	}
	if loc.script != 0 && loc.region != 0 {
		if loc.lang != 0 {
			// already fully specified
			return loc, nil
		}
		// Search matches for und-script-region.
		list := likelyRegion[loc.region : loc.region+1]
		if x := list[0]; x.flags&isList != 0 {
			list = likelyRegionList[x.lang : x.lang+uint16(x.script)]
		}
		for _, x := range list {
			// Deviating from the spec. See match_test.go for details.
			if scriptID(x.script) == loc.script {
				loc.setUndefinedLang(langID(x.lang))
				return loc, nil
			}
		}
	}
	if loc.lang != 0 {
		// Search matches for lang-script and lang-region, where lang != und.
		if loc.lang < langNoIndexOffset {
			x := likelyLang[loc.lang]
			if x.flags&isList != 0 {
				list := likelyLangList[x.region : x.region+uint16(x.script)]
				if loc.script != 0 {
					for _, x := range list {
						if scriptID(x.script) == loc.script && x.flags&scriptInFrom != 0 {
							loc.setUndefinedRegion(regionID(x.region))
							return loc, nil
						}
					}
				} else if loc.region != 0 {
					for _, x := range list {
						if regionID(x.region) == loc.region && x.flags&regionInFrom != 0 {
							loc.setUndefinedScript(scriptID(x.script))
							return loc, nil
						}
					}
				}
			}
		}
	} else {
		// Search matches for und-script.
		if loc.script != 0 {
			x := likelyScript[loc.script]
			if x.region != 0 {
				loc.setUndefinedRegion(regionID(x.region))
				loc.setUndefinedLang(langID(x.lang))
				return loc, nil
			}
		}
		// Search matches for und-region.
		if loc.region != 0 {
			x := likelyRegion[loc.region]
			if x.flags&isList != 0 {
				x = likelyRegionList[x.lang]
			}
			if x.script != 0 && x.flags != scriptInFrom {
				loc.setUndefinedLang(langID(x.lang))
				loc.setUndefinedScript(scriptID(x.script))
				return loc, nil
			}
		}
	}
	// Search matches for lang.
	if loc.lang < langNoIndexOffset {
		x := likelyLang[loc.lang]
		if x.flags&isList != 0 {
			x = likelyLangList[x.region]
		}
		if x.region != 0 {
			loc.setUndefinedScript(scriptID(x.script))
			loc.setUndefinedRegion(regionID(x.region))
			if loc.lang == 0 {
				loc.lang = lang_en // default language
			}
			return loc, nil
		}
	}
	return loc, MissingLikelyTagsData
}

func (loc *ID) setTagsFrom(id ID) {
	loc.lang = id.lang
	loc.script = id.script
	loc.region = id.region
}

// minimize removes the region or script subtags from loc such that
// loc.addLikelySubtags() == loc.minimize().addLikelySubtags().
func (loc ID) minimize() (ID, error) {
	loc, err := minimizeTags(loc)
	if err != nil {
		return loc, err
	}
	loc.remakeString()
	return loc, nil
}

// minimizeTags mimics the behavior of the ICU 51 C implementation.
func minimizeTags(loc ID) (ID, error) {
	if loc.equalTags(und) {
		return loc, nil
	}
	max, err := addTags(loc)
	if err != nil {
		return loc, err
	}
	for _, id := range [...]ID{
		{lang: loc.lang},
		{lang: loc.lang, region: loc.region},
		{lang: loc.lang, script: loc.script},
	} {
		if x, err := addTags(id); err == nil && max.equalTags(x) {
			loc.setTagsFrom(id)
			break
		}
	}
	return loc, nil
}

// regionDistance computes the distance between two regions based
// on the distance in the graph of region containments as defined in CLDR.
// It iterates over increasingly inclusive sets of groups, represented as
// bit vectors, until the source bit vector has bits in common with the
// destination vector.
func regionDistance(a, b regionID) int {
	if a == b {
		return 0
	}
	p, q := regionInclusion[a], regionInclusion[b]
	if p < nRegionGroups {
		p, q = q, p
	}
	set := regionInclusionBits
	if q < nRegionGroups && set[p]&(1<<q) != 0 {
		return 1
	}
	d := 2
	for goal := set[q]; set[p]&goal == 0; p = regionInclusionNext[p] {
		d++
	}
	return d
}
