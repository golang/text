package bidi

import (
	"log"
	"testing"
)

type runInformation struct {
	str   string
	dir   Direction
	start int
	end   int
}

func TestSimple(t *testing.T) {
	str := "Hellö"
	p := Paragraph{}
	p.SetString(str)
	order, err := p.Order()
	if err != nil {
		log.Fatal(err)
	}
	expectedRuns := []runInformation{
		{"Hellö", LeftToRight, 0, 4},
	}

	if !p.IsLeftToRight() {
		t.Error("p.IsLeftToRight() == false; want true")
	}
	if nr, want := order.NumRuns(), len(expectedRuns); nr != want {
		t.Errorf("order.NumRuns() = %d; want %d", nr, want)
	}
	for i, want := range expectedRuns {
		r := order.Run(i)
		if got := r.String(); got != want.str {
			t.Errorf("Run(%d) = %q; want %q", i, got, want.str)
		}
		if s, e := r.Pos(); s != want.start || e != want.end {
			t.Errorf("Run(%d).start = %d, .end = %d; want start %d, end %d", i, s, e, want.start, want.end)
		}
		if d := r.Direction(); d != want.dir {
			t.Errorf("Run(%d).Direction = %d; want %d", i, d, want.dir)
		}
	}
}

func TestMixed(t *testing.T) {
	str := `العاشر ليونيكود (Unicode Conference)، الذي سيعقد في 10-12 آذار 1997 مبدينة`
	p := Paragraph{}
	p.SetString(str)
	order, err := p.Order()
	if err != nil {
		log.Fatal(err)
	}
	if p.IsLeftToRight() {
		t.Error("p.IsLeftToRight() == true; want false")
	}

	expectedRuns := []runInformation{
		{"العاشر ليونيكود (", RightToLeft, 0, 16},
		{"Unicode Conference", LeftToRight, 17, 34},
		{")، الذي سيعقد في ", RightToLeft, 35, 51},
		{"10", LeftToRight, 52, 53},
		{"-", RightToLeft, 54, 54},
		{"12", LeftToRight, 55, 56},
		{" آذار ", RightToLeft, 57, 62},
		{"1997", LeftToRight, 63, 66},
		{" مبدينة", RightToLeft, 67, 73},
	}

	if nr, want := order.NumRuns(), len(expectedRuns); nr != want {
		t.Errorf("order.NumRuns() = %d; want %d", nr, want)
	}

	for i, want := range expectedRuns {
		r := order.Run(i)
		if got := r.String(); got != want.str {
			t.Errorf("Run(%d) = %q; want %q", i, got, want.str)
		}
		if s, e := r.Pos(); s != want.start || e != want.end {
			t.Errorf("Run(%d).start = %d, .end = %d; want start = %d, end = %d", i, s, e, want.start, want.end)
		}
		if d := r.Direction(); d != want.dir {
			t.Errorf("Run(%d).Direction = %d; want %d", i, d, want.dir)
		}
	}
}

func TestExplicitIsolate(t *testing.T) {
	// https://www.w3.org/International/articles/inline-bidi-markup/uba-basics.en#beyond
	str := "The names of these states in Arabic are \u2067مصر\u2069, \u2067البحرين\u2069 and \u2067الكويت\u2069 respectively."
	p := Paragraph{}
	p.SetString(str)
	order, err := p.Order()
	if err != nil {
		log.Fatal(err)
	}
	if !p.IsLeftToRight() {
		t.Error("p.IsLeftToRight() == false; want true")
	}

	expectedRuns := []runInformation{
		{"The names of these states in Arabic are \u2067", LeftToRight, 0, 40},
		{"مصر", RightToLeft, 41, 43},
		{"\u2069, \u2067", LeftToRight, 44, 47},
		{"البحرين", RightToLeft, 48, 54},
		{"\u2069 and \u2067", LeftToRight, 55, 61},
		{"الكويت", RightToLeft, 62, 67},
		{"\u2069 respectively.", LeftToRight, 68, 82},
	}

	if nr, want := order.NumRuns(), len(expectedRuns); nr != want {
		t.Errorf("order.NumRuns() = %d; want %d", nr, want)
	}

	for i, want := range expectedRuns {
		r := order.Run(i)
		if got := r.String(); got != want.str {
			t.Errorf("Run(%d) = %q; want %q", i, got, want.str)
		}
		if s, e := r.Pos(); s != want.start || e != want.end {
			t.Errorf("Run(%d).start = %d, .end = %d; want start = %d, end = %d", i, s, e, want.start, want.end)
		}
		if d := r.Direction(); d != want.dir {
			t.Errorf("Run(%d).Direction = %d; want %d", i, d, want.dir)
		}
	}
}

func TestWithoutExplicitIsolate(t *testing.T) {
	str := "The names of these states in Arabic are مصر, البحرين and الكويت respectively."
	p := Paragraph{}
	p.SetString(str)
	order, err := p.Order()
	if err != nil {
		log.Fatal(err)
	}
	if !p.IsLeftToRight() {
		t.Error("p.IsLeftToRight() == false; want true")
	}

	expectedRuns := []runInformation{
		{"The names of these states in Arabic are ", LeftToRight, 0, 39},
		{"مصر, البحرين", RightToLeft, 40, 51},
		{" and ", LeftToRight, 52, 56},
		{"الكويت", RightToLeft, 57, 62},
		{" respectively.", LeftToRight, 63, 76},
	}

	if nr, want := order.NumRuns(), len(expectedRuns); nr != want {
		t.Errorf("order.NumRuns() = %d; want %d", nr, want)
	}

	for i, want := range expectedRuns {
		r := order.Run(i)
		if got := r.String(); got != want.str {
			t.Errorf("Run(%d) = %q; want %q", i, got, want.str)
		}
		if s, e := r.Pos(); s != want.start || e != want.end {
			t.Errorf("Run(%d).start = %d, .end = %d; want start = %d, end = %d", i, s, e, want.start, want.end)
		}
		if d := r.Direction(); d != want.dir {
			t.Errorf("Run(%d).Direction = %d; want %d", i, d, want.dir)
		}
	}
}

func TestLongUTF8(t *testing.T) {
	str := `𠀀`
	p := Paragraph{}
	p.SetString(str)
	order, err := p.Order()
	if err != nil {
		log.Fatal(err)
	}
	if !p.IsLeftToRight() {
		t.Error("p.IsLeftToRight() == false; want true")
	}

	expectedRuns := []runInformation{
		{"𠀀", LeftToRight, 0, 0},
	}

	if nr, want := order.NumRuns(), len(expectedRuns); nr != want {
		t.Errorf("order.NumRuns() = %d; want %d", nr, want)
	}

	for i, want := range expectedRuns {
		r := order.Run(i)
		if got := r.String(); got != want.str {
			t.Errorf("Run(%d) = %q; want %q", i, got, want.str)
		}
		if s, e := r.Pos(); s != want.start || e != want.end {
			t.Errorf("Run(%d).start = %d, .end = %d; want start = %d, end = %d", i, s, e, want.start, want.end)
		}
		if d := r.Direction(); d != want.dir {
			t.Errorf("Run(%d).Direction = %d; want %d", i, d, want.dir)
		}
	}
}

func TestLLongUTF8(t *testing.T) {
	strTester := []struct {
		str string
		l   int
	}{
		{"ö", 2},
		{"ॡ", 3},
		{`𠀀`, 4},
	}
	for _, st := range strTester {
		str := st.str
		want := st.l
		if _, l := LookupString(str); l != want {
			t.Errorf("LookupString(%q) length = %d; want %d", str, l, want)
		}

	}

}

func TestMixedSimple(t *testing.T) {
	str := `Uا`
	p := Paragraph{}
	p.SetString(str)
	order, err := p.Order()
	if err != nil {
		log.Fatal(err)
	}
	if !p.IsLeftToRight() {
		t.Error("p.IsLeftToRight() == false; want true")
	}

	expectedRuns := []runInformation{
		{"U", LeftToRight, 0, 0},
		{"ا", RightToLeft, 1, 1},
	}

	if nr, want := order.NumRuns(), len(expectedRuns); nr != want {
		t.Errorf("order.NumRuns() = %d; want %d", nr, want)
	}

	for i, want := range expectedRuns {
		r := order.Run(i)
		if got := r.String(); got != want.str {
			t.Errorf("Run(%d) = %q; want %q", i, got, want.str)
		}
		if s, e := r.Pos(); s != want.start || e != want.end {
			t.Errorf("Run(%d).start = %d, .end = %d; want start = %d, end = %d", i, s, e, want.start, want.end)
		}
		if d := r.Direction(); d != want.dir {
			t.Errorf("Run(%d).Direction = %d; want %d", i, d, want.dir)
		}
	}
}

func TestDefaultDirection(t *testing.T) {
	str := "+"
	p := Paragraph{}
	p.SetString(str, DefaultDirection(RightToLeft))
	_, err := p.Order()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if want, dir := false, p.IsLeftToRight(); want != dir {
		t.Errorf("p.IsLeftToRight() = %t; want %t", dir, want)
	}
	p.SetString(str, DefaultDirection(LeftToRight))
	_, err = p.Order()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	if want, dir := true, p.IsLeftToRight(); want != dir {
		t.Errorf("p.IsLeftToRight() = %t; want %t", dir, want)
	}

}

func TestEmpty(t *testing.T) {
	p := Paragraph{}
	p.SetBytes([]byte{})
	o, err := p.Order()
	if err != nil {
		t.Error("p.Order() return err != nil; want err == nil")
	}
	if nr := o.NumRuns(); nr != 0 {
		t.Errorf("o.NumRuns() = %d; want 0", nr)
	}
}

func TestNewline(t *testing.T) {
	str := "Hello\nworld"
	p := Paragraph{}
	n, err := p.SetString(str)
	if err != nil {
		t.Error(err)
	}
	// 6 is the length up to and including the \n
	if want := 6; n != want {
		t.Errorf("SetString(%q) = nil, %d; want nil, %d", str, n, want)
	}
}

func TestDoubleSetString(t *testing.T) {
	str := "العاشر ليونيكود (Unicode Conference)،"
	p := Paragraph{}
	_, err := p.SetString(str)
	if err != nil {
		t.Error(err)
	}
	_, err = p.SetString(str)
	if err != nil {
		t.Error(err)
	}
	_, err = p.Order()
	if err != nil {
		t.Error(err)
	}
}

func TestReverseString(t *testing.T) {
	testcase := []struct {
		input string
		want  string
	}{
		{"(Hello)", "(olleH)"},
		{"nice (world) placé́́", "é́́calp (dlrow) ecin"},
		{"äu", "uä"},
		{"uä", "äu"},
		{"é́́abé́́é́́", "é́́é́́baé́́"},
		{"é́́é́́baé́́", "é́́abé́́é́́"},
		{"✍🏾✍🏾ab✍🏾✍🏾", "✍🏾✍🏾ba✍🏾✍🏾"},
	}
	for _, tc := range testcase {
		if str := ReverseString(tc.input); str != tc.want {
			t.Errorf("ReverseString(%s) = %q; want %q", tc.input, str, tc.want)
		}
	}
}

func TestAppendReverse(t *testing.T) {
	testcase := []struct {
		inString  string
		outString string
		want      string
	}{
		{"", "Hëllo", "Hëllo"},
		{"nice (wörld) placé́́", "", "é́́calp (dlröw) ecin"},
		{"nice (wörld) placé́́", "Hëllo", "Hëlloé́́calp (dlröw) ecin"},
		{"✍🏾✍🏾ab✍🏾✍🏾", "✍🏾✍🏾ba✍🏾✍🏾", "✍🏾✍🏾ba✍🏾✍🏾✍🏾✍🏾ba✍🏾✍🏾"},
	}
	for _, tc := range testcase {
		if r := AppendReverse([]byte(tc.outString), []byte(tc.inString)); string(r) != tc.want {
			t.Errorf("AppendReverse([]byte(%q), []byte(%q) = %q; want %q", tc.outString, tc.inString, string(r), tc.want)
		}
	}

}
