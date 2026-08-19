package domain

type RemediationStatus string

const (
	RemediationOpen           RemediationStatus = "open"
	RemediationInProgress     RemediationStatus = "in_progress"
	RemediationAwaitingReview RemediationStatus = "awaiting_review"
	RemediationClosed         RemediationStatus = "closed"
)

type RemediationPriority string

const (
	PriorityLow      RemediationPriority = "low"
	PriorityMedium   RemediationPriority = "medium"
	PriorityHigh     RemediationPriority = "high"
	PriorityCritical RemediationPriority = "critical"
)

func HasCoverageGap(current, target CoverageLevel) bool {
	rank := map[CoverageLevel]int{
		CoverageNone:        0,
		CoveragePartial:     1,
		CoverageSubstantial: 2,
		CoverageFull:        3,
	}
	return rank[current] < rank[target]
}

func CanTransitionRemediation(from, to RemediationStatus) bool {
	return from == RemediationOpen && to == RemediationInProgress ||
		from == RemediationInProgress && to == RemediationAwaitingReview ||
		from == RemediationAwaitingReview && (to == RemediationInProgress || to == RemediationClosed)
}

func ValidRemediationPriority(priority RemediationPriority) bool {
	return priority == PriorityLow || priority == PriorityMedium || priority == PriorityHigh || priority == PriorityCritical
}
