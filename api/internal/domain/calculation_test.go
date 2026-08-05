package domain

import "testing"

func TestScoreMapsCoverageLevels(t *testing.T) {
	cases := []struct { level CoverageLevel; want int }{
		{CoverageNone, 0}, {CoveragePartial, 1}, {CoverageSubstantial, 2}, {CoverageFull, 3},
	}
	for _, tc := range cases {
		got, err := Score(tc.level)
		if err != nil || got != tc.want { t.Fatalf("Score(%q) = %d, %v; want %d", tc.level, got, err, tc.want) }
	}
}

func TestScoreRejectsUnknownLevel(t *testing.T) {
	if _, err := Score(CoverageLevel("invalid")); err == nil { t.Fatal("expected invalid coverage error") }
}

func TestCalculateSummaryExcludesOutOfScopeRows(t *testing.T) {
	got := CalculateSummary([]ProfileScore{
		{Included: true, Current: CoverageFull, Target: CoverageFull},
		{Included: true, Current: CoveragePartial, Target: CoverageFull},
		{Included: false, Current: CoverageNone, Target: CoverageFull},
	})
	if got.CoveragePct != 66.66666666666666 || got.IncludedCount != 2 { t.Fatalf("unexpected summary: %+v", got) }
}

func TestCalculateSummaryReturnsZeroWhenNothingIncluded(t *testing.T) {
	got := CalculateSummary([]ProfileScore{{Included: false, Current: CoverageFull}})
	if got.CoveragePct != 0 || got.IncludedCount != 0 { t.Fatalf("unexpected summary: %+v", got) }
}
