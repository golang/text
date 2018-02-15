// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// Language tag table generator.
// Data read from the web.

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/text/internal/gen"
	"golang.org/x/text/language/internal"
	"golang.org/x/text/unicode/cldr"
)

var (
	test = flag.Bool("test",
		false,
		"test existing tables; can be used to compare web data with package data.")
	outputFile = flag.String("output",
		"tables.go",
		"output file for generated tables")
)

var comment = []string{
	`
matchLang holds pairs of langIDs of base languages that are typically
mutually intelligible. Each pair is associated with a confidence and
whether the intelligibility goes one or both ways.`,
	`
matchScript holds pairs of scriptIDs where readers of one script
can typically also read the other. Each is associated with a confidence.`,
	`
nRegionGroups is the number of region groups.`,
	`
regionInclusion maps region identifiers to sets of regions in regionInclusionBits,
where each set holds all groupings that are directly connected in a region
containment graph.`,
	`
regionInclusionBits is an array of bit vectors where every vector represents
a set of region groupings.  These sets are used to compute the distance
between two regions for the purpose of language matching.`,
	`
regionInclusionNext marks, for each entry in regionInclusionBits, the set of
all groups that are reachable from the groups set in the respective entry.`,
}

type builder struct {
	w    *gen.CodeWriter
	hw   io.Writer // MultiWriter for w and w.Hash
	data *cldr.CLDR
	supp *cldr.SupplementalData

	// lang   index
	region index
	script index
}

func (b *builder) langIndex(s string) uint16 {
	return uint16(language.MustParseBase(s))
}

type index func(s string) int

func (i index) index(s string) int {
	return i(s)
}

func newBuilder(w *gen.CodeWriter) *builder {
	r := gen.OpenCLDRCoreZip()
	defer r.Close()
	d := &cldr.Decoder{}
	data, err := d.DecodeZip(r)
	if err != nil {
		log.Fatal(err)
	}
	b := builder{
		w:    w,
		hw:   io.MultiWriter(w, w.Hash),
		data: data,
		supp: data.Supplemental(),

		script: func(s string) int {
			return int(language.MustParseScript(s))
		},

		region: func(s string) int {
			return int(language.MustParseRegion(s))
		},
	}
	return &b
}

var commentIndex = make(map[string]string)

func init() {
	for _, s := range comment {
		key := strings.TrimSpace(strings.SplitN(s, " ", 2)[0])
		commentIndex[key] = s
	}
}

func (b *builder) comment(name string) {
	if s := commentIndex[name]; len(s) > 0 {
		b.w.WriteComment(s)
	} else {
		fmt.Fprintln(b.w)
	}
}

func (b *builder) pf(f string, x ...interface{}) {
	fmt.Fprintf(b.hw, f, x...)
	fmt.Fprint(b.hw, "\n")
}

func (b *builder) p(x ...interface{}) {
	fmt.Fprintln(b.hw, x...)
}

func (b *builder) addSize(s int) {
	b.w.Size += s
	b.pf("// Size: %d bytes", s)
}

func (b *builder) writeConst(name string, x interface{}) {
	b.comment(name)
	b.w.WriteConst(name, x)
}

// writeConsts computes f(v) for all v in values and writes the results
// as constants named _v to a single constant block.
func (b *builder) writeConsts(f func(string) int, values ...string) {
	b.pf("const (")
	for _, v := range values {
		b.pf("\t_%s = %v", v, f(v))
	}
	b.pf(")")
}

// writeType writes the type of the given value, which must be a struct.
func (b *builder) writeType(value interface{}) {
	b.comment(reflect.TypeOf(value).Name())
	b.w.WriteType(value)
}

func (b *builder) writeSlice(name string, ss interface{}) {
	b.writeSliceAddSize(name, 0, ss)
}

func (b *builder) writeSliceAddSize(name string, extraSize int, ss interface{}) {
	b.comment(name)
	b.w.Size += extraSize
	v := reflect.ValueOf(ss)
	t := v.Type().Elem()
	b.pf("// Size: %d bytes, %d elements", v.Len()*int(t.Size())+extraSize, v.Len())

	fmt.Fprintf(b.w, "var %s = ", name)
	b.w.WriteArray(ss)
	b.p()
}

// TODO: region inclusion data will probably not be use used in future matchers.

var langConsts = []string{
	"af", "am", "ar", "az", "bg", "bn", "ca", "cs", "da", "de", "el", "en", "es",
	"et", "fa", "fi", "fil", "fr", "gu", "he", "hi", "hr", "hu", "hy", "id", "is",
	"it", "ja", "ka", "kk", "km", "kn", "ko", "ky", "lo", "lt", "lv", "mk", "ml",
	"mn", "mo", "mr", "ms", "mul", "my", "nb", "ne", "nl", "no", "pa", "pl", "pt",
	"ro", "ru", "sh", "si", "sk", "sl", "sq", "sr", "sv", "sw", "ta", "te", "th",
	"tl", "tn", "tr", "uk", "ur", "uz", "vi", "zh", "zu",

	// constants for grandfathered tags (if not already defined)
	"jbo", "ami", "bnn", "hak", "tlh", "lb", "nv", "pwn", "tao", "tay", "tsu",
	"nn", "sfb", "vgt", "sgg", "cmn", "nan", "hsn",
}

var scriptConsts = []string{
	"Latn", "Hani", "Hans", "Hant", "Qaaa", "Qaai", "Qabx", "Zinh", "Zyyy",
	"Zzzz",
}

var regionConsts = []string{
	"001", "419", "BR", "CA", "ES", "GB", "MD", "PT", "UK", "US",
	"ZZ", "XA", "XC", "XK", // Unofficial tag for Kosovo.
}

func (b *builder) writeConstants() {
	b.writeConsts(func(s string) int { return int(b.langIndex(s)) }, langConsts...)
	b.writeConsts(b.region, regionConsts...)
	b.writeConsts(b.script, scriptConsts...)
}

type mutualIntelligibility struct {
	want, have uint16
	distance   uint8
	oneway     bool
}

type scriptIntelligibility struct {
	wantLang, haveLang     uint16
	wantScript, haveScript uint8
	distance               uint8
	// Always oneway
}

type regionIntelligibility struct {
	lang     uint16 // compact language id
	script   uint8  // 0 means any
	group    uint8  // 0 means any; if bit 7 is set it means inverse
	distance uint8
	// Always twoway.
}

// writeMatchData writes tables with languages and scripts for which there is
// mutual intelligibility. The data is based on CLDR's languageMatching data.
// Note that we use a different algorithm than the one defined by CLDR and that
// we slightly modify the data. For example, we convert scores to confidence levels.
// We also drop all region-related data as we use a different algorithm to
// determine region equivalence.
func (b *builder) writeMatchData() {
	lm := b.supp.LanguageMatching.LanguageMatches
	cldr.MakeSlice(&lm).SelectAnyOf("type", "written_new")

	regionHierarchy := map[string][]string{}
	for _, g := range b.supp.TerritoryContainment.Group {
		regions := strings.Split(g.Contains, " ")
		regionHierarchy[g.Type] = append(regionHierarchy[g.Type], regions...)
	}
	regionToGroups := make([]uint8, language.NumRegions)

	idToIndex := map[string]uint8{}
	for i, mv := range lm[0].MatchVariable {
		if i > 6 {
			log.Fatalf("Too many groups: %d", i)
		}
		idToIndex[mv.Id] = uint8(i + 1)
		// TODO: also handle '-'
		for _, r := range strings.Split(mv.Value, "+") {
			todo := []string{r}
			for k := 0; k < len(todo); k++ {
				r := todo[k]
				regionToGroups[b.region.index(r)] |= 1 << uint8(i)
				todo = append(todo, regionHierarchy[r]...)
			}
		}
	}
	b.writeSlice("regionToGroups", regionToGroups)

	// maps language id to in- and out-of-group region.
	paradigmLocales := [][3]uint16{}
	locales := strings.Split(lm[0].ParadigmLocales[0].Locales, " ")
	for i := 0; i < len(locales); i += 2 {
		x := [3]uint16{}
		for j := 0; j < 2; j++ {
			pc := strings.SplitN(locales[i+j], "-", 2)
			x[0] = b.langIndex(pc[0])
			if len(pc) == 2 {
				x[1+j] = uint16(b.region.index(pc[1]))
			}
		}
		paradigmLocales = append(paradigmLocales, x)
	}
	b.writeSlice("paradigmLocales", paradigmLocales)

	b.writeType(mutualIntelligibility{})
	b.writeType(scriptIntelligibility{})
	b.writeType(regionIntelligibility{})

	matchLang := []mutualIntelligibility{}
	matchScript := []scriptIntelligibility{}
	matchRegion := []regionIntelligibility{}
	// Convert the languageMatch entries in lists keyed by desired language.
	for _, m := range lm[0].LanguageMatch {
		// Different versions of CLDR use different separators.
		desired := strings.Replace(m.Desired, "-", "_", -1)
		supported := strings.Replace(m.Supported, "-", "_", -1)
		d := strings.Split(desired, "_")
		s := strings.Split(supported, "_")
		if len(d) != len(s) {
			log.Fatalf("not supported: desired=%q; supported=%q", desired, supported)
			continue
		}
		distance, _ := strconv.ParseInt(m.Distance, 10, 8)
		switch len(d) {
		case 2:
			if desired == supported && desired == "*_*" {
				continue
			}
			// language-script pair.
			matchScript = append(matchScript, scriptIntelligibility{
				wantLang:   uint16(b.langIndex(d[0])),
				haveLang:   uint16(b.langIndex(s[0])),
				wantScript: uint8(b.script.index(d[1])),
				haveScript: uint8(b.script.index(s[1])),
				distance:   uint8(distance),
			})
			if m.Oneway != "true" {
				matchScript = append(matchScript, scriptIntelligibility{
					wantLang:   uint16(b.langIndex(s[0])),
					haveLang:   uint16(b.langIndex(d[0])),
					wantScript: uint8(b.script.index(s[1])),
					haveScript: uint8(b.script.index(d[1])),
					distance:   uint8(distance),
				})
			}
		case 1:
			if desired == supported && desired == "*" {
				continue
			}
			if distance == 1 {
				// nb == no is already handled by macro mapping. Check there
				// really is only this case.
				if d[0] != "no" || s[0] != "nb" {
					log.Fatalf("unhandled equivalence %s == %s", s[0], d[0])
				}
				continue
			}
			// TODO: consider dropping oneway field and just doubling the entry.
			matchLang = append(matchLang, mutualIntelligibility{
				want:     uint16(b.langIndex(d[0])),
				have:     uint16(b.langIndex(s[0])),
				distance: uint8(distance),
				oneway:   m.Oneway == "true",
			})
		case 3:
			if desired == supported && desired == "*_*_*" {
				continue
			}
			if desired != supported {
				// This is now supported by CLDR, but only one case, which
				// should already be covered by paradigm locales. For instance,
				// test case "und, en, en-GU, en-IN, en-GB ; en-ZA ; en-GB" in
				// testdata/CLDRLocaleMatcherTest.txt tests this.
				if supported != "en_*_GB" {
					log.Fatalf("not supported: desired=%q; supported=%q", desired, supported)
				}
				continue
			}
			ri := regionIntelligibility{
				lang:     b.langIndex(d[0]),
				distance: uint8(distance),
			}
			if d[1] != "*" {
				ri.script = uint8(b.script.index(d[1]))
			}
			switch {
			case d[2] == "*":
				ri.group = 0x80 // not contained in anything
			case strings.HasPrefix(d[2], "$!"):
				ri.group = 0x80
				d[2] = "$" + d[2][len("$!"):]
				fallthrough
			case strings.HasPrefix(d[2], "$"):
				ri.group |= idToIndex[d[2]]
			}
			matchRegion = append(matchRegion, ri)
		default:
			log.Fatalf("not supported: desired=%q; supported=%q", desired, supported)
		}
	}
	sort.SliceStable(matchLang, func(i, j int) bool {
		return matchLang[i].distance < matchLang[j].distance
	})
	b.writeSlice("matchLang", matchLang)

	sort.SliceStable(matchScript, func(i, j int) bool {
		return matchScript[i].distance < matchScript[j].distance
	})
	b.writeSlice("matchScript", matchScript)

	sort.SliceStable(matchRegion, func(i, j int) bool {
		return matchRegion[i].distance < matchRegion[j].distance
	})
	b.writeSlice("matchRegion", matchRegion)
}

func main() {
	gen.Init()

	w := gen.NewCodeWriter()
	defer w.WriteGoFile("tables.go", "language")

	b := newBuilder(w)
	gen.WriteCLDRVersion(w)

	b.writeConstants()
	b.writeMatchData()
}
