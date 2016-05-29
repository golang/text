// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

type property int

// The order of these constants matter. A Profile may consider runes to be
// allowed either from pValid or idDisOrFreePVal.
const (
	unassigned property = iota
	disallowed
	contextO
	contextJ
	idDisOrFreePVal
	pValid
)
