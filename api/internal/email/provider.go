package email

import "context"

// EmailProvider defines the interface that all email senders must implement.
type EmailProvider interface {
	Send(ctx context.Context, message EmailMessage) (SendResult, error)
}
