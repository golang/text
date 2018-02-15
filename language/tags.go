// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package language

// TODO: Various sets of commonly use tags and regions.

// MustParse is like Parse, but panics if the given BCP 47 tag cannot be parsed.
// It simplifies safe initialization of Tag values.
func MustParse(s string) Tag {
	t, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return t
}

// MustParse is like Parse, but panics if the given BCP 47 tag cannot be parsed.
// It simplifies safe initialization of Tag values.
func (c CanonType) MustParse(s string) Tag {
	t, err := c.Parse(s)
	if err != nil {
		panic(err)
	}
	return t
}

// MustParseBase is like ParseBase, but panics if the given base cannot be parsed.
// It simplifies safe initialization of Base values.
func MustParseBase(s string) Base {
	b, err := ParseBase(s)
	if err != nil {
		panic(err)
	}
	return b
}

// MustParseScript is like ParseScript, but panics if the given script cannot be
// parsed. It simplifies safe initialization of Script values.
func MustParseScript(s string) Script {
	scr, err := ParseScript(s)
	if err != nil {
		panic(err)
	}
	return scr
}

// MustParseRegion is like ParseRegion, but panics if the given region cannot be
// parsed. It simplifies safe initialization of Region values.
func MustParseRegion(s string) Region {
	r, err := ParseRegion(s)
	if err != nil {
		panic(err)
	}
	return r
}

var (
	und = Tag{}

	Und Tag = Tag{}

	// TODO: use compact tags once the transition is completed.

	Afrikaans            Tag = Raw.MustParse("af")      // Tag{lang: _af}                //  af
	Amharic              Tag = Raw.MustParse("am")      // Tag{lang: _am}                //  am
	Arabic               Tag = Raw.MustParse("ar")      // Tag{lang: _ar}                //  ar
	ModernStandardArabic Tag = Raw.MustParse("ar-001")  // Tag{lang: _ar, region: _001}  //  ar-001
	Azerbaijani          Tag = Raw.MustParse("az")      // Tag{lang: _az}                //  az
	Bulgarian            Tag = Raw.MustParse("bg")      // Tag{lang: _bg}                //  bg
	Bengali              Tag = Raw.MustParse("bn")      // Tag{lang: _bn}                //  bn
	Catalan              Tag = Raw.MustParse("ca")      // Tag{lang: _ca}                //  ca
	Czech                Tag = Raw.MustParse("cs")      // Tag{lang: _cs}                //  cs
	Danish               Tag = Raw.MustParse("da")      // Tag{lang: _da}                //  da
	German               Tag = Raw.MustParse("de")      // Tag{lang: _de}                //  de
	Greek                Tag = Raw.MustParse("el")      // Tag{lang: _el}                //  el
	English              Tag = Raw.MustParse("en")      // Tag{lang: _en}                //  en
	AmericanEnglish      Tag = Raw.MustParse("en-US")   // Tag{lang: _en, region: _US}   //  en-US
	BritishEnglish       Tag = Raw.MustParse("en-GB")   // Tag{lang: _en, region: _GB}   //  en-GB
	Spanish              Tag = Raw.MustParse("es")      // Tag{lang: _es}                //  es
	EuropeanSpanish      Tag = Raw.MustParse("es-ES")   // Tag{lang: _es, region: _ES}   //  es-ES
	LatinAmericanSpanish Tag = Raw.MustParse("es-419")  // Tag{lang: _es, region: _419}  //  es-419
	Estonian             Tag = Raw.MustParse("et")      // Tag{lang: _et}                //  et
	Persian              Tag = Raw.MustParse("fa")      // Tag{lang: _fa}                //  fa
	Finnish              Tag = Raw.MustParse("fi")      // Tag{lang: _fi}                //  fi
	Filipino             Tag = Raw.MustParse("fil")     // Tag{lang: _fil}               //  fil
	French               Tag = Raw.MustParse("fr")      // Tag{lang: _fr}                //  fr
	CanadianFrench       Tag = Raw.MustParse("fr-CA")   // Tag{lang: _fr, region: _CA}   //  fr-CA
	Gujarati             Tag = Raw.MustParse("gu")      // Tag{lang: _gu}                //  gu
	Hebrew               Tag = Raw.MustParse("he")      // Tag{lang: _he}                //  he
	Hindi                Tag = Raw.MustParse("hi")      // Tag{lang: _hi}                //  hi
	Croatian             Tag = Raw.MustParse("hr")      // Tag{lang: _hr}                //  hr
	Hungarian            Tag = Raw.MustParse("hu")      // Tag{lang: _hu}                //  hu
	Armenian             Tag = Raw.MustParse("hy")      // Tag{lang: _hy}                //  hy
	Indonesian           Tag = Raw.MustParse("id")      // Tag{lang: _id}                //  id
	Icelandic            Tag = Raw.MustParse("is")      // Tag{lang: _is}                //  is
	Italian              Tag = Raw.MustParse("it")      // Tag{lang: _it}                //  it
	Japanese             Tag = Raw.MustParse("ja")      // Tag{lang: _ja}                //  ja
	Georgian             Tag = Raw.MustParse("ka")      // Tag{lang: _ka}                //  ka
	Kazakh               Tag = Raw.MustParse("kk")      // Tag{lang: _kk}                //  kk
	Khmer                Tag = Raw.MustParse("km")      // Tag{lang: _km}                //  km
	Kannada              Tag = Raw.MustParse("kn")      // Tag{lang: _kn}                //  kn
	Korean               Tag = Raw.MustParse("ko")      // Tag{lang: _ko}                //  ko
	Kirghiz              Tag = Raw.MustParse("ky")      // Tag{lang: _ky}                //  ky
	Lao                  Tag = Raw.MustParse("lo")      // Tag{lang: _lo}                //  lo
	Lithuanian           Tag = Raw.MustParse("lt")      // Tag{lang: _lt}                //  lt
	Latvian              Tag = Raw.MustParse("lv")      // Tag{lang: _lv}                //  lv
	Macedonian           Tag = Raw.MustParse("mk")      // Tag{lang: _mk}                //  mk
	Malayalam            Tag = Raw.MustParse("ml")      // Tag{lang: _ml}                //  ml
	Mongolian            Tag = Raw.MustParse("mn")      // Tag{lang: _mn}                //  mn
	Marathi              Tag = Raw.MustParse("mr")      // Tag{lang: _mr}                //  mr
	Malay                Tag = Raw.MustParse("ms")      // Tag{lang: _ms}                //  ms
	Burmese              Tag = Raw.MustParse("my")      // Tag{lang: _my}                //  my
	Nepali               Tag = Raw.MustParse("ne")      // Tag{lang: _ne}                //  ne
	Dutch                Tag = Raw.MustParse("nl")      // Tag{lang: _nl}                //  nl
	Norwegian            Tag = Raw.MustParse("no")      // Tag{lang: _no}                //  no
	Punjabi              Tag = Raw.MustParse("pa")      // Tag{lang: _pa}                //  pa
	Polish               Tag = Raw.MustParse("pl")      // Tag{lang: _pl}                //  pl
	Portuguese           Tag = Raw.MustParse("pt")      // Tag{lang: _pt}                //  pt
	BrazilianPortuguese  Tag = Raw.MustParse("pt-BR")   // Tag{lang: _pt, region: _BR}   //  pt-BR
	EuropeanPortuguese   Tag = Raw.MustParse("pt-PT")   // Tag{lang: _pt, region: _PT}   //  pt-PT
	Romanian             Tag = Raw.MustParse("ro")      // Tag{lang: _ro}                //  ro
	Russian              Tag = Raw.MustParse("ru")      // Tag{lang: _ru}                //  ru
	Sinhala              Tag = Raw.MustParse("si")      // Tag{lang: _si}                //  si
	Slovak               Tag = Raw.MustParse("sk")      // Tag{lang: _sk}                //  sk
	Slovenian            Tag = Raw.MustParse("sl")      // Tag{lang: _sl}                //  sl
	Albanian             Tag = Raw.MustParse("sq")      // Tag{lang: _sq}                //  sq
	Serbian              Tag = Raw.MustParse("sr")      // Tag{lang: _sr}                //  sr
	SerbianLatin         Tag = Raw.MustParse("sr-Latn") // Tag{lang: _sr, script: _Latn} //  sr-Latn
	Swedish              Tag = Raw.MustParse("sv")      // Tag{lang: _sv}                //  sv
	Swahili              Tag = Raw.MustParse("sw")      // Tag{lang: _sw}                //  sw
	Tamil                Tag = Raw.MustParse("ta")      // Tag{lang: _ta}                //  ta
	Telugu               Tag = Raw.MustParse("te")      // Tag{lang: _te}                //  te
	Thai                 Tag = Raw.MustParse("th")      // Tag{lang: _th}                //  th
	Turkish              Tag = Raw.MustParse("tr")      // Tag{lang: _tr}                //  tr
	Ukrainian            Tag = Raw.MustParse("uk")      // Tag{lang: _uk}                //  uk
	Urdu                 Tag = Raw.MustParse("ur")      // Tag{lang: _ur}                //  ur
	Uzbek                Tag = Raw.MustParse("uz")      // Tag{lang: _uz}                //  uz
	Vietnamese           Tag = Raw.MustParse("vi")      // Tag{lang: _vi}                //  vi
	Chinese              Tag = Raw.MustParse("zh")      // Tag{lang: _zh}                //  zh
	SimplifiedChinese    Tag = Raw.MustParse("zh-Hans") // Tag{lang: _zh, script: _Hans} //  zh-Hans
	TraditionalChinese   Tag = Raw.MustParse("zh-Hant") // Tag{lang: _zh, script: _Hant} //  zh-Hant
	Zulu                 Tag = Raw.MustParse("zu")      // Tag{lang: _zu}                //  zu
)
