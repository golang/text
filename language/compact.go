// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package language

import "golang.org/x/text/language/internal"

type compactID uint16

func getCoreIndex(t language.Tag) (id compactID, ok bool) {
	x, ok := coreTags[language.GetCoreKey(t)]
	return compactID(x), ok
}
