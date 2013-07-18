// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package encoding_test

import (
	"io"
	"os"
	"strings"

	"code.google.com/p/go.text/encoding"
	"code.google.com/p/go.text/transform"
)

func ExampleDecodeWindows1252() {
	sr := strings.NewReader("Gar\xe7on !")
	tr := transform.NewReader(sr, encoding.Windows1252.NewDecoder())
	io.Copy(os.Stdout, tr)
	// Output: Garçon !
}

/*
TODO: enable this example when transform.Chain is checked in.

func ExampleUTF8Validator() {
	for i := 0; i < 2; i++ {
		transformer := encoding.Windows1252.NewEncoder()
		if i == 1 {
			transformer = transform.Chain(encoding.UTF8Validator, transformer)
		}
		dst := make([]byte, 256)
		src := []byte("abc\xffxyz") // src is invalid UTF-8.
		nDst, nSrc, err := transformer.Transform(dst, src, true)
		fmt.Printf("i=%d: produced %q, consumed %q, error %v\n",
			i, dst[:nDst], src[:nSrc], err)
	}
	// Output:
	// i=0: produced "abc\x1axyz", consumed "abc\xffxyz", error <nil>
	// i=1: produced "abc", consumed "abc", error encoding: invalid UTF-8
}
*/
