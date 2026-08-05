package domain

import "fmt"

func Score(level CoverageLevel) (int, error) {
	switch level {
	case CoverageNone: return 0, nil
	case CoveragePartial: return 1, nil
	case CoverageSubstantial: return 2, nil
	case CoverageFull: return 3, nil
	default: return 0, fmt.Errorf("invalid coverage level: %q", level)
	}
}

func CalculateSummary(rows []ProfileScore) Summary {
	var total int
	var count int
	for _, row := range rows {
		if !row.Included { continue }
		current, err := Score(row.Current)
		if err != nil { continue }
		total += current
		count++
	}
	if count == 0 { return Summary{} }
	return Summary{CoveragePct: float64(total) / float64(count) / 3 * 100, IncludedCount: count}
}
