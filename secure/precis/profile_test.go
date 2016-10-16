// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package precis

import (
	"math/rand"
	"testing"
	"unicode"

	"golang.org/x/text/transform"
)

// copyOrbit is a Transformer for the sole purpose of testing the apply method,
// testing that apply will always call Span for the prefix of the input that
// remains identical and then call Transform for the remainder. It will produce
// inconsistent output for other usage patterns.
// Provided that copyOrbit is used this way, the first t bytes of the output
// will be identical to the input and the remaining output will be the result
// of calling caseOrbit on the remaining input bytes.
type copyOrbit int

func (t copyOrbit) Reset() {}
func (t copyOrbit) Span(src []byte, atEOF bool) (n int, err error) {
	if int(t) == len(src) {
		return int(t), nil
	}
	return int(t), transform.ErrEndOfSpan
}

// Transform implements transform.Transformer specifically for testing the apply method.
// See documentation of copyOrbit before using this method.
func (t copyOrbit) Transform(dst, src []byte, atEOF bool) (nDst, nSrc int, err error) {
	n := copy(dst, src)
	for i, c := range dst[:n] {
		dst[i] = orbitCase(c)
	}
	return n, n, nil
}

func orbitCase(c byte) byte {
	if unicode.IsLower(rune(c)) {
		return byte(unicode.ToUpper(rune(c)))
	} else {
		return byte(unicode.ToLower(rune(c)))
	}
}

func TestBuffers(t *testing.T) {
	want := "Those who cannot remember the past are condemned to compute it."

	spans := rand.Perm(len(want) + 1)

	// Compute the result of applying copyOrbit(span) transforms in reverse.
	input := []byte(want)
	for i := len(spans) - 1; i >= 0; i-- {
		for j := spans[i]; j < len(input); j++ {
			input[j] = orbitCase(input[j])
		}
	}

	// Apply the copyOrbit(span) transforms.
	b := buffers{src: input}
	for _, n := range spans {
		b.apply(copyOrbit(n))
		if n%11 == 0 {
			b.apply(transform.Nop)
		}
	}
	if got := string(b.src); got != want {
		t.Errorf("got %q; want %q", got, want)
	}
}
