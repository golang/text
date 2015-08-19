// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gen

import (
	"bytes"
	"fmt"
	"hash"
	"hash/fnv"
	"io"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"
)

// This file contains utilities for generating code.

// TODO: other write methods like:
// - slices, maps, types, etc.

// CodeWriter is a utility for writing structured code. It computes the content
// hash and size of written content. It ensures there are newlines between
// written code blocks.
type CodeWriter struct {
	buf  bytes.Buffer
	Size int
	Hash hash.Hash32 // content hash
	// For comments we skip the usual one-line separator if they are followed by
	// a code block.
	skipSep bool
}

func (w *CodeWriter) Write(p []byte) (n int, err error) {
	return w.buf.Write(p)
}

// NewCodeWriter returns a new CodeWriter.
func NewCodeWriter() *CodeWriter {
	return &CodeWriter{Hash: fnv.New32()}
}

// WriteGoFile appends the buffer with the total size of all created structures
// and writes it as a Go file to the the given file with the given package name.
func (w *CodeWriter) WriteGoFile(filename, pkg string) {
	sz := w.Size
	w.WriteComment("Total table size %d bytes (%dKiB); checksum: %X\n", sz, sz/1024, w.Hash.Sum32())
	WriteGoFile(filename, pkg, w.buf.Bytes())
	// ioutil.WriteFile(filename, w.buf.Bytes(), 0777)
	w.buf.Reset()
}

func (w *CodeWriter) printf(f string, x ...interface{}) {
	fmt.Fprintf(w, f, x...)
}

func (w *CodeWriter) insertSep() {
	if w.skipSep {
		w.skipSep = false
		return
	}
	// Use at least two newlines to ensure a blank space between the previous
	// block. WriteGoFile will remove extraneous newlines.
	w.printf("\n\n")
}

// WriteComment writes a comment block. All line starts are prefixed with "//".
// Initial empty lines are gobbled. The indentation for the first line is
// stripped from consecutive lines.
func (w *CodeWriter) WriteComment(comment string, args ...interface{}) {
	s := fmt.Sprintf(comment, args...)
	s = strings.Trim(s, "\n")

	// Use at least two newlines to ensure a blank space between the previous
	// block. WriteGoFile will remove extraneous newlines.
	w.printf("\n\n// ")
	w.skipSep = true

	// strip first indent level.
	sep := "\n"
	for ; len(s) > 0 && (s[0] == '\t' || s[0] == ' '); s = s[1:] {
		sep += s[:1]
	}

	strings.NewReplacer(sep, "\n// ", "\n", "\n// ").WriteString(w, s)

	w.printf("\n")
}

func (w *CodeWriter) writeSizeInfo(size int) {
	w.printf("// Size: %d bytes\n", size)
	w.Size += size
}

// WriteConst writes a constant of the given name and value.
func (w *CodeWriter) WriteConst(name string, x interface{}) {
	w.insertSep()
	if s, ok := x.(string); ok {
		w.writeSizeInfo(len(s))
		w.printf("const %s = ", name)
		w.WriteString(s)
		w.printf("\n")
	} else {
		w.printf("const %s = %#v\n", name, x)
	}
}

// WriteVar writes a variable of the given name and value.
func (w *CodeWriter) WriteVar(name, x interface{}) {
	w.insertSep()
	if s, ok := x.(string); ok {
		w.writeSizeInfo(len(s) + int(reflect.TypeOf(s).Size()))
		w.printf("var %s = ", name)
		w.WriteString(s)
		w.printf("\n")
	} else {
		w.printf("var %s = %#v\n", name, x)
	}
}

// WriteString writes a string literal.
func (w *CodeWriter) WriteString(s string) {
	io.WriteString(w.Hash, s) // content hash

	const maxInline = 40
	if len(s) <= maxInline {
		w.printf("%q", s)
		return
	}

	// We will render the string as a multi-line string.
	const maxWidth = 80 - 4 - len(`"`) - len(`" +`)

	// When starting on its own line, go fmt indents line 2+ an extra level.
	n, max := maxWidth, maxWidth-4

	// Print "" +\n, if a string does not start on its own line.
	b := w.buf.Bytes()
	if p := len(bytes.TrimRight(b, " \t")); p > 0 && b[p-1] != '\n' {
		w.printf("\"\" +\n")
		n, max = maxWidth, maxWidth
	}

	w.printf(`"`)

	for sz, p := 0, 0; p < len(s); {
		var r rune
		r, sz = utf8.DecodeRuneInString(s[p:])
		out := s[p : p+sz]
		chars := 1
		if !unicode.IsPrint(r) || r == utf8.RuneError {
			switch sz {
			case 1:
				out = fmt.Sprintf("\\x%02x", s[p])
			case 2, 3:
				out = fmt.Sprintf("\\u%04x", r)
			case 4:
				out = fmt.Sprintf("\\U%08x", r)
			}
			chars = len(out)
		}
		if n -= chars; n < 0 {
			w.printf("\" +\n\"")
			n = max - len(out)
		}
		w.printf("%s", out)
		p += sz
	}
	w.printf(`"`)
}
