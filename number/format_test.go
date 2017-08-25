// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package number

import (
	"testing"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func TestWrongVerb(t *testing.T) {
	testCases := []struct {
		f    Formatter
		fmt  string
		want string
	}{{
		f:    Decimal(12),
		fmt:  "%e",
		want: "%!e(int=12)",
	}, {
		f:    Scientific(12),
		fmt:  "%f",
		want: "%!f(int=12)",
	}, {
		f:    Engineering(12),
		fmt:  "%f",
		want: "%!f(int=12)",
	}, {
		f:    Percent(12),
		fmt:  "%e",
		want: "%!e(int=12)",
	}}
	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			tag := language.Und
			got := message.NewPrinter(tag).Sprintf(tc.fmt, tc.f)
			if got != tc.want {
				t.Errorf("got %q; want %q", got, tc.want)
			}
		})
	}
}
