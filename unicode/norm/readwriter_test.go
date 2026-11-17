// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package norm

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"
)

var bufSizes = []int{1, 2, 3, 4, 5, 6, 7, 8, 100, 101, 102, 103, 4000, 4001, 4002, 4003}

func readFunc(size int) appendFunc {
	return func(f Form, out []byte, s string) []byte {
		out = append(out, s...)
		r := f.Reader(bytes.NewBuffer(out))
		buf := make([]byte, size)
		result := []byte{}
		for n, err := 0, error(nil); err == nil; {
			n, err = r.Read(buf)
			result = append(result, buf[:n]...)
		}
		return result
	}
}

func TestReader(t *testing.T) {
	for _, s := range bufSizes {
		name := fmt.Sprintf("TestReader%d", s)
		runNormTests(t, name, readFunc(s))
	}
}

func TestReaderSplitRune(t *testing.T) {
	input := strings.Repeat("\u0634\u064e\u0651", 16)
	var readers []io.Reader
	for i := 0; i < len(input); i += 3 {
		readers = append(readers, strings.NewReader(input[i:i+3]))
	}
	// MultiReader returns at most one piece per Read, splitting every other
	// piece inside a combining mark.
	got, err := io.ReadAll(NFC.Reader(io.MultiReader(readers...)))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != input {
		t.Errorf("Reader output = %q; want %q", got, input)
	}
}

func writeFunc(size int) appendFunc {
	return func(f Form, out []byte, s string) []byte {
		in := append(out, s...)
		result := new(bytes.Buffer)
		w := f.Writer(result)
		buf := make([]byte, size)
		for n := 0; len(in) > 0; in = in[n:] {
			n = copy(buf, in)
			_, _ = w.Write(buf[:n])
		}
		w.Close()
		return result.Bytes()
	}
}

func TestWriterSplitRune(t *testing.T) {
	input := strings.Repeat("\u0634\u064e\u0651", 16)
	got := writeFunc(3)(NFC, nil, input)
	if string(got) != input {
		t.Errorf("Writer output = %q; want %q", got, input)
	}
}

func TestWriter(t *testing.T) {
	for _, s := range bufSizes {
		name := fmt.Sprintf("TestWriter%d", s)
		runNormTests(t, name, writeFunc(s))
	}
}
