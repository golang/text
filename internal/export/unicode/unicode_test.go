// Copyright 2019 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unicode

import (
	"testing"
	"unicode"

	"github.com/liquid-dev/text/internal/gen"
	"github.com/liquid-dev/text/internal/testtext"
	"github.com/liquid-dev/text/internal/ucd"
)

// TestScripts tests for all runes whether they are included in the correct
// script and, indirectly, whether each script exists.
func TestScripts(t *testing.T) {
	testtext.SkipIfNotLong(t)

	ucd.Parse(gen.OpenUCDFile("Scripts.txt"), func(p *ucd.Parser) {
		r := p.Rune(0)
		script := p.String(1)
		if !unicode.Is(unicode.Scripts[script], r) {
			t.Errorf("%U: not in script %q", r, script)
		}
	})
}
