package domain_test

import (
	"testing"

	"compliance/api/internal/domain"
)

func TestCanTransitionResponse(t *testing.T) {
	cases := []struct {
		from domain.ResponseStatus
		to   domain.ResponseStatus
		want bool
	}{
		{domain.ResponseDraft, domain.ResponseSubmitted, true},
		{domain.ResponseNeedsMoreInfo, domain.ResponseSubmitted, true},
		{domain.ResponseSubmitted, domain.ResponseReviewed, true},
		{domain.ResponseSubmitted, domain.ResponseNeedsMoreInfo, true},
		{domain.ResponseDraft, domain.ResponseReviewed, false},
		{domain.ResponseReviewed, domain.ResponseSubmitted, false},
		{domain.ResponseSubmitted, domain.ResponseDraft, false},
	}

	for _, testCase := range cases {
		if got := domain.CanTransitionResponse(testCase.from, testCase.to); got != testCase.want {
			t.Errorf("%s -> %s: got %v, want %v", testCase.from, testCase.to, got, testCase.want)
		}
	}
}

func TestResponseEditability(t *testing.T) {
	cases := []struct {
		status domain.ResponseStatus
		want   bool
	}{
		{domain.ResponseDraft, true},
		{domain.ResponseNeedsMoreInfo, true},
		{domain.ResponseSubmitted, false},
		{domain.ResponseReviewed, false},
	}

	for _, testCase := range cases {
		if got := domain.CanEditResponse(testCase.status); got != testCase.want {
			t.Errorf("%s editable: got %v, want %v", testCase.status, got, testCase.want)
		}
	}
}
