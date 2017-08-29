// Copyright 2017 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package language implements BCP 47 language tags and related functionality.
//
// The Tag type, which is used to represent languages, is agnostic to the
// meaning of its subtags. Tags are not fully canonicalized to preserve
// information that may be valuable in certain contexts. As a consequence, two
// different tags may represent identical languages.
//
// Initializing language- or locale-specific components usually consists of
// two steps. The first step is to select a display language based on the
// preferred languages of the user and the languages supported by an application.
// The second step is to create the language-specific services based on
// this selection. Each is discussed in more details below.
//
// Matching preferred against supported languages
//
// An application may support various languages. This list is typically limited
// by the languages for which there exists translations of the user interface.
// Similarly, a user may provide a list of preferred languages which is limited
// by the languages understood by this user.
// An application should use a Matcher to find the best supported language based
// on the user's preferred list.
// Matchers are aware of the intricacies of equivalence between languages.
// The default Matcher implementation takes into account things such as
// deprecated subtags, legacy tags, and mutual intelligibility between scripts
// and languages.
//
// A Matcher for English, Australian English, Danish, and standard Mandarin can
// be defined as follows:
//
//		var matcher = language.NewMatcher([]language.Tag{
//			language.English,   // The first language is used as fallback.
// 			language.MustParse("en-AU"),
//			language.Danish,
//			language.Chinese,
//		})
//
// The following code selects the best match for someone speaking Spanish and
// Norwegian:
//
// 		preferred := []language.Tag{ language.Spanish, language.Norwegian }
//		tag, _, _ := matcher.Match(preferred...)
//
// In this case, the best match is Danish, as Danish is sufficiently a match to
// Norwegian to not have to fall back to the default.
// See ParseAcceptLanguage on how to handle the Accept-Language HTTP header.
//
// Selecting language-specific services
//
// One should always use the Tag returned by the Matcher to create an instance
// of any of the language-specific services provided by the text repository.
// This prevents the mixing of languages, such as having a different language for
// messages and display names, as well as improper casing or sorting order for
// the selected language.
// Using the returned Tag also allows user-defined settings, such as collation
// order or numbering system to be transparently passed as options.
//
// If you have language-specific data in your application, however, it will in
// most cases suffice to use the index returned by the matcher to identify
// the user language.
// The following loop provides an alternative in case this is not sufficient:
//
// 		supported := map[language.Tag]data{
//			language.English:            enData,
// 			language.MustParse("en-AU"): enAUData,
//			language.Danish:             daData,
//			language.Chinese:            zhData,
// 		}
//		tag, _, _ := matcher.Match(preferred...)
//		for ; tag != language.Und; tag = tag.Parent() {
//			if v, ok := supported[tag]; ok {
//				return v
//			}
//		}
// 		return enData // should not reach here
//
// Repeatedly taking the Parent of the tag returned by Match will eventually
// match one of the tags used to initialize the Matcher.
//
// Canonicalization
//
// By default, only legacy and deprecated tags are converted into their
// canonical equivalent. All other information is preserved. This approach makes
// the confidence scores more accurate and allows matchers to distinguish
// between variants that are otherwise lost.
//
// As a consequence, two tags that should be treated as identical according to
// BCP 47 or CLDR, like "en-Latn" and "en", will be represented differently. The
// Matchers will handle such distinctions, though, and are aware of the
// equivalence relations. The CanonType type can be used to alter the
// canonicalization form.
//
// References
//
// BCP 47 - Tags for Identifying Languages
// http://tools.ietf.org/html/bcp47
package language // import "golang.org/x/text/language"
