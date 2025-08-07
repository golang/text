// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package message_test

import (
	"errors"
	"reflect"
	"testing"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

func TestErorrf(t *testing.T) {
	wrapped := errors.New("inner error")
	p := message.NewPrinter(language.Und)
	for _, test := range []struct {
		err        error
		wantText   string
		wantUnwrap error
		wantSplit  []error
	}{{
		err:        p.Errorf("%w", wrapped),
		wantText:   "inner error",
		wantUnwrap: wrapped,
	}, {
		err:        p.Errorf("added context: %w", wrapped),
		wantText:   "added context: inner error",
		wantUnwrap: wrapped,
	}, {
		err:        p.Errorf("%w with added context", wrapped),
		wantText:   "inner error with added context",
		wantUnwrap: wrapped,
	}, {
		err:        p.Errorf("%s %w %v", "prefix", wrapped, "suffix"),
		wantText:   "prefix inner error suffix",
		wantUnwrap: wrapped,
	}, {
		err:        p.Errorf("%[2]s: %[1]w", wrapped, "positional verb"),
		wantText:   "positional verb: inner error",
		wantUnwrap: wrapped,
	}, {
		err:      p.Errorf("%v", wrapped),
		wantText: "inner error",
	}, {
		err:      p.Errorf("added context: %v", wrapped),
		wantText: "added context: inner error",
	}, {
		err:      p.Errorf("%v with added context", wrapped),
		wantText: "inner error with added context",
	}, {
		err:      p.Errorf("%w is not an error", "not-an-error"),
		wantText: "%!w(string=not-an-error) is not an error",
	}, {
		err:       p.Errorf("wrapped two errors: %w %w", errString("1"), errString("2")),
		wantText:  "wrapped two errors: 1 2",
		wantSplit: []error{errString("1"), errString("2")},
	}, {
		err:       p.Errorf("wrapped three errors: %w %w %w", errString("1"), errString("2"), errString("3")),
		wantText:  "wrapped three errors: 1 2 3",
		wantSplit: []error{errString("1"), errString("2"), errString("3")},
	}, {
		err:       p.Errorf("wrapped nil error: %w %w %w", errString("1"), nil, errString("2")),
		wantText:  "wrapped nil error: 1 %!w(<nil>) 2",
		wantSplit: []error{errString("1"), errString("2")},
	}, {
		err:       p.Errorf("wrapped one non-error: %w %w %w", errString("1"), "not-an-error", errString("3")),
		wantText:  "wrapped one non-error: 1 %!w(string=not-an-error) 3",
		wantSplit: []error{errString("1"), errString("3")},
	}, {
		err:       p.Errorf("wrapped errors out of order: %[3]w %[2]w %[1]w", errString("1"), errString("2"), errString("3")),
		wantText:  "wrapped errors out of order: 3 2 1",
		wantSplit: []error{errString("1"), errString("2"), errString("3")},
	}, {
		err:       p.Errorf("wrapped several times: %[1]w %[1]w %[2]w %[1]w", errString("1"), errString("2")),
		wantText:  "wrapped several times: 1 1 2 1",
		wantSplit: []error{errString("1"), errString("2")},
	}, {
		err:        p.Errorf("%w", nil),
		wantText:   "%!w(<nil>)",
		wantUnwrap: nil, // still nil
	}} {
		if got, want := errors.Unwrap(test.err), test.wantUnwrap; got != want {
			t.Errorf("Formatted error: %v\nerrors.Unwrap() = %v, want %v", test.err, got, want)
		}
		if got, want := splitErr(test.err), test.wantSplit; !reflect.DeepEqual(got, want) {
			t.Errorf("Formatted error: %v\nUnwrap() []error = %v, want %v", test.err, got, want)
		}
		if got, want := test.err.Error(), test.wantText; got != want {
			t.Errorf("err.Error() = %q, want %q", got, want)
		}
	}
}

func splitErr(err error) []error {
	if e, ok := err.(interface{ Unwrap() []error }); ok {
		return e.Unwrap()
	}
	return nil
}

type errString string

func (e errString) Error() string { return string(e) }
