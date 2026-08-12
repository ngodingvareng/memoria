// Package mailer sends transactional email (currently just the
// password-reset link) over plain SMTP — no third-party API, matching
// how object storage in this codebase points at whatever S3-compatible
// endpoint is configured rather than a specific vendor's SDK.
package mailer

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	// From is used both as the envelope sender and the message's From
	// header.
	From string
}

type SMTPMailer struct {
	cfg SMTPConfig
}

func NewSMTPMailer(cfg SMTPConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

// Send implements [usecase.Mailer]. ctx is accepted for interface
// consistency with the rest of this codebase's external-service calls
// (Storage, PasswordHasher, ...) but net/smtp's SendMail predates
// context support and can't actually be cancelled mid-flight.
func (m *SMTPMailer) Send(_ context.Context, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%s", m.cfg.Host, m.cfg.Port)

	// PlainAuth only when credentials are configured — lets a local
	// dev catch-all (e.g. Mailpit/MailHog) that accepts unauthenticated
	// SMTP work without also requiring dummy credentials.
	var auth smtp.Auth
	if m.cfg.Username != "" {
		auth = smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	}

	// SendMail negotiates STARTTLS itself when the server advertises
	// the extension (which is what most SMTP providers on port 587
	// require) — no separate TLS handling needed here.
	if err := smtp.SendMail(
		addr,
		auth,
		m.cfg.From,
		[]string{to},
		buildMessage(m.cfg.From, to, subject, body),
	); err != nil {
		return fmt.Errorf("sending email to %s: %w", to, err)
	}
	return nil
}

// buildMessage is a minimal RFC 822 message — plain text only, no
// multipart/HTML, which is all a reset-link email needs.
func buildMessage(from, to, subject, body string) []byte {
	var msg strings.Builder
	msg.WriteString("From: ")
	msg.WriteString(from)
	msg.WriteString("\r\n")
	msg.WriteString("To: ")
	msg.WriteString(to)
	msg.WriteString("\r\n")
	msg.WriteString("Subject: ")
	msg.WriteString(subject)
	msg.WriteString("\r\n")
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)
	return []byte(msg.String())
}
