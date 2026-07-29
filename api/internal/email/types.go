package email

import "time"

// EmailType represents the category of the transactional email.
type EmailType string

const (
	EmailQuizInvitation        EmailType = "invitation"
	EmailAchievement           EmailType = "achievement"
	EmailCertificate           EmailType = "certificate"
	EmailCommunityNotification EmailType = "notification"
	EmailWeeklyReport          EmailType = "weekly_report"
	EmailAdminAnnouncement     EmailType = "admin"
	EmailSecurityAlert         EmailType = "security_alert"
)

// EmailRecipient represents a single recipient display name and email address.
type EmailRecipient struct {
	Name    string
	Address string
}

// EmailMessage contains the core headers and bodies of a compiled transactional email.
type EmailMessage struct {
	MessageID      string
	IdempotencyKey string
	EmailType      EmailType
	To             []EmailRecipient
	From           EmailRecipient
	ReplyTo        *EmailRecipient
	Subject        string
	HTMLBody       string
	PlainText      string
}

// SendResult records successful provider acceptance metadata.
type SendResult struct {
	ProviderMessageID string
	AcceptedAt        time.Time
}
