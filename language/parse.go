// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package language

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// isAlpha returns true if the byte is not a digit.
// b must be an ASCII letter or digit.
func isAlpha(b byte) bool {
	return b > '9'
}

// isAlphaNum returns true if the string contains only ASCII letters or digits.
func isAlphaNum(s []byte) bool {
	for _, c := range s {
		if !('a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9') {
			return false
		}
	}
	return true
}

var (
	errUnknown  = errors.New("language: unknown language, script, region or currency")
	errEmpty    = errors.New("language: empty language tag")
	errInvalid  = errors.New("language: invalid")
	errTrailSep = errors.New("language: trailing separator")
)

// scanner is used to scan BCP 47 tokens, which are separated by _ or -.
type scanner struct {
	b     []byte
	bytes [64]byte // small buffer to cover most common cases
	token []byte
	start int // start position of the current token
	end   int // end position of the current token
	next  int // next point for scan
	err   error
	done  bool
}

func makeScannerString(s string) scanner {
	scan := scanner{}
	if len(s) <= len(scan.bytes) {
		scan.b = scan.bytes[:copy(scan.bytes[:], s)]
	} else {
		scan.b = []byte(s)
	}
	scan.init()
	return scan
}

func (s *scanner) init() {
	for i, c := range s.b {
		if c == '_' {
			s.b[i] = '-'
		}
	}
	s.scan()
}

// restToLower converts the string between start and end to lower case.
func (s *scanner) toLower(start, end int) {
	for i := start; i < end; i++ {
		c := s.b[i]
		if 'A' <= c && c <= 'Z' {
			s.b[i] += 'a' - 'A'
		}
	}
}

func (s *scanner) setError(e error) {
	if s.err == nil {
		s.err = e
	}
}

func (s *scanner) setErrorf(f string, x ...interface{}) {
	s.setError(fmt.Errorf(f, x...))
}

// replace replaces the current token with repl.
func (s *scanner) replace(repl string) {
	if end := s.start + len(repl); end != s.end {
		diff := end - s.end
		if end < cap(s.b) {
			b := make([]byte, len(s.b)+diff)
			copy(b, s.b[:s.start])
			copy(b[end:], s.b[s.end:])
			s.b = b
		} else {
			s.b = append(s.b[end:], s.b[s.end:]...)
		}
		s.next += diff
		s.end = end
	}
	copy(s.b[s.start:], repl)
}

// gobble removes the current token from the input.
// Caller must call scan after calling gobble.
func (s *scanner) gobble() {
	if s.start == 0 {
		s.b = s.b[:+copy(s.b, s.b[s.next:])]
		s.end = 0
	} else {
		s.b = s.b[:s.start-1+copy(s.b[s.start-1:], s.b[s.end:])]
		s.end = s.start - 1
	}
	s.next = s.start
}

// scan parses the next token of a BCP 47 string.  Tokens that are larger
// than 8 characters or include non-alphanumeric characters result in an error
// and are gobbled and removed from the output.
// It returns the end position of the last token consumed.
func (s *scanner) scan() (end int) {
	end = s.end
	s.token = nil
	for s.start = s.next; s.next < len(s.b); {
		i := bytes.IndexByte(s.b[s.next:], '-')
		if i == -1 {
			s.end = len(s.b)
			s.next = len(s.b)
			i = s.end - s.start
		} else {
			s.end = s.next + i
			s.next = s.end + 1
		}
		token := s.b[s.start:s.end]
		if i < 1 || i > 8 || !isAlphaNum(token) {
			s.setErrorf("language: invalid token %q", token)
			s.gobble()
			continue
		}
		s.token = token
		return end
	}
	if n := len(s.b); n > 0 && s.b[n-1] == '-' {
		s.setError(errTrailSep)
		s.b = s.b[:len(s.b)-1]
	}
	s.done = true
	return end
}

// acceptMinSize parses multiple tokens of the given size or greater.
// It returns the end position of the last token consumed.
func (s *scanner) acceptMinSize(min int) (end int) {
	end = s.end
	s.scan()
	for ; len(s.token) >= min; s.scan() {
		end = s.end
	}
	return end
}

// Parse parses the given BCP 47 string and returns a valid Tag.
// If parsing failed it returns an error and any part of the tag
// that could be parsed.
// If parsing succeeded but an unknown option was found, it
// returns the valid Locale and an error.
// It accepts tags in the BCP 47 format and extensions to this standard
// defined in
// http://www.unicode.org/reports/tr35/#Unicode_Language_and_Locale_Identifiers.
func Parse(s string) (t Tag, err error) {
	// TODO: consider supporting old-style locale key-value pairs.
	if s == "" {
		return und, errEmpty
	}
	t = und
	if lang, ok := tagAlias[s]; ok {
		t.lang = langID(lang)
		return
	}
	scan := makeScannerString(s)
	if len(scan.token) >= 4 {
		if !strings.EqualFold(s, "root") {
			return und, errInvalid
		}
		return und, nil
	}
	return parse(&scan, s)
}

func parse(scan *scanner, s string) (t Tag, err error) {
	t = und
	var end int
	if n := len(scan.token); n <= 1 {
		scan.toLower(0, len(scan.b))
		end = parsePrivate(scan)
	} else if n >= 4 {
		return und, errInvalid
	} else { // the usual case
		t, end = parseTag(scan)
		if n := len(scan.token); n == 1 {
			t.pExt = uint16(end)
			end = parseExtensions(scan)
		}
	}
	if end < len(scan.b) {
		scan.setErrorf("language: invalid parts %q", scan.b[end:])
		scan.b = scan.b[:end]
	}
	if len(scan.b) < len(s) {
		s = s[:len(scan.b)]
	}
	if len(s) > 0 && cmp(s, scan.b) == 0 {
		t.str = &s
	} else if t.pVariant < uint8(end) {
		s = string(scan.b)
		t.str = &s
	}
	return t, scan.err
}

// parseTag parses language, script, region and variants.
// It returns a Tag and the end position in the input that was parsed.
func parseTag(scan *scanner) (t Tag, end int) {
	var e error
	// TODO: set an error if an unknown lang, script or region is encountered.
	t.lang, e = getLangID(scan.token)
	scan.setError(e)
	scan.replace(t.lang.String())
	langStart := scan.start
	end = scan.scan()
	for len(scan.token) == 3 && isAlpha(scan.token[0]) {
		// From http://tools.ietf.org/html/bcp47, <lang>-<extlang> tags are equivalent
		// to a tag of the form <extlang>.
		if lang, e := getLangID(scan.token); lang != 0 {
			t.lang = lang
			scan.setError(e)
			copy(scan.b[langStart:], lang.String())
			scan.b[langStart+3] = '-'
			scan.start = langStart + 4
		} else {
			scan.setError(e)
		}
		scan.gobble()
		end = scan.scan()
	}
	if len(scan.token) == 4 && isAlpha(scan.token[0]) {
		t.script, e = getScriptID(script, scan.token)
		if t.script == 0 {
			scan.setError(e)
			scan.gobble()
		}
		end = scan.scan()
	}
	if n := len(scan.token); n >= 2 && n <= 3 {
		t.region, e = getRegionID(scan.token)
		if t.region == 0 {
			scan.setError(e)
			scan.gobble()
		} else {
			scan.replace(t.region.String())
		}
		end = scan.scan()
	}
	scan.toLower(scan.start, len(scan.b))
	t.pVariant = byte(end)
	end = parseVariants(scan, end)
	t.pExt = uint16(end)
	return t, end
}

// parseVariants scans tokens as long as each token is a valid variant string.
// Duplicate variants are removed.
func parseVariants(scan *scanner, end int) int {
	start := scan.start
	for ; len(scan.token) >= 4; scan.scan() {
		// TODO: validate and sort variants
		if bytes.Index(scan.b[start:scan.start], scan.token) != -1 {
			scan.gobble()
			continue
		}
		end = scan.end
		const maxVariantSize = 60000 // more than enough, ensures pExt will be valid.
		if end > maxVariantSize {
			break
		}
	}
	return end
}

type bytesSort [][]byte

func (b bytesSort) Len() int {
	return len(b)
}

func (b bytesSort) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b bytesSort) Less(i, j int) bool {
	return bytes.Compare(b[i], b[j]) == -1
}

// parseExtensions parses and normalizes the extensions in the buffer.
// It returns the last position of scan.b that is part of any extension.
// TODO: return errors.
func parseExtensions(scan *scanner) int {
	start := scan.start
	exts := [][]byte{}
	private := []byte{}
	end := scan.end
	for len(scan.token) == 1 {
		start := scan.start
		extension := []byte{}
		ext := scan.token[0]
		switch ext {
		case 'u':
			attrEnd := scan.acceptMinSize(3)
			end = attrEnd
			var key []byte
			for last := []byte{}; len(scan.token) == 2; last = key {
				key = scan.token
				end = scan.acceptMinSize(3)
				// TODO: check key value validity
				if bytes.Compare(key, last) != 1 {
					p := attrEnd + 1
					scan.next = p
					keys := [][]byte{}
					for scan.scan(); len(scan.token) == 2; {
						keyStart := scan.start
						end = scan.acceptMinSize(3)
						keys = append(keys, scan.b[keyStart:end])
					}
					sort.Sort(bytesSort(keys))
					copy(scan.b[p:], bytes.Join(keys, []byte{'-'}))
					break
				}
			}
		case 't':
			scan.scan()
			if n := len(scan.token); n >= 2 && n <= 3 && isAlpha(scan.token[1]) {
				_, end = parseTag(scan)
				scan.toLower(start, end)
			}
			for len(scan.token) == 2 && !isAlpha(scan.token[1]) {
				end = scan.acceptMinSize(3)
			}
		case 'x':
			end = scan.acceptMinSize(1)
		default:
			end = scan.acceptMinSize(2)
		}
		extension = scan.b[start:end]
		if len(extension) < 3 {
			scan.setErrorf("language: empty extension %q", string(ext))
			continue
		} else if len(exts) == 0 && (ext == 'x' || scan.next >= len(scan.b)) {
			return end
		} else if ext == 'x' {
			private = extension
			break
		}
		exts = append(exts, extension)
	}
	if scan.next < len(scan.b) {
		scan.setErrorf("language: invalid trailing characters %q", scan.b[scan.end:])
	}
	sort.Sort(bytesSort(exts))
	if len(private) > 0 {
		exts = append(exts, private)
	}
	scan.b = append(scan.b[:start], bytes.Join(exts, []byte{'-'})...)
	return len(scan.b)
}

func parsePrivate(scan *scanner) int {
	if len(scan.token) == 0 || scan.token[0] != 'x' {
		scan.setErrorf("language: invalid language tag %q", scan.b)
		return scan.start
	}
	return parseExtensions(scan)
}

// A Part identifies a part of the language tag.
type Part byte

const (
	TagPart Part = iota // The tag excluding extensions.
	LanguagePart
	ScriptPart
	RegionPart
	VariantPart
)

var partNames = []string{"Tag", "Language", "Script", "Region", "Variant"}

func (p Part) String() string {
	if p > VariantPart {
		return string(p)
	}
	return partNames[p]
}

// Extension returns the Part identifier for extension e, which must be 0-9 or a-z.
func Extension(e byte) Part {
	return Part(e)
}

// Compose returns a language tag composed from the given parts or an error
// if any of the strings for the parts are ill-formed.
func Compose(m map[Part]string) (t Tag, err error) {
	t = und
	var scan scanner
	scan.b = scan.bytes[:0]
	add := func(p Part) {
		if s, ok := m[p]; ok {
			if len(scan.b) > 0 {
				scan.b = append(scan.b, '-')
			}
			if p > VariantPart {
				scan.b = append(scan.b, byte(p), '-')
			}
			scan.b = append(scan.b, s...)
		}
	}
	for p := TagPart; p <= VariantPart; p++ {
		if p == TagPart && m[p] != "" {
			for i := LanguagePart; i <= VariantPart; i++ {
				if _, ok := m[i]; ok {
					return und, fmt.Errorf("language: cannot specify both Tag and %s", partNames[i])
				}
			}
		}
		add(p)
	}
	for p := Part('0'); p < Part('9'); p++ {
		add(p)
	}
	for p := Part('a'); p < Part('w'); p++ {
		add(p)
	}
	for p := Part('y'); p < Part('z'); p++ {
		add(p)
	}
	add(Part('x'))
	scan.init()
	if len(scan.token) >= 4 {
		if !strings.EqualFold(string(scan.b), "root") {
			return und, errInvalid
		}
		return und, nil
	}
	return parse(&scan, "")
}

// Part returns the part of the language tag indicated by p.
// The one-letter section identifier, if applicable, is not included.
// Components are separated by a '-'.
func (t Tag) Part(p Part) string {
	s := ""
	switch p {
	case TagPart:
		s = t.String()
		if t.pExt > 0 {
			s = s[:t.pExt]
		}
	case LanguagePart:
		s = t.lang.String()
	case ScriptPart:
		if t.script != 0 {
			s = t.script.String()
		}
	case RegionPart:
		if t.region != 0 {
			s = t.region.String()
		}
	case VariantPart:
		if t.str != nil && uint16(t.pVariant) < t.pExt {
			s = (*t.str)[t.pVariant+1 : t.pExt]
		}
	default:
		if t.str != nil {
			str := *t.str
			for i := int(t.pExt); i < len(str)-1; {
				end, name, ext := getExtension(str, i)
				if name == byte(p) {
					return ext
				}
				i = end
			}
		}
	}
	return s
}

// Parts returns all parts of the language tag in a map.
func (t Tag) Parts() map[Part]string {
	m := make(map[Part]string)
	m[LanguagePart] = t.lang.String()
	if t.script != 0 {
		m[ScriptPart] = t.script.String()
	}
	if t.region != 0 {
		m[RegionPart] = t.region.String()
	}
	if t.str != nil {
		s := *t.str
		if uint16(t.pVariant) < t.pExt {
			m[VariantPart] = s[t.pVariant+1 : t.pExt]
		}
		for i := int(t.pExt); i < len(s)-1; {
			end, name, ext := getExtension(s, i)
			m[Extension(name)] = ext
			i = end
		}
	}
	return m
}

// getExtension returns the name, body and end position of the extension.
func getExtension(s string, p int) (end int, name byte, ext string) {
	if s[p] == '-' {
		p++
	}
	if s[p] == 'x' {
		return len(s), s[p], s[p+2:]
	}
	end = nextExtension(s, p)
	return end, s[p], s[p+2 : end]
}

// nextExtension finds the next extension within the string, searching
// for the -<char>- pattern from position p.
// In the fast majority of cases, language tags will have at most
// one extension and extensions tend to be small.
func nextExtension(s string, p int) int {
	for n := len(s) - 3; p < n; {
		if s[p] == '-' {
			if s[p+2] == '-' {
				return p
			}
			p += 3
		} else {
			p++
		}
	}
	return len(s)
}

var (
	acceptErr = errors.New("ParseAcceptLanguage: syntax error")
	acceptRe  = regexp.MustCompile(`^ *(?:([\w-]+|\*)(?: *; *q *= *([0-9\.]+))?)? *$`)
)

// ParseAcceptLanguage parses the contents of a Accept-Language header as
// defined in http://www.google.com/url?q=http://www.ietf.org/rfc/rfc2616.txt
// and returns a list of Tags and a list of corresponding quality weights.
// The Tags will be sorted by highest weight first and then by first occurrence.
// Tags with a weight of zero will be dropped. An error will be returned if the
// input could not be parsed.
func ParseAcceptLanguage(s string) (tag []Tag, q []float32, err error) {
	for start, end := 0, 0; start < len(s); start = end + 1 {
		for end = start; end < len(s) && s[end] != ','; end++ {
		}
		m := acceptRe.FindStringSubmatch(s[start:end])
		if m == nil {
			return nil, nil, acceptErr
		}
		if len(m[1]) > 0 {
			w := 1.0
			if len(m[2]) > 0 {
				if w, err = strconv.ParseFloat(m[2], 32); err != nil {
					return nil, nil, err
				}
				// Drop tags with a quality weight of 0.
				if w <= 0 {
					continue
				}
			}
			t, err := Parse(m[1])
			if err != nil {
				id, ok := acceptFallback[m[1]]
				if !ok {
					return nil, nil, err
				}
				t = Tag{lang: id}
			}
			tag = append(tag, t)
			q = append(q, float32(w))
		}
	}
	sortStable(&tagSort{tag, q})
	return tag, q, nil
}

// Add hack mapping to deal with a small number of cases that that occur
// in Accept-Language (with reasonable frequency).
var acceptFallback = map[string]langID{
	"english": lang_en,
	"deutsch": lang_de,
	"italian": lang_it,
	"french":  lang_fr,
	"*":       lang_mul, // defined in the spec to match all languages.
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
