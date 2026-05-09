package when

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	naturaldate "github.com/anatol/naturaldate.go"
)

type Filter struct {
	From    string
	To      string
	FromNow bool
}

var fromTo = regexp.MustCompile(`(?i)^from\s+(.+?)\s+to\s+(.+)$`)

func Parse(input string, base time.Time) (Filter, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		input = "today"
	}

	lower := strings.ToLower(input)
	switch lower {
	case "today":
		return single(base), nil
	case "tomorrow", "amanha", "amanhã":
		return single(base.AddDate(0, 0, 1)), nil
	case "from-now", "from now", "now":
		f := single(base)
		f.FromNow = true
		return f, nil
	case "weekend":
		return weekend(base, false), nil
	case "next-weekend", "next weekend":
		return weekend(base, true), nil
	}

	if strings.Contains(input, "..") {
		parts := strings.SplitN(input, "..", 2)
		return ParseRange(parts[0], parts[1], base)
	}
	if matches := fromTo.FindStringSubmatch(input); len(matches) == 3 {
		return ParseRange(matches[1], matches[2], base)
	}
	if t, err := time.ParseInLocation(time.DateOnly, input, saoPaulo()); err == nil {
		return single(t), nil
	}
	if !looksNaturalDate(input) {
		return Filter{}, fmt.Errorf("invalid when %q", input)
	}

	t, err := naturaldate.Parse(input, base, naturaldate.WithDirection(naturaldate.Future))
	if err != nil {
		return Filter{}, fmt.Errorf("invalid when %q: %w", input, err)
	}
	if t.Equal(base) {
		return Filter{}, fmt.Errorf("invalid when %q", input)
	}
	return single(t.In(saoPaulo())), nil
}

func looksNaturalDate(input string) bool {
	input = strings.ToLower(input)
	tokens := []string{
		"monday", "tuesday", "wednesday", "thursday", "friday", "saturday", "sunday",
		"mon", "tue", "wed", "thu", "fri", "sat", "sun",
		"january", "february", "march", "april", "june", "july", "august", "september", "october", "november", "december",
		"jan", "feb", "mar", "apr", "jun", "jul", "aug", "sep", "oct", "nov", "dec",
		"next", "last", "ago", "minute", "hour", "day", "week", "month", "year", "am", "pm",
	}
	for _, r := range input {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	for _, token := range tokens {
		if strings.Contains(input, token) {
			return true
		}
	}
	return false
}

func ParseRange(fromExpr, toExpr string, base time.Time) (Filter, error) {
	from, err := Parse(fromExpr, base)
	if err != nil {
		return Filter{}, err
	}
	to, err := Parse(toExpr, base)
	if err != nil {
		return Filter{}, err
	}
	return Filter{From: from.From, To: to.To, FromNow: from.FromNow}, nil
}

func single(t time.Time) Filter {
	date := t.In(saoPaulo()).Format(time.DateOnly)
	return Filter{From: date, To: date}
}

func weekend(base time.Time, next bool) Filter {
	daysUntilSaturday := (int(time.Saturday) - int(base.Weekday()) + 7) % 7
	if daysUntilSaturday == 0 || next {
		daysUntilSaturday += 7
	}
	saturday := base.AddDate(0, 0, daysUntilSaturday)
	sunday := saturday.AddDate(0, 0, 1)
	return Filter{From: saturday.In(saoPaulo()).Format(time.DateOnly), To: sunday.In(saoPaulo()).Format(time.DateOnly)}
}

func saoPaulo() *time.Location {
	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return time.FixedZone("America/Sao_Paulo", -3*60*60)
	}
	return loc
}
