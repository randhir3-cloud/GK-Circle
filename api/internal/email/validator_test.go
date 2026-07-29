package email

import (
	"strings"
	"testing"
)

func TestValidateMessage_Valid(t *testing.T) {
	msg := EmailMessage{
		MessageID:      "msg-12345",
		IdempotencyKey: "quiz-invitation/inv-12345",
		EmailType:      EmailQuizInvitation,
		To: []EmailRecipient{
			{Name: "Candidate One", Address: "candidate1@gkcircle.com"},
		},
		From:      EmailRecipient{Name: "GK Circle", Address: "notifications@mail.gkcircle.com"},
		Subject:   "You are invited to a quiz!",
		HTMLBody:  "<p>Hello</p>",
		PlainText: "Hello",
	}

	if err := ValidateMessage(msg); err != nil {
		t.Errorf("Expected valid message, got error: %v", err)
	}
}

func TestValidateMessage_InvalidRecipient(t *testing.T) {
	msg := EmailMessage{
		MessageID:      "msg-123",
		IdempotencyKey: "inv-123",
		EmailType:      EmailQuizInvitation,
		To: []EmailRecipient{
			{Name: "Candidate One", Address: "invalid-email"},
		},
		From:      EmailRecipient{Name: "GK", Address: "notifications@mail.gkcircle.com"},
		Subject:   "Subject",
		HTMLBody:  "<p>HTML</p>",
		PlainText: "Plain",
	}

	if err := ValidateMessage(msg); err == nil {
		t.Error("Expected error for invalid recipient address, got nil")
	}
}

func TestValidateMessage_MultipleAddressesRejected(t *testing.T) {
	msg := EmailMessage{
		MessageID:      "msg-123",
		IdempotencyKey: "inv-123",
		EmailType:      EmailQuizInvitation,
		To: []EmailRecipient{
			{Name: "Candidate One", Address: "a@gkcircle.com, b@gkcircle.com"},
		},
		From:      EmailRecipient{Name: "GK", Address: "notifications@mail.gkcircle.com"},
		Subject:   "Subject",
		HTMLBody:  "<p>HTML</p>",
		PlainText: "Plain",
	}

	if err := ValidateMessage(msg); err == nil {
		t.Error("Expected error for multiple addresses in one recipient field, got nil")
	}
}

func TestValidateMessage_EmptySubject(t *testing.T) {
	msg := EmailMessage{
		MessageID:      "msg-123",
		IdempotencyKey: "inv-123",
		EmailType:      EmailQuizInvitation,
		To: []EmailRecipient{
			{Name: "Candidate One", Address: "c1@gkcircle.com"},
		},
		From:      EmailRecipient{Name: "GK", Address: "notifications@mail.gkcircle.com"},
		Subject:   "",
		HTMLBody:  "<p>HTML</p>",
		PlainText: "Plain",
	}

	if err := ValidateMessage(msg); err != ErrEmptySubject {
		t.Errorf("Expected ErrEmptySubject, got %v", err)
	}
}

func TestValidateMessage_SubjectTooLong(t *testing.T) {
	longSubject := strings.Repeat("A", 181)
	msg := EmailMessage{
		MessageID:      "msg-123",
		IdempotencyKey: "inv-123",
		EmailType:      EmailQuizInvitation,
		To: []EmailRecipient{
			{Name: "Candidate One", Address: "c1@gkcircle.com"},
		},
		From:      EmailRecipient{Name: "GK", Address: "notifications@mail.gkcircle.com"},
		Subject:   longSubject,
		HTMLBody:  "<p>HTML</p>",
		PlainText: "Plain",
	}

	if err := ValidateMessage(msg); err != ErrSubjectTooLong {
		t.Errorf("Expected ErrSubjectTooLong, got %v", err)
	}
}

func TestValidateMessage_CRLFInjection(t *testing.T) {
	// CRLF in subject
	msg := EmailMessage{
		MessageID:      "msg-123",
		IdempotencyKey: "inv-123",
		EmailType:      EmailQuizInvitation,
		To: []EmailRecipient{
			{Name: "Candidate", Address: "c1@gkcircle.com"},
		},
		From:      EmailRecipient{Name: "GK", Address: "notifications@mail.gkcircle.com"},
		Subject:   "Subject\nInjection",
		HTMLBody:  "<p>HTML</p>",
		PlainText: "Plain",
	}
	if err := ValidateMessage(msg); err != ErrCRLFInjection {
		t.Errorf("Expected ErrCRLFInjection for subject, got %v", err)
	}

	// CRLF in recipient display name
	msg.Subject = "Valid Subject"
	msg.To[0].Name = "Candidate\rName"
	if err := ValidateMessage(msg); err != ErrCRLFInjection {
		t.Errorf("Expected ErrCRLFInjection for recipient display name, got %v", err)
	}
}

func TestValidateMessage_EmptyBody(t *testing.T) {
	msg := EmailMessage{
		MessageID:      "msg-123",
		IdempotencyKey: "inv-123",
		EmailType:      EmailQuizInvitation,
		To: []EmailRecipient{
			{Name: "Candidate One", Address: "c1@gkcircle.com"},
		},
		From:      EmailRecipient{Name: "GK", Address: "notifications@mail.gkcircle.com"},
		Subject:   "Subject",
		HTMLBody:  "",
		PlainText: "",
	}

	if err := ValidateMessage(msg); err != ErrEmptyBody {
		t.Errorf("Expected ErrEmptyBody, got %v", err)
	}
}

func TestValidateMessage_InvalidSender(t *testing.T) {
	msg := EmailMessage{
		MessageID:      "msg-123",
		IdempotencyKey: "inv-123",
		EmailType:      EmailQuizInvitation,
		To: []EmailRecipient{
			{Name: "Candidate One", Address: "c1@gkcircle.com"},
		},
		From:      EmailRecipient{Name: "GK", Address: "invalid-sender"},
		Subject:   "Subject",
		HTMLBody:  "<p>HTML</p>",
		PlainText: "Plain",
	}

	if err := ValidateMessage(msg); err != ErrInvalidSender {
		t.Errorf("Expected ErrInvalidSender, got %v", err)
	}
}

func TestValidateMessage_InvalidReplyTo(t *testing.T) {
	replyTo := EmailRecipient{Name: "Support", Address: "invalid-reply"}
	msg := EmailMessage{
		MessageID:      "msg-123",
		IdempotencyKey: "inv-123",
		EmailType:      EmailQuizInvitation,
		To: []EmailRecipient{
			{Name: "Candidate One", Address: "c1@gkcircle.com"},
		},
		From:      EmailRecipient{Name: "GK", Address: "notifications@mail.gkcircle.com"},
		ReplyTo:   &replyTo,
		Subject:   "Subject",
		HTMLBody:  "<p>HTML</p>",
		PlainText: "Plain",
	}

	if err := ValidateMessage(msg); err != ErrInvalidReplyTo {
		t.Errorf("Expected ErrInvalidReplyTo, got %v", err)
	}
}

func TestValidateMessage_EmptyIdempotencyKey(t *testing.T) {
	msg := EmailMessage{
		MessageID:      "msg-123",
		IdempotencyKey: "",
		EmailType:      EmailQuizInvitation,
		To: []EmailRecipient{
			{Name: "Candidate One", Address: "c1@gkcircle.com"},
		},
		From:      EmailRecipient{Name: "GK", Address: "notifications@mail.gkcircle.com"},
		Subject:   "Subject",
		HTMLBody:  "<p>HTML</p>",
		PlainText: "Plain",
	}

	if err := ValidateMessage(msg); err != ErrEmptyIdempotencyKey {
		t.Errorf("Expected ErrEmptyIdempotencyKey, got %v", err)
	}
}

func TestValidateMessage_TooManyRecipients(t *testing.T) {
	recipients := make([]EmailRecipient, MaxRecipientsPerMessage+1)
	for i := 0; i <= MaxRecipientsPerMessage; i++ {
		recipients[i] = EmailRecipient{Name: "Candidate", Address: "candidate@gkcircle.com"}
	}

	msg := EmailMessage{
		MessageID:      "msg-123",
		IdempotencyKey: "inv-123",
		EmailType:      EmailQuizInvitation,
		To:             recipients,
		From:           EmailRecipient{Name: "GK", Address: "notifications@mail.gkcircle.com"},
		Subject:        "Subject",
		HTMLBody:       "<p>HTML</p>",
		PlainText:      "Plain",
	}

	if err := ValidateMessage(msg); err != ErrTooManyRecipients {
		t.Errorf("Expected ErrTooManyRecipients, got %v", err)
	}
}
