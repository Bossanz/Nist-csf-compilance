package notifications

import (
	"bytes"
	"context"
	"log"
	"net/smtp"
	"strings"
	"testing"
)

func TestLogEmailSenderWritesRecipientSubjectAndBody(t *testing.T) {
	var output bytes.Buffer
	sender := newLogEmailSender(log.New(&output, "", 0))
	err := sender.Send(context.Background(), EmailMessage{To: "person@example.com", Subject: "Test subject", Text: "Test body"})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"person@example.com", "Test subject", "Test body"} {
		if !strings.Contains(output.String(), expected) {
			t.Fatalf("expected log to contain %q, got %q", expected, output.String())
		}
	}
}

func TestSMTPEmailSenderBuildsPlainTextMessage(t *testing.T) {
	var gotAddress, gotFrom string
	var gotRecipients []string
	var gotMessage []byte
	sender := &smtpEmailSender{
		host: "smtp.example.com",
		port: "587",
		from: "no-reply@example.com",
		sendMail: func(address string, _ smtp.Auth, from string, recipients []string, message []byte) error {
			gotAddress, gotFrom, gotRecipients, gotMessage = address, from, recipients, message
			return nil
		},
	}

	err := sender.Send(context.Background(), EmailMessage{To: "person@example.com", Subject: "Subject", Text: "Plain body"})
	if err != nil {
		t.Fatal(err)
	}
	if gotAddress != "smtp.example.com:587" || gotFrom != "no-reply@example.com" || len(gotRecipients) != 1 || gotRecipients[0] != "person@example.com" {
		t.Fatalf("unexpected SMTP envelope: address=%q from=%q recipients=%#v", gotAddress, gotFrom, gotRecipients)
	}
	message := string(gotMessage)
	if !strings.Contains(message, "To: person@example.com") || !strings.Contains(message, "Subject: Subject") || !strings.HasSuffix(message, "\r\n\r\nPlain body") {
		t.Fatalf("unexpected plain-text message: %q", message)
	}
}

func TestLogModeIsSelectedByDefault(t *testing.T) {
	sender, err := NewFromEnv(log.New(&bytes.Buffer{}, "", 0), func(string) string { return "" })
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := sender.(*logEmailSender); !ok {
		t.Fatalf("expected log sender by default, got %T", sender)
	}
}

func TestSMTPModeRequiresHostAndFrom(t *testing.T) {
	sender, err := NewFromEnv(log.Default(), func(key string) string {
		if key == "EMAIL_MODE" {
			return "smtp"
		}
		return ""
	})
	if err == nil || sender != nil || !strings.Contains(err.Error(), "SMTP_HOST") {
		t.Fatalf("expected SMTP configuration error, sender=%T err=%v", sender, err)
	}
}
