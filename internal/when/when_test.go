package when

import (
	"testing"
	"time"
)

func TestParseDeterministicDates(t *testing.T) {
	base := time.Date(2026, 5, 8, 20, 16, 0, 0, saoPaulo())

	tests := []struct {
		input   string
		from    string
		to      string
		fromNow bool
	}{
		{"today", "2026-05-08", "2026-05-08", false},
		{"tomorrow", "2026-05-09", "2026-05-09", false},
		{"from-now", "2026-05-08", "2026-05-08", true},
		{"2026-05-10", "2026-05-10", "2026-05-10", false},
		{"today..tomorrow", "2026-05-08", "2026-05-09", false},
		{"from tomorrow to sunday", "2026-05-09", "2026-05-10", false},
		{"weekend", "2026-05-09", "2026-05-10", false},
		{"next-weekend", "2026-05-16", "2026-05-17", false},
	}

	for _, tt := range tests {
		got, err := Parse(tt.input, base)
		if err != nil {
			t.Fatalf("Parse(%q): %v", tt.input, err)
		}
		if got.From != tt.from || got.To != tt.to || got.FromNow != tt.fromNow {
			t.Fatalf("Parse(%q) = %#v, want %s..%s fromNow=%v", tt.input, got, tt.from, tt.to, tt.fromNow)
		}
	}
}

func TestParseNaturalDateFallbackUsesFutureDirection(t *testing.T) {
	base := time.Date(2026, 5, 8, 20, 16, 0, 0, saoPaulo()) // Friday

	got, err := Parse("next friday", base)
	if err != nil {
		t.Fatal(err)
	}
	if got.From != "2026-05-15" || got.To != "2026-05-15" {
		t.Fatalf("expected next future friday, got %#v", got)
	}
}

func TestParseRangeFromToWithNaturalDateFallback(t *testing.T) {
	base := time.Date(2026, 5, 8, 20, 16, 0, 0, saoPaulo())

	got, err := ParseRange("tomorrow", "sunday", base)
	if err != nil {
		t.Fatal(err)
	}
	if got.From != "2026-05-09" || got.To != "2026-05-10" {
		t.Fatalf("unexpected range: %#v", got)
	}
}

func TestParseInvalidWhen(t *testing.T) {
	_, err := Parse("not a date expression", time.Date(2026, 5, 8, 20, 16, 0, 0, saoPaulo()))
	if err == nil {
		t.Fatal("expected invalid input error")
	}
}

func TestParsePortugueseTomorrowAndEmptyDefaultsToToday(t *testing.T) {
	base := time.Date(2026, 12, 30, 8, 0, 0, 0, saoPaulo())
	got, err := Parse("amanhã", base)
	if err != nil || got.From != "2026-12-31" || got.To != "2026-12-31" {
		t.Fatalf("Parse(amanhã): %#v %v", got, err)
	}
	gotEmpty, err := Parse("   ", base)
	if err != nil || gotEmpty.From != "2026-12-30" {
		t.Fatalf("empty when: %#v %v", gotEmpty, err)
	}
}

func TestParseRangesAcrossYearBoundary(t *testing.T) {
	base := time.Date(2026, 12, 30, 12, 0, 0, 0, saoPaulo())
	gotDot, err := Parse("2026-12-31..2027-01-02", base)
	if err != nil || gotDot.From != "2026-12-31" || gotDot.To != "2027-01-02" {
		t.Fatalf("dot range: %#v %v", gotDot, err)
	}

	gotPhrase, err := Parse("from 2026-12-31 to 2027-01-03", base)
	if err != nil || gotPhrase.From != "2026-12-31" || gotPhrase.To != "2027-01-03" {
		t.Fatalf("from-to phrase: %#v %v", gotPhrase, err)
	}
}

func TestWeekendAnchoredOnWednesdayVsSunday(t *testing.T) {
	// Wednesday 2026-05-06: nearest weekend is Sat May 9
	wed := time.Date(2026, 5, 6, 18, 0, 0, 0, saoPaulo())
	gotWed, err := Parse("weekend", wed)
	if err != nil || gotWed.From != "2026-05-09" || gotWed.To != "2026-05-10" {
		t.Fatalf("wed weekend: %#v %v", gotWed, err)
	}

	// Sunday base: SDK chooses "next weekend" because daysUntilSaturday==0 path adds 7 unless next flag
	sun := time.Date(2026, 5, 10, 10, 0, 0, 0, saoPaulo())
	gotSun, err := Parse("weekend", sun)
	if err != nil {
		t.Fatal(err)
	}
	if gotSun.From != "2026-05-16" || gotSun.To != "2026-05-17" {
		t.Fatalf("sunday-this-week pushes forward: %#v", gotSun)
	}
}

func TestParseRangeTwoExplicitDatesViaParseRange(t *testing.T) {
	base := time.Date(2026, 8, 1, 15, 0, 0, 0, saoPaulo())
	got, err := ParseRange("2026-08-10", "2026-08-20", base)
	if err != nil || got.From != "2026-08-10" || got.To != "2026-08-20" || got.FromNow {
		t.Fatalf("%#v %v", got, err)
	}
}

func TestFromNowParsesAcrossMidnightDSTSafe(t *testing.T) {
	base := time.Date(2027, 2, 27, 22, 0, 0, 0, saoPaulo())
	got, err := Parse("from-now", base)
	if err != nil || !got.FromNow || got.From != "2027-02-27" {
		t.Fatalf("%#v %v", got, err)
	}
}
