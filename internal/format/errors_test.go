// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package format_test

import (
	"errors"
	"slices"
	"testing"

	"golang.org/x/text/internal/format"
)

func TestErrorf(t *testing.T) {
	wrapped := errors.New("inner error")
	for _, test := range []struct {
		fmtStr      string
		args        []any
		wantWrapped []int
	}{
		0: {
			fmtStr:      "%w",
			args:        []any{wrapped},
			wantWrapped: []int{0},
		},
		1: {
			fmtStr:      "%w %v%w",
			args:        []any{wrapped, 1, wrapped},
			wantWrapped: []int{0, 2},
		},
	} {
		p := format.Parser{}
		p.Reset(test.args)
		p.SetFormat(test.fmtStr)
		for p.Scan() {
		}
		if slices.Compare(test.wantWrapped, p.WrappedErrs) != 0 {
			t.Errorf("wrong wrapped: got=%v, want=%v", p.WrappedErrs, test.wantWrapped)
		}
	}
}
