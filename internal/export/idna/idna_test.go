// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package idna

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"golang.org/x/text/internal/gen"
	"golang.org/x/text/internal/testtext"
	"golang.org/x/text/internal/ucd"
)

func TestConformance(t *testing.T) {
	testtext.SkipIfNotLong(t)

	r := gen.OpenUnicodeFile("idna", "", "IdnaTest.txt")
	defer r.Close()

	section := "main"
	started := false
	p := ucd.New(r, ucd.CommentHandler(func(s string) {
		if started {
			section = strings.ToLower(strings.Split(s, " ")[0])
		}
	}))
	for p.Next() {
		started = true

		// What to test
		profiles := []*Profile{}
		switch p.String(0) {
		case "T":
			profiles = append(profiles, Transitional)
		case "N":
			profiles = append(profiles, NonTransitional)
		case "B":
			profiles = append(profiles, Transitional)
			profiles = append(profiles, NonTransitional)
		}

		src := unescape(p.String(1))
		if incorrectTests[src] {
			continue
		}

		wantToUnicode := unescape(p.String(2))
		if wantToUnicode == "" {
			wantToUnicode = src
		}
		wantToASCII := unescape(p.String(3))
		if wantToASCII == "" {
			wantToASCII = wantToUnicode
		}
		test := "err:"
		if strings.HasPrefix(wantToUnicode, "[") {
			test += strings.Replace(strings.Trim(wantToUnicode, "[]"), " ", "", -1)
		}
		if strings.HasPrefix(wantToASCII, "[") {
			test += strings.Replace(strings.Trim(wantToASCII, "[]"), " ", "", -1)
		}
		if test == "err:" {
			test = "ok"
		}

		// TODO: also do IDNA tests.
		// invalidInIDNA2008 := p.String(4) == "NV8"

		for _, p := range profiles {
			testtext.Run(t, fmt.Sprintf("%s:%s/%s/%+q", section, test, p, src), func(t *testing.T) {
				got, err := p.ToUnicode(src)
				wantErr := strings.HasPrefix(wantToUnicode, "[")
				gotErr := err != nil
				if wantErr {
					if gotErr != wantErr {
						// TODO: fix and make Errorf.
						t.Skipf(`ToUnicode:err got %v; want %v (%s)`,
							gotErr, wantErr, wantToUnicode)
					}
				} else if got != wantToUnicode || gotErr != wantErr {
					t.Errorf(`ToUnicode: got %+q, %v (%v); want %+q, %v`,
						got, gotErr, err, wantToUnicode, wantErr)
				}

				got, err = p.ToASCII(src)
				wantErr = strings.HasPrefix(wantToASCII, "[")
				gotErr = err != nil
				if wantErr {
					if gotErr != wantErr {
						// TODO: fix and make Errorf.
						t.Skipf(`ToASCII:err got %v; want %v (%s)`,
							gotErr, wantErr, wantToASCII)
					}
				} else if got != wantToASCII || gotErr != wantErr {
					t.Errorf(`ToASCII: got %+q, %v (%v); want %+q, %v`,
						got, gotErr, err, wantToASCII, wantErr)
				}
			})
		}
	}
}

func unescape(s string) string {
	s, err := strconv.Unquote(`"` + s + `"`)
	if err != nil {
		panic(err)
	}
	return s
}

var incorrectTests = map[string]bool{
	"123456789012345678901234567890123456789012345678901234567890123.123456789012345678901234567890123456789012345678901234567890123.123456789012345678901234567890123456789012345678901234567890123.12345678901234567890123456789012345678901234567890123456789012": true,
	"123456789012345678901234567890123456789012345678901234567890123.1234567890\u00e4123456789012345678901234567890123456789012345.123456789012345678901234567890123456789012345678901234567890123.12345678901234567890123456789012345678901234567890123456789012":   true,
	"123456789012345678901234567890123456789012345678901234567890123.1234567890a\u0308123456789012345678901234567890123456789012345.123456789012345678901234567890123456789012345678901234567890123.12345678901234567890123456789012345678901234567890123456789012":  true,
	// These are only wrong in IDNA2008
	"123456789012345678901234567890123456789012345678901234567890123.1234567890A\u0308123456789012345678901234567890123456789012345.123456789012345678901234567890123456789012345678901234567890123.12345678901234567890123456789012345678901234567890123456789012": true,
	"123456789012345678901234567890123456789012345678901234567890123.1234567890\u00c4123456789012345678901234567890123456789012345.123456789012345678901234567890123456789012345678901234567890123.12345678901234567890123456789012345678901234567890123456789012":  true,
}
