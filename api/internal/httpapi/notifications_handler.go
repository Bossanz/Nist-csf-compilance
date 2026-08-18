package httpapi

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"

	"compliance/api/internal/notifications"
	"compliance/api/internal/store"
	"github.com/jackc/pgx/v5"
)

type notificationRecipientStore interface {
	ListProjectReviewerEmails(context.Context, string) ([]string, error)
	GetAssignedAssessorEmail(context.Context, string, string) (string, error)
	ListOrganizationEmailsByRoles(context.Context, string, []string) ([]string, error)
}

func (h *Handler) sendNotification(ctx context.Context, event string, recipients []string, subject, text string) {
	if h.EmailSender == nil {
		return
	}
	seen := make(map[string]struct{}, len(recipients))
	for _, recipient := range recipients {
		recipient = strings.TrimSpace(recipient)
		if recipient == "" {
			continue
		}
		key := strings.ToLower(recipient)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		if err := h.EmailSender.Send(ctx, notifications.EmailMessage{To: recipient, Subject: subject, Text: text}); err != nil {
			log.Printf("notification email failed event=%s to=%s: %v", event, recipient, err)
		}
	}
}

func (h *Handler) notifyInvitation(ctx context.Context, invitation store.Invitation, rawToken string) {
	if strings.TrimSpace(invitation.Email) == "" || strings.TrimSpace(rawToken) == "" {
		return
	}
	inviteURL := strings.TrimRight(h.AppOrigin, "/") + "/invite/" + url.PathEscape(rawToken)
	text := fmt.Sprintf("You have been invited to CSF Compliance as %s.\n\nCreate your account here:\n%s\n\nThis invitation link expires soon and can only be used once.", invitation.Role, inviteURL)
	h.sendNotification(ctx, "invitation.created", []string{invitation.Email}, "You're invited to CSF Compliance", text)
}

func (h *Handler) notifyResponseSubmitted(ctx context.Context, projectID, subcategoryID string) {
	recipients, err := h.notificationRecipients().ListProjectReviewerEmails(ctx, projectID)
	if err != nil {
		log.Printf("notification recipients lookup failed event=response.submitted project=%s: %v", projectID, err)
		return
	}
	text := fmt.Sprintf("A stakeholder response is ready for review.\n\nProject: %s\nOutcome: %s", projectID, subcategoryID)
	h.sendNotification(ctx, "response.submitted", recipients, "Response ready for review", text)
}

func (h *Handler) notifyResponseReviewed(ctx context.Context, projectID, subcategoryID, status, comment string) {
	recipient, err := h.notificationRecipients().GetAssignedAssessorEmail(ctx, projectID, subcategoryID)
	if errors.Is(err, pgx.ErrNoRows) {
		return
	}
	if err != nil {
		log.Printf("notification recipient lookup failed event=response.reviewed project=%s outcome=%s: %v", projectID, subcategoryID, err)
		return
	}
	text := fmt.Sprintf("Your stakeholder response was reviewed.\n\nProject: %s\nOutcome: %s\nStatus: %s\nReviewer comment: %s", projectID, subcategoryID, status, comment)
	h.sendNotification(ctx, "response.reviewed", []string{recipient}, "Your response was reviewed", text)
}

func (h *Handler) notifyProjectFinalized(ctx context.Context, project store.Project) {
	recipients, err := h.notificationRecipients().ListOrganizationEmailsByRoles(ctx, project.OrganizationID, []string{"org_admin", "reviewer"})
	if err != nil {
		log.Printf("notification recipients lookup failed event=project.finalized project=%s: %v", project.ID, err)
		return
	}
	text := fmt.Sprintf("The compliance project has been finalized.\n\nProject: %s\nProject ID: %s", project.Name, project.ID)
	h.sendNotification(ctx, "project.finalized", recipients, "Project finalized", text)
}

func (h *Handler) notificationRecipients() notificationRecipientStore {
	recipients, ok := h.Store.(notificationRecipientStore)
	if !ok {
		return unavailableNotificationRecipientStore{}
	}
	return recipients
}

type unavailableNotificationRecipientStore struct{}

func (unavailableNotificationRecipientStore) ListProjectReviewerEmails(context.Context, string) ([]string, error) {
	return nil, errors.New("notification recipient store unavailable")
}

func (unavailableNotificationRecipientStore) GetAssignedAssessorEmail(context.Context, string, string) (string, error) {
	return "", errors.New("notification recipient store unavailable")
}

func (unavailableNotificationRecipientStore) ListOrganizationEmailsByRoles(context.Context, string, []string) ([]string, error) {
	return nil, errors.New("notification recipient store unavailable")
}
