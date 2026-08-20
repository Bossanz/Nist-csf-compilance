package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"compliance/api/internal/store"
)

var (
	ErrForbidden            = errors.New("forbidden")
	ErrDuplicateInvitation  = errors.New("duplicate invitation")
	ErrInvalidInvitation    = errors.New("invalid invitation")
	ErrInvalidProjectAccess = errors.New("auditor invitation requires project access")
	ErrInvitationNotPending = errors.New("invitation is not pending")
	ErrWeakPassword         = errors.New("password must be at least 8 characters")
)

type InvitationRepository interface {
	HasActiveOrPendingEmail(context.Context, string) (bool, error)
	CreateInvitation(context.Context, store.Invitation) (store.Invitation, error)
	AcceptInvitation(context.Context, string, string, string, time.Time) (store.User, error)
}

type invitationLifecycleRepository interface {
	ResendInvitation(context.Context, string, string, string, string, time.Time, time.Time) (store.Invitation, error)
	CancelInvitation(context.Context, string, string, string, time.Time) (store.Invitation, error)
}

type InvitationService struct {
	repository InvitationRepository
	now        func() time.Time
}

func NewInvitationService(repository InvitationRepository, now func() time.Time) *InvitationService {
	return &InvitationService{repository: repository, now: now}
}

func isCounselorRole(role string) bool { return role == "counselor_admin" || role == "counselor" }
func isStakeholderRole(role string) bool {
	return role == "org_admin" || role == "assessor" || role == "reviewer" || role == "viewer" || role == "auditor"
}

func (s *InvitationService) Invite(ctx context.Context, actor store.User, organizationID *string, email, role string, projectIDs ...string) (store.Invitation, string, error) {
	email = NormalizeEmail(email)
	userType := "stakeholder"
	if isCounselorRole(role) {
		userType = "counselor"
		if actor.Role != "counselor_admin" || organizationID != nil {
			return store.Invitation{}, "", ErrForbidden
		}
	} else if isStakeholderRole(role) {
		if organizationID == nil {
			return store.Invitation{}, "", ErrForbidden
		}
		if role == "auditor" && (actor.UserType != "stakeholder" || actor.Role != "org_admin" || actor.OrganizationID == nil || *actor.OrganizationID != *organizationID) {
			return store.Invitation{}, "", ErrForbidden
		}
		if role == "auditor" && len(projectIDs) == 0 {
			return store.Invitation{}, "", ErrInvalidProjectAccess
		}
		if actor.UserType == "stakeholder" && (actor.Role != "org_admin" || actor.OrganizationID == nil || *actor.OrganizationID != *organizationID) {
			return store.Invitation{}, "", ErrForbidden
		}
		if actor.UserType == "counselor" && !isCounselorRole(actor.Role) {
			return store.Invitation{}, "", ErrForbidden
		}
	} else {
		return store.Invitation{}, "", ErrForbidden
	}
	duplicate, err := s.repository.HasActiveOrPendingEmail(ctx, email)
	if err != nil {
		return store.Invitation{}, "", err
	}
	if duplicate {
		return store.Invitation{}, "", ErrDuplicateInvitation
	}
	raw, hash, err := NewToken()
	if err != nil {
		return store.Invitation{}, "", err
	}
	invitation := store.Invitation{OrganizationID: organizationID, Email: email, UserType: userType, Role: role, TokenHash: hash, InvitedBy: actor.ID, ExpiresAt: s.now().Add(72 * time.Hour), ProjectIDs: append([]string(nil), projectIDs...)}
	created, err := s.repository.CreateInvitation(ctx, invitation)
	if errors.Is(err, store.ErrInvalidProjectAccess) {
		return store.Invitation{}, "", ErrInvalidProjectAccess
	}
	if err != nil {
		return store.Invitation{}, "", err
	}
	return created, raw, nil
}

func (s *InvitationService) Resend(ctx context.Context, actor store.User, organizationID, invitationID string) (store.Invitation, string, error) {
	if actor.UserType != "stakeholder" || actor.Role != "org_admin" || actor.OrganizationID == nil || *actor.OrganizationID != organizationID {
		return store.Invitation{}, "", ErrForbidden
	}
	repository, ok := s.repository.(invitationLifecycleRepository)
	if !ok {
		return store.Invitation{}, "", errors.New("invitation lifecycle repository unavailable")
	}
	raw, hash, err := NewToken()
	if err != nil {
		return store.Invitation{}, "", err
	}
	invitation, err := repository.ResendInvitation(ctx, organizationID, invitationID, hash, actor.ID, s.now().Add(72*time.Hour), s.now())
	if errors.Is(err, store.ErrInvitationNotPending) {
		return store.Invitation{}, "", ErrInvitationNotPending
	}
	if err != nil {
		return store.Invitation{}, "", err
	}
	return invitation, raw, nil
}

func (s *InvitationService) Cancel(ctx context.Context, actor store.User, organizationID, invitationID string) (store.Invitation, error) {
	if actor.UserType != "stakeholder" || actor.Role != "org_admin" || actor.OrganizationID == nil || *actor.OrganizationID != organizationID {
		return store.Invitation{}, ErrForbidden
	}
	repository, ok := s.repository.(invitationLifecycleRepository)
	if !ok {
		return store.Invitation{}, errors.New("invitation lifecycle repository unavailable")
	}
	invitation, err := repository.CancelInvitation(ctx, organizationID, invitationID, actor.ID, s.now())
	if errors.Is(err, store.ErrInvitationNotPending) {
		return store.Invitation{}, ErrInvitationNotPending
	}
	if err != nil {
		return store.Invitation{}, err
	}
	return invitation, nil
}

func (s *InvitationService) Accept(ctx context.Context, rawToken, name, password string) (store.User, error) {
	name = strings.TrimSpace(name)
	if name == "" || len(password) < 8 {
		return store.User{}, ErrWeakPassword
	}
	hash, err := HashPassword(password)
	if err != nil {
		return store.User{}, err
	}
	user, err := s.repository.AcceptInvitation(ctx, HashToken(rawToken), name, hash, s.now())
	if errors.Is(err, store.ErrInvalidInvitation) || errors.Is(err, store.ErrInvalidProjectAccess) {
		return store.User{}, ErrInvalidInvitation
	}
	return user, err
}
