package domain

import "errors"

var ErrInvalidResponseTransition = errors.New("invalid response transition")

type ResponseStatus string

const (
	ResponseDraft         ResponseStatus = "draft"
	ResponseSubmitted     ResponseStatus = "submitted"
	ResponseReviewed      ResponseStatus = "reviewed"
	ResponseNeedsMoreInfo ResponseStatus = "needs_more_info"
)

func CanTransitionResponse(from, to ResponseStatus) bool {
	if (from == ResponseDraft || from == ResponseNeedsMoreInfo) && to == ResponseSubmitted {
		return true
	}
	return from == ResponseSubmitted && (to == ResponseReviewed || to == ResponseNeedsMoreInfo)
}
