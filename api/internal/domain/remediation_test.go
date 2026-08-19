package domain

import "testing"

func TestHasCoverageGapUsesCoverageOrder(t *testing.T) {
	tests := []struct {
		current CoverageLevel
		target  CoverageLevel
		want    bool
	}{
		{CoverageNone, CoveragePartial, true},
		{CoveragePartial, CoverageSubstantial, true},
		{CoverageSubstantial, CoverageFull, true},
		{CoverageFull, CoverageFull, false},
		{CoverageFull, CoveragePartial, false},
	}

	for _, test := range tests {
		if got := HasCoverageGap(test.current, test.target); got != test.want {
			t.Fatalf("HasCoverageGap(%q, %q) = %v; want %v", test.current, test.target, got, test.want)
		}
	}
}

func TestCanTransitionRemediation(t *testing.T) {
	allowed := [][2]RemediationStatus{
		{RemediationOpen, RemediationInProgress},
		{RemediationInProgress, RemediationAwaitingReview},
		{RemediationAwaitingReview, RemediationInProgress},
		{RemediationAwaitingReview, RemediationClosed},
	}
	for _, transition := range allowed {
		if !CanTransitionRemediation(transition[0], transition[1]) {
			t.Fatalf("expected transition %q -> %q to be allowed", transition[0], transition[1])
		}
	}

	rejected := [][2]RemediationStatus{
		{RemediationOpen, RemediationAwaitingReview},
		{RemediationInProgress, RemediationClosed},
		{RemediationClosed, RemediationInProgress},
	}
	for _, transition := range rejected {
		if CanTransitionRemediation(transition[0], transition[1]) {
			t.Fatalf("expected transition %q -> %q to be rejected", transition[0], transition[1])
		}
	}
}

func TestValidRemediationPriority(t *testing.T) {
	for _, priority := range []RemediationPriority{PriorityLow, PriorityMedium, PriorityHigh, PriorityCritical} {
		if !ValidRemediationPriority(priority) {
			t.Fatalf("expected priority %q to be valid", priority)
		}
	}
	if ValidRemediationPriority(RemediationPriority("urgent")) {
		t.Fatal("expected unknown priority to be invalid")
	}
}
