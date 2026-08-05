package domain

type CoverageLevel string

const (
	CoverageNone CoverageLevel = "none"
	CoveragePartial CoverageLevel = "partial"
	CoverageSubstantial CoverageLevel = "substantial"
	CoverageFull CoverageLevel = "full"
)

type ProfileScore struct {
	Included bool
	Current CoverageLevel
	Target CoverageLevel
}

type Summary struct {
	CoveragePct float64 `json:"coveragePct"`
	IncludedCount int `json:"includedCount"`
	PendingCount int `json:"pendingCount"`
	RejectedCount int `json:"rejectedCount"`
}
