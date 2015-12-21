// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package precis

import (
	"golang.org/x/text/unicode/norm"
	"golang.org/x/text/width"
)

var (
	UsernameCaseMapped    Profile = usernamecasemap   // Implements the UsernameCaseMapped profile specified in RFC 7613.
	UsernameCasePreserved Profile = usernamenocasemap // Implements the UsernameCasePreserved profile specified in RFC 7613.
)

// TODO: mvl: "Ultimately, I would manually define the structs for the internal
// profiles. This avoid pulling in unneeded tables when they are not used."
var (
	usernamecasemap Profile = NewIdentifier(
		AllowWide,
		FoldCase,
		Norm(norm.NFC),
		// TODO: BIDI rule
	)
	usernamenocasemap Profile = NewIdentifier(
		AllowWide,
		Norm(norm.NFC),
		Width(width.Fold), // TODO: Is this correct?
		// TODO: BIDI rule
	)
)
