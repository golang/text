// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package precis

import (
	"testing"
)

func BenchmarkEnforce(b *testing.B) {
	for i := 0; i < b.N; i++ {
		UsernameCaseMapped.String("Malvolio")
	}
}
