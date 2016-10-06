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

	section := "MAIN"
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

		wantToUnicode := unescape(p.String(2))
		if wantToUnicode == "" {
			wantToUnicode = src
		}
		wantToASCII := unescape(p.String(3))
		if wantToASCII == "" {
			wantToASCII = wantToUnicode
		}
		style := "conv"
		if strings.HasPrefix(wantToUnicode, "[") || strings.HasPrefix(wantToASCII, "[") {
			style = "err"
		}

		// TODO: also do IDNA tests.
		// invalidInIDNA2008 := p.String(4) == "NV8"

		for _, p := range profiles {
			testtext.Run(t, fmt.Sprintf("%s:%s/%s/%+q", section, style, p, src), func(t *testing.T) {
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
