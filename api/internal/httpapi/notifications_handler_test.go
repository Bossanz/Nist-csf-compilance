package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"compliance/api/internal/store"
)

type fakeNotificationStore struct {
	fakeStore
	reviewerEmails     []string
	assessorEmail      string
	organizationEmails []string
}

func (f *fakeNotificationStore) ListProjectReviewerEmails(context.Context, string) ([]string, error) {
	return f.reviewerEmails, nil
}

func (f *fakeNotificationStore) GetAssignedAssessorEmail(context.Context, string, string) (string, error) {
	return f.assessorEmail, nil
}

func (f *fakeNotificationStore) ListOrganizationEmailsByRoles(context.Context, string, []string) ([]string, error) {
	return f.organizationEmails, nil
}

func TestInvitationCreationSendsInviteEmail(t *testing.T) {
	sender := &fakeEmailSender{}
	organizationID := "org-1"
	handler := &Handler{Store: &fakeNotificationStore{}, EmailSender: sender, AppOrigin: "http://localhost:3000"}
	invitation := store.Invitation{ID: "inv-1", OrganizationID: &organizationID, Email: "stakeholder@example.com", Role: "assessor"}
	response := httptest.NewRecorder()
	handler.writeInvitationResult(response, httptest.NewRequest(http.MethodPost, "/", nil), invitation, "invite-token", nil)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected invitation response, got %d %s", response.Code, response.Body.String())
	}
	if len(sender.messages) != 1 || sender.messages[0].To != invitation.Email || sender.messages[0].Subject == "" || !containsAll(sender.messages[0].Text, "invite-token", "assessor") {
		t.Fatalf("unexpected invitation email: %#v", sender.messages)
	}
}

func TestSubmittedResponseNotifiesProjectReviewers(t *testing.T) {
	sender := &fakeEmailSender{}
	handler := &Handler{
		Store:       &fakeNotificationStore{reviewerEmails: []string{"reviewer-one@example.com", "reviewer-two@example.com"}},
		EmailSender: sender,
		AppOrigin:   "http://localhost:3000",
	}
	handler.notifyResponseSubmitted(context.Background(), "project-1", "subcategory-1")
	if len(sender.messages) != 2 {
		t.Fatalf("expected two reviewer emails, got %#v", sender.messages)
	}
	for _, message := range sender.messages {
		if message.Subject != "Response ready for review" || !containsAll(message.Text, "project-1", "subcategory-1") {
			t.Fatalf("unexpected reviewer email: %#v", message)
		}
	}
}

func TestReviewedResponseNotifiesAssignedAssessor(t *testing.T) {
	sender := &fakeEmailSender{}
	handler := &Handler{
		Store:       &fakeNotificationStore{assessorEmail: "assessor@example.com"},
		EmailSender: sender,
	}
	handler.notifyResponseReviewed(context.Background(), "project-1", "subcategory-1", "needs_more_info", "Please add the access review evidence.")
	if len(sender.messages) != 1 || sender.messages[0].To != "assessor@example.com" || !containsAll(sender.messages[0].Text, "needs_more_info", "Please add the access review evidence.") {
		t.Fatalf("unexpected assessor email: %#v", sender.messages)
	}
}

func TestFinalizedProjectNotifiesOrganizationReviewersAndAdmins(t *testing.T) {
	sender := &fakeEmailSender{}
	organizationID := "org-1"
	handler := &Handler{
		Store:       &fakeNotificationStore{organizationEmails: []string{"admin@example.com", "reviewer@example.com"}},
		EmailSender: sender,
	}
	handler.notifyProjectFinalized(context.Background(), store.Project{ID: "project-1", OrganizationID: organizationID, Name: "RU Registration"})
	if len(sender.messages) != 2 {
		t.Fatalf("expected two finalization emails, got %#v", sender.messages)
	}
	for _, message := range sender.messages {
		if message.Subject != "Project finalized" || !containsAll(message.Text, "RU Registration", "project-1") {
			t.Fatalf("unexpected finalization email: %#v", message)
		}
	}
}

func containsAll(value string, expected ...string) bool {
	for _, part := range expected {
		if !strings.Contains(value, part) {
			return false
		}
	}
	return true
}
