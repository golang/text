// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

// This program generates the trie for width operations. The generated table
// includes width category information as well as the normalization mappings.
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"

	"golang.org/x/text/internal/gen"
	"golang.org/x/text/internal/triegen"
)

// See gen_common.go for flags.

func main() {
	gen.Init()
	genTables()
	genTests()
	repackage("gen_trieval.go", "trieval.go")
	repackage("gen_common.go", "common_test.go")
}

func genTables() {
	t := triegen.NewTrie("width")

	getWidthData(func(r rune, tag elem, alt rune) {
		t.Insert(r, uint64(tag|elem(alt&0x3FFF)))
	})

	w := &bytes.Buffer{}

	sz, err := t.Gen(w)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "// Total table size %d bytes (%dKiB)\n", sz, sz/1024)

	gen.WriteGoFile(*outputFile, "width", w.Bytes())
}

func genTests() {
	type keyval struct{ key, val rune }
	m := []keyval{}

	getWidthData(func(r rune, tag elem, alt rune) {
		if alt != 0 {
			m = append(m, keyval{r, alt})
		}
	})

	w := &bytes.Buffer{}

	fmt.Fprintf(w, "\nvar foldRunes = map[rune]rune{\n")
	for _, kv := range m {
		fmt.Fprintf(w, "\t0x%X: 0x%X,\n", kv.key, kv.val)
	}
	fmt.Fprintln(w, "}")
	gen.WriteGoFile("runes_test.go", "width", w.Bytes())
}

// repackage rewrites a file from belonging to package main to belonging to
// package width.
func repackage(inFile, outFile string) {
	src, err := ioutil.ReadFile(inFile)
	if err != nil {
		log.Fatalf("reading %s: %v", inFile, err)
	}
	const toDelete = "package main\n\n"
	i := bytes.Index(src, []byte(toDelete))
	if i < 0 {
		log.Fatalf("Could not find %q in gen_trieval.go", toDelete)
	}
	w := &bytes.Buffer{}
	w.Write(src[i+len(toDelete):])
	gen.WriteGoFile(outFile, "width", w.Bytes())
}
