package generator

import (
	"context"
	"strings"
	"testing"
)

func TestNowGenerator(t *testing.T) {
	g := &NowGenerator{}
	ctx := context.Background()
	row := testRow{}

	v, err := g.Generate(ctx, row)
	if err != nil {
		t.Fatal(err)
	}
	ts := v.(string)
	if !strings.HasSuffix(ts, "Z") {
		t.Errorf("expected Z suffix, got %q", ts)
	}
	if !strings.Contains(ts, "T") {
		t.Errorf("expected T separator, got %q", ts)
	}
}

func TestTimeAgoGenerator(t *testing.T) {
	g := &TimeAgoGenerator{MaxDays: 30}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 50; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		ts := v.(string)
		if !strings.HasSuffix(ts, "Z") {
			t.Errorf("expected Z suffix, got %q", ts)
		}
		if !strings.Contains(ts, "T") {
			t.Errorf("expected T separator, got %q", ts)
		}
	}
}

func TestBusinessDaysGenerator(t *testing.T) {
	g := &BusinessDaysGenerator{MinYear: 2024, MaxYear: 2026}
	ctx := context.Background()
	row := testRow{}

	for i := 0; i < 100; i++ {
		v, err := g.Generate(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		date := v.(string)
		if !strings.Contains(date, "2024") && !strings.Contains(date, "2025") && !strings.Contains(date, "2026") {
			t.Errorf("date %q outside expected year range", date)
		}
		if !strings.Contains(date, "-") {
			t.Errorf("expected date format YYYY-MM-DD, got %q", date)
		}
	}
}
