// Copyright 2026 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package collate

import (
	"bytes"
	"testing"

	"golang.org/x/text/language"
)

func TestForceKeys(t *testing.T) {
	inputs := []string{"", "\x00", "a", "A", "á", "a\u0301", "a\x00", "a\x01", "aa", "\xff", "Québécois", "Que\u0301be\u0301cois"}
	for _, test := range []struct {
		name string
		opts []Option
	}{
		{"Force", []Option{Force}},
		{"LooseForce", []Option{Loose, Force}},
	} {
		t.Run(test.name, func(t *testing.T) {
			c := New(language.English, test.opts...)
			for _, a := range inputs {
				for _, b := range inputs {
					want := c.CompareString(a, b)
					var buf Buffer
					ka, kb := c.Key(&buf, []byte(a)), c.Key(&buf, []byte(b))
					if got := bytes.Compare(ka, kb); got != want {
						t.Errorf("Key(%q, %q) order = %d, want %d", a, b, got, want)
					}
					buf.Reset()
					ka, kb = c.KeyFromString(&buf, a), c.KeyFromString(&buf, b)
					if got := bytes.Compare(ka, kb); got != want {
						t.Errorf("KeyFromString(%q, %q) order = %d, want %d", a, b, got, want)
					}
				}
			}
		})
	}
}
