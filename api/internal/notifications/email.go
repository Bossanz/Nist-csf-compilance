package notifications

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"strings"
)

type EmailMessage struct {
	To      string
	Subject string
	Text    string
}

type EmailSender interface {
	Send(context.Context, EmailMessage) error
}

type logEmailSender struct {
	logger *log.Logger
}

func newLogEmailSender(logger *log.Logger) *logEmailSender {
	if logger == nil {
		logger = log.Default()
	}
	return &logEmailSender{logger: logger}
}

func (s *logEmailSender) Send(ctx context.Context, message EmailMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.logger.Printf("email mode=log to=%s subject=%q body=%s", message.To, message.Subject, message.Text)
	return nil
}

type smtpSendMailFunc func(address string, auth smtp.Auth, from string, recipients []string, message []byte) error

type smtpEmailSender struct {
	host     string
	port     string
	username string
	password string
	from     string
	sendMail smtpSendMailFunc
}

func (s *smtpEmailSender) Send(ctx context.Context, message EmailMessage) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	auth := smtp.Auth(nil)
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}
	if s.sendMail == nil {
		s.sendMail = smtp.SendMail
	}
	return s.sendMail(s.host+":"+s.port, auth, s.from, []string{message.To}, plainTextMessage(s.from, message))
}

func plainTextMessage(from string, message EmailMessage) []byte {
	text := strings.ReplaceAll(strings.ReplaceAll(message.Text, "\r\n", "\n"), "\r", "\n")
	text = strings.ReplaceAll(text, "\n", "\r\n")
	return []byte("To: " + message.To + "\r\n" +
		"From: " + from + "\r\n" +
		"Subject: " + message.Subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n\r\n" + text)
}

func NewFromEnv(logger *log.Logger, getenv func(string) string) (EmailSender, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	mode := strings.ToLower(strings.TrimSpace(getenv("EMAIL_MODE")))
	if mode == "" || mode == "log" {
		return newLogEmailSender(logger), nil
	}
	if mode != "smtp" {
		return nil, fmt.Errorf("unsupported EMAIL_MODE %q", mode)
	}
	host := strings.TrimSpace(getenv("SMTP_HOST"))
	from := strings.TrimSpace(getenv("SMTP_FROM"))
	if host == "" {
		return nil, fmt.Errorf("SMTP_HOST is required when EMAIL_MODE=smtp")
	}
	if from == "" {
		return nil, fmt.Errorf("SMTP_FROM is required when EMAIL_MODE=smtp")
	}
	port := strings.TrimSpace(getenv("SMTP_PORT"))
	if port == "" {
		port = "587"
	}
	if _, err := strconv.Atoi(port); err != nil {
		return nil, fmt.Errorf("SMTP_PORT must be numeric")
	}
	return &smtpEmailSender{host: host, port: port, username: getenv("SMTP_USERNAME"), password: getenv("SMTP_PASSWORD"), from: from}, nil
}
