package store

import (
	"context"
	"os"
	"testing"
)

func TestWriteAuditPersistsRequestMetadata(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	data, err := New(context.Background(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	projectID, organizationID, counselorID, _, _ := setupScopeSubmissionFixture(t, data)
	organizationPtr := organizationID
	projectPtr := projectID
	if err := data.WriteAudit(context.Background(), AuditEvent{
		ActorUserID:    counselorID,
		ActorRole:      "counselor",
		OrganizationID: &organizationPtr,
		ProjectID:      &projectPtr,
		Action:         "audit.metadata_test",
		EntityType:     "project",
		Result:         "success",
		RequestID:      "request-987",
		IPAddress:      "192.0.2.44",
		UserAgent:      "integration-test/1.0",
		Metadata:       map[string]any{"safe": true},
	}); err != nil {
		t.Fatal(err)
	}
	events, err := data.ListProjectAuditEvents(context.Background(), projectID)
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].ActorRole != "counselor" || events[0].Result != "success" || events[0].RequestID != "request-987" || events[0].IPAddress != "192.0.2.44" || events[0].UserAgent != "integration-test/1.0" {
		t.Fatalf("unexpected persisted audit metadata: %#v", events)
	}
}

func TestWriteAuditAllowsAnonymousActor(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	data, err := New(context.Background(), databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	if err := data.WriteAudit(context.Background(), AuditEvent{Action: "auth.login_failed", EntityType: "session", Result: "failure", RequestID: "request-anonymous", Metadata: map[string]any{"email": "unknown@example.com"}}); err != nil {
		t.Fatal(err)
	}
}
