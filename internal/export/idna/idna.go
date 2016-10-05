// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:generate go run gen.go gen_trieval.go gen_common.go

// http://www.unicode.org/reports/tr46

// Package idna implements IDNA2008 using the compatibility processing
// defined by UTS (Unicode Technical Standard) #46, which defines a standard to
// deal with the transition from IDNA2003.
//
// IDNA2008 (Internationalized Domain Names for Applications), is defined in RFC
// 5890, RFC 5891, RFC 5892, RFC 5893 and RFC 5894.
// UTS #46 is defined in http://www.unicode.org/reports/tr46.
// See http://unicode.org/cldr/utility/idna.jsp for a visualization of the
// differences between these two standards.
package idna // import "golang.org/x/text/internal/export/idna"
