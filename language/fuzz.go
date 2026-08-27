// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build gofuzz
// +build gofuzz

package language

func FuzzAcceptLanguage(data []byte) int {
	_, _, _ = ParseAcceptLanguage(string(data))
	return 1
}
