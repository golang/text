// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

// This program generates tables.go:
//	go run maketables.go | gofmt > tables.go

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"unicode/utf8"

	"code.google.com/p/go.text/encoding"
)

const ascii = "\x00\x01\x02\x03\x04\x05\x06\x07\x08\x09\x0a\x0b\x0c\x0d\x0e\x0f" +
	"\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f" +
	` !"#$%&'()*+,-./0123456789:;<=>?` +
	`@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_` +
	"`abcdefghijklmnopqrstuvwxyz{|}~\u007f"

var encodings = []struct {
	name        string
	varName     string
	replacement byte
	mapping     string
}{
	{
		"IBM Code Page 437",
		"CodePage437",
		encoding.ASCIISub,
		ascii +
			"ÇüéâäàåçêëèïîìÄÅÉæÆôöòûùÿÖÜ¢£¥₧ƒ" +
			"áíóúñÑªº¿⌐¬½¼¡«»░▒▓│┤╡╢╖╕╣║╗╝╜╛┐" +
			"└┴┬├─┼╞╟╚╔╩╦╠═╬╧╨╤╥╙╘╒╓╫╪┘┌█▄▌▐▀" +
			"αßΓπΣσµτΦΘΩδ∞∅∈∩≡±≥≤⌠⌡÷≈°•·√ⁿ²∎\u00a0",
	},
	{
		"IBM Code Page 866",
		"CodePage866",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-ibm866.txt",
	},
	{
		"ISO 8859-2",
		"ISO8859_2",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-iso-8859-2.txt",
	},
	{
		"ISO 8859-3",
		"ISO8859_3",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-iso-8859-3.txt",
	},
	{
		"ISO 8859-4",
		"ISO8859_4",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-iso-8859-4.txt",
	},
	{
		"ISO 8859-5",
		"ISO8859_5",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-iso-8859-5.txt",
	},
	{
		"ISO 8859-6",
		"ISO8859_6",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-iso-8859-6.txt",
	},
	{
		"ISO 8859-7",
		"ISO8859_7",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-iso-8859-7.txt",
	},
	{
		"ISO 8859-8",
		"ISO8859_8",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-iso-8859-8.txt",
	},
	{
		"ISO 8859-10",
		"ISO8859_10",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-iso-8859-10.txt",
	},
	{
		"ISO 8859-13",
		"ISO8859_13",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-iso-8859-13.txt",
	},
	{
		"ISO 8859-14",
		"ISO8859_14",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-iso-8859-14.txt",
	},
	{
		"ISO 8859-15",
		"ISO8859_15",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-iso-8859-15.txt",
	},
	{
		"ISO 8859-16",
		"ISO8859_16",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-iso-8859-16.txt",
	},
	{
		"KOI8-R",
		"KOI8R",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-koi8-r.txt",
	},
	{
		"KOI8-U",
		"KOI8U",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-koi8-u.txt",
	},
	{
		"Macintosh",
		"Macintosh",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-macintosh.txt",
	},
	{
		"Macintosh Cyrillic",
		"MacintoshCyrillic",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-x-mac-cyrillic.txt",
	},
	{
		"Windows 874",
		"Windows874",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-windows-874.txt",
	},
	{
		"Windows 1250",
		"Windows1250",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-windows-1250.txt",
	},
	{
		"Windows 1251",
		"Windows1251",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-windows-1251.txt",
	},
	{
		"Windows 1252",
		"Windows1252",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-windows-1252.txt",
	},
	{
		"Windows 1253",
		"Windows1253",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-windows-1253.txt",
	},
	{
		"Windows 1254",
		"Windows1254",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-windows-1254.txt",
	},
	{
		"Windows 1255",
		"Windows1255",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-windows-1255.txt",
	},
	{
		"Windows 1256",
		"Windows1256",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-windows-1256.txt",
	},
	{
		"Windows 1257",
		"Windows1257",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-windows-1257.txt",
	},
	{
		"Windows 1258",
		"Windows1258",
		encoding.ASCIISub,
		"http://encoding.spec.whatwg.org/index-windows-1258.txt",
	},
}

func getWHATWG(url string) string {
	res, err := http.Get(url)
	if err != nil {
		log.Fatalf("%q: Get: %v", url, err)
	}
	defer res.Body.Close()

	mapping := make([]rune, 128)
	for i := range mapping {
		mapping[i] = '\ufffd'
	}

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		s := strings.TrimSpace(scanner.Text())
		if s == "" || s[0] == '#' {
			continue
		}
		x, y := 0, 0
		if _, err := fmt.Sscanf(s, "%d\t0x%x", &x, &y); err != nil {
			log.Fatalf("could not parse %q", s)
		}
		if x < 0 || 128 <= x {
			log.Fatalf("code %d is out of range", x)
		}
		if 0x80 <= y && y < 0xa0 {
			// We diverge from the WHATWG spec by mapping control characters
			// in the range [0x80, 0xa0) to U+FFFD.
			continue
		}
		mapping[x] = rune(y)
	}
	return ascii + string(mapping)
}

func main() {
	buf := make([]byte, 8)
	fmt.Printf("// generated by go run maketables.go; DO NOT EDIT\n\n")
	fmt.Printf("package charmap\n\n")
	fmt.Printf("import \"code.google.com/p/go.text/encoding\"\n\n")
	for _, e := range encodings {
		if strings.HasPrefix(e.mapping, "http://encoding.spec.whatwg.org/") {
			e.mapping = getWHATWG(e.mapping)
		}

		asciiSuperset, low := strings.HasPrefix(e.mapping, ascii), 0x00
		if asciiSuperset {
			low = 0x80
		}
		lowerVarName := strings.ToLower(e.varName[:1]) + e.varName[1:]
		if strings.HasPrefix(e.varName, "ISO") {
			lowerVarName = "iso" + e.varName[3:]
		}
		fmt.Printf("// %s is the %s encoding.\n", e.varName, e.name)
		fmt.Printf("var %s encoding.Encoding = &%s\n\nvar %s = charmap{\nname: %q,\n",
			e.varName, lowerVarName, lowerVarName, e.name)
		fmt.Printf("asciiSuperset: %t,\n", asciiSuperset)
		fmt.Printf("low: 0x%02x,\n", low)
		fmt.Printf("replacement: 0x%02x,\n", e.replacement)

		fmt.Printf("decode: [256]utf8Enc{\n")
		i, backMapping := 0, map[rune]byte{}
		for _, c := range e.mapping {
			if _, ok := backMapping[c]; !ok {
				backMapping[c] = byte(i)
			}
			for j := range buf {
				buf[j] = 0
			}
			n := utf8.EncodeRune(buf, c)
			if n > 3 {
				panic(fmt.Sprintf("rune %q (%U) is too long", c, c))
			}
			fmt.Printf("{%d,[3]byte{0x%02x,0x%02x,0x%02x}},", n, buf[0], buf[1], buf[2])
			if i%2 == 1 {
				fmt.Printf("\n")
			}
			i++
		}
		fmt.Printf("},\n")

		fmt.Printf("encode: [256]uint32{\n")
		encode := make([]uint32, 0, 256)
		for c, i := range backMapping {
			encode = append(encode, uint32(i)<<24|uint32(c))
		}
		sort.Sort(byRune(encode))
		for len(encode) < cap(encode) {
			encode = append(encode, encode[len(encode)-1])
		}
		for i, enc := range encode {
			fmt.Printf("0x%08x,", enc)
			if i%8 == 7 {
				fmt.Printf("\n")
			}
		}
		fmt.Printf("},\n}\n")
	}
}

type byRune []uint32

func (b byRune) Len() int           { return len(b) }
func (b byRune) Less(i, j int) bool { return b[i]&0xffffff < b[j]&0xffffff }
func (b byRune) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
