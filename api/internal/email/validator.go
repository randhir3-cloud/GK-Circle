package email

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
)

// MaxRecipientsPerMessage bounds the upper limit of recipients in a single email send.
const MaxRecipientsPerMessage = 50

// ValidateMessage strictly validates all headers and fields in the EmailMessage.
func ValidateMessage(msg EmailMessage) error {
	if msg.MessageID == "" {
		return ErrEmptyMessageID
	}
	if msg.IdempotencyKey == "" {
		return ErrEmptyIdempotencyKey
	}
	if len(msg.To) == 0 {
		return ErrEmptyRecipient
	}
	if len(msg.To) > MaxRecipientsPerMessage {
		return ErrTooManyRecipients
	}
	if msg.Subject == "" {
		return ErrEmptySubject
	}
	if len([]rune(msg.Subject)) > 180 {
		return ErrSubjectTooLong
	}
	if msg.HTMLBody == "" && msg.PlainText == "" {
		return ErrEmptyBody
	}

	// Validate sender address
	if err := validateAddress(msg.From); err != nil {
		return ErrInvalidSender
	}

	// Validate recipient addresses
	for _, rec := range msg.To {
		if err := validateAddress(rec); err != nil {
			return fmt.Errorf("invalid recipient address: %w", err)
		}
	}

	// Validate reply-to address if present
	if msg.ReplyTo != nil {
		if err := validateAddress(*msg.ReplyTo); err != nil {
			return ErrInvalidReplyTo
		}
	}

	// CRLF injection checks
	if hasCRLF(msg.Subject) {
		return ErrCRLFInjection
	}
	if hasCRLF(msg.From.Name) || hasCRLF(msg.From.Address) {
		return ErrCRLFInjection
	}
	for _, rec := range msg.To {
		if hasCRLF(rec.Name) || hasCRLF(rec.Address) {
			return ErrCRLFInjection
		}
	}
	if msg.ReplyTo != nil {
		if hasCRLF(msg.ReplyTo.Name) || hasCRLF(msg.ReplyTo.Address) {
			return ErrCRLFInjection
		}
	}

	return nil
}

func validateAddress(rec EmailRecipient) error {
	if rec.Address == "" {
		return errors.New("empty address")
	}

	// Ensure only one mailbox is present (ParseAddress parses the first one it finds)
	if strings.Contains(rec.Address, ",") || strings.Contains(rec.Address, ";") {
		return errors.New("multiple addresses not allowed where one is expected")
	}

	parsed, err := mail.ParseAddress(rec.Address)
	if err != nil {
		return err
	}

	// Parsed address must match the cleaned input (excluding names in brackets)
	cleanedInput := strings.TrimSpace(rec.Address)
	if parsed.Address != cleanedInput && !strings.Contains(cleanedInput, "<") {
		return errors.New("address parsing mismatch")
	}

	if len(rec.Name) > 255 {
		return errors.New("display name too long")
	}

	return nil
}

func hasCRLF(s string) bool {
	return strings.ContainsAny(s, "\r\n")
}
