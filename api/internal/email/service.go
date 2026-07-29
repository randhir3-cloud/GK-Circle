package email

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"go.uber.org/zap"
)

// MetricsHook provides telemetry hooks to record latency and failures without adding dependency.
type MetricsHook interface {
	RecordEmailAccepted(emailType string, provider string, latency time.Duration)
	RecordEmailFailure(emailType string, provider string, kind string)
	RecordEmailRetry(emailType string, provider string, attempt int)
}

// NopMetricsHook implements a no-op telemetry MetricsHook.
type NopMetricsHook struct{}

func (NopMetricsHook) RecordEmailAccepted(string, string, time.Duration) {}
func (NopMetricsHook) RecordEmailFailure(string, string, string)         {}
func (NopMetricsHook) RecordEmailRetry(string, string, int)              {}

// TransactionalEmailService implements transactional email delivery pipelines.
type TransactionalEmailService struct {
	config      config.EmailConfig
	provider    EmailProvider
	renderer    *TemplateRenderer
	urlBuilder  *AppURLBuilder
	logger      *zap.Logger
	clock       Clock
	metricsHook MetricsHook
}

// NewTransactionalEmailService constructs a TransactionalEmailService instance.
func NewTransactionalEmailService(
	cfg config.EmailConfig,
	provider EmailProvider,
	renderer *TemplateRenderer,
	urlBuilder *AppURLBuilder,
	logger *zap.Logger,
	clock Clock,
	metricsHook MetricsHook,
) *TransactionalEmailService {
	if logger == nil {
		logger = zap.NewNop()
	}
	if clock == nil {
		clock = SystemClock{}
	}
	if metricsHook == nil {
		metricsHook = NopMetricsHook{}
	}

	return &TransactionalEmailService{
		config:      cfg,
		provider:    provider,
		renderer:    renderer,
		urlBuilder:  urlBuilder,
		logger:      logger,
		clock:       clock,
		metricsHook: metricsHook,
	}
}

// HashIdempotencyKey computes a truncated SHA-256 hash of the idempotency key for safe logging.
func HashIdempotencyKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:8])
}

// QuizInvitationInput represents parameters for quiz invitations.
type QuizInvitationInput struct {
	InvitationID string
	Recipient    EmailRecipient
	InviterName  string
	QuizTitle    string
	QuizID       string
	ExpiresAt    time.Time
}

// SendQuizInvitation renders and delivers a quiz invitation email.
func (s *TransactionalEmailService) SendQuizInvitation(ctx context.Context, input QuizInvitationInput) (SendResult, error) {
	if !isSafeIdentifier(input.QuizID) || !isSafeIdentifier(input.InvitationID) {
		return SendResult{}, errors.New("unsafe business identifiers")
	}

	invURL, err := s.urlBuilder.QuizInvitation(input.QuizID, input.InvitationID)
	if err != nil {
		return SendResult{}, fmt.Errorf("failed to build invitation URL: %w", err)
	}

	data := map[string]interface{}{
		"RecipientName": input.Recipient.Name,
		"InviterName":   input.InviterName,
		"QuizTitle":     input.QuizTitle,
		"InvitationURL": invURL,
		"ExpiresAt":     input.ExpiresAt.Format(time.RFC1123),
	}

	subject, html, text, err := s.renderer.Render(TemplateQuizInvitation, data)
	if err != nil {
		s.logger.Error("Template rendering failed", zap.String("email_type", string(EmailQuizInvitation)), zap.Error(err))
		s.metricsHook.RecordEmailFailure(string(EmailQuizInvitation), string(s.config.Provider), "render_failure")
		return SendResult{}, err
	}

	msg := EmailMessage{
		MessageID:      uuid.New().String(),
		IdempotencyKey: fmt.Sprintf("quiz-invitation/%s", input.InvitationID),
		EmailType:      EmailQuizInvitation,
		To:             []EmailRecipient{input.Recipient},
		From:           EmailRecipient{Name: s.config.FromName, Address: s.config.From},
		Subject:        subject,
		HTMLBody:       html,
		PlainText:      text,
	}

	if s.config.ReplyTo != "" {
		msg.ReplyTo = &EmailRecipient{Name: s.config.FromName, Address: s.config.ReplyTo}
	}

	return s.sendAndLog(ctx, msg, "invitation", "v1")
}

// AchievementInput represents achievement email parameters.
type AchievementInput struct {
	AchievementID    string
	Recipient        EmailRecipient
	AchievementTitle string
	Description      string
	EarnedAt         time.Time
}

// SendAchievement renders and delivers an achievement email.
func (s *TransactionalEmailService) SendAchievement(ctx context.Context, input AchievementInput) (SendResult, error) {
	if !isSafeIdentifier(input.AchievementID) {
		return SendResult{}, errors.New("unsafe achievement identifier")
	}

	data := map[string]interface{}{
		"RecipientName":    input.Recipient.Name,
		"AchievementTitle": input.AchievementTitle,
		"Description":      input.Description,
		"EarnedAt":         input.EarnedAt.Format(time.RFC1123),
	}

	subject, html, text, err := s.renderer.Render(TemplateAchievement, data)
	if err != nil {
		s.logger.Error("Template rendering failed", zap.String("email_type", string(EmailAchievement)), zap.Error(err))
		s.metricsHook.RecordEmailFailure(string(EmailAchievement), string(s.config.Provider), "render_failure")
		return SendResult{}, err
	}

	msg := EmailMessage{
		MessageID:      uuid.New().String(),
		IdempotencyKey: fmt.Sprintf("achievement/%s", input.AchievementID),
		EmailType:      EmailAchievement,
		To:             []EmailRecipient{input.Recipient},
		From:           EmailRecipient{Name: s.config.FromName, Address: s.config.From},
		Subject:        subject,
		HTMLBody:       html,
		PlainText:      text,
	}

	if s.config.ReplyTo != "" {
		msg.ReplyTo = &EmailRecipient{Name: s.config.FromName, Address: s.config.ReplyTo}
	}

	return s.sendAndLog(ctx, msg, "achievement", "v1")
}

// CertificateInput represents certificate parameters.
type CertificateInput struct {
	CertificateID  string
	Recipient      EmailRecipient
	CourseTitle    string
	CertificateURL string
	IssuedAt       time.Time
}

// SendCertificate renders and delivers a certificate completion email.
func (s *TransactionalEmailService) SendCertificate(ctx context.Context, input CertificateInput) (SendResult, error) {
	if !isSafeIdentifier(input.CertificateID) {
		return SendResult{}, errors.New("unsafe certificate identifier")
	}

	data := map[string]interface{}{
		"RecipientName":  input.Recipient.Name,
		"CourseTitle":    input.CourseTitle,
		"CertificateURL": input.CertificateURL,
		"IssuedAt":       input.IssuedAt.Format(time.RFC1123),
	}

	subject, html, text, err := s.renderer.Render(TemplateCertificate, data)
	if err != nil {
		s.logger.Error("Template rendering failed", zap.String("email_type", string(EmailCertificate)), zap.Error(err))
		s.metricsHook.RecordEmailFailure(string(EmailCertificate), string(s.config.Provider), "render_failure")
		return SendResult{}, err
	}

	msg := EmailMessage{
		MessageID:      uuid.New().String(),
		IdempotencyKey: fmt.Sprintf("certificate/%s", input.CertificateID),
		EmailType:      EmailCertificate,
		To:             []EmailRecipient{input.Recipient},
		From:           EmailRecipient{Name: s.config.FromName, Address: s.config.From},
		Subject:        subject,
		HTMLBody:       html,
		PlainText:      text,
	}

	if s.config.ReplyTo != "" {
		msg.ReplyTo = &EmailRecipient{Name: s.config.FromName, Address: s.config.ReplyTo}
	}

	return s.sendAndLog(ctx, msg, "certificate", "v1")
}

// WeeklyProgressInput represents weekly reports parameters.
type WeeklyProgressInput struct {
	Recipient      EmailRecipient
	CompletedCount int
	AverageScore   int
	RankProgress   string
	ReportURL      string
	WeekStartDate  time.Time
}

// SendWeeklyProgress renders and delivers a weekly progress report digest.
func (s *TransactionalEmailService) SendWeeklyProgress(ctx context.Context, input WeeklyProgressInput) (SendResult, error) {
	data := map[string]interface{}{
		"RecipientName":  input.Recipient.Name,
		"CompletedCount": input.CompletedCount,
		"AverageScore":   input.AverageScore,
		"RankProgress":   input.RankProgress,
		"ReportURL":      input.ReportURL,
	}

	subject, html, text, err := s.renderer.Render(TemplateWeeklyReport, data)
	if err != nil {
		s.logger.Error("Template rendering failed", zap.String("email_type", string(EmailWeeklyReport)), zap.Error(err))
		s.metricsHook.RecordEmailFailure(string(EmailWeeklyReport), string(s.config.Provider), "render_failure")
		return SendResult{}, err
	}

	msg := EmailMessage{
		MessageID:      uuid.New().String(),
		IdempotencyKey: fmt.Sprintf("weekly-progress/%s-%s", input.Recipient.Address, input.WeekStartDate.Format("20060102")),
		EmailType:      EmailWeeklyReport,
		To:             []EmailRecipient{input.Recipient},
		From:           EmailRecipient{Name: s.config.FromName, Address: s.config.From},
		Subject:        subject,
		HTMLBody:       html,
		PlainText:      text,
	}

	if s.config.ReplyTo != "" {
		msg.ReplyTo = &EmailRecipient{Name: s.config.FromName, Address: s.config.ReplyTo}
	}

	return s.sendAndLog(ctx, msg, "weekly_report", "v1")
}

// CommunityNotificationInput represents community updates parameters.
type CommunityNotificationInput struct {
	NotificationID    string
	Recipient         EmailRecipient
	NotificationTitle string
	ContentBody       string
	ActionURL         string
}

// SendCommunityNotification renders and delivers a community announcement update.
func (s *TransactionalEmailService) SendCommunityNotification(ctx context.Context, input CommunityNotificationInput) (SendResult, error) {
	if !isSafeIdentifier(input.NotificationID) {
		return SendResult{}, errors.New("unsafe notification identifier")
	}

	data := map[string]interface{}{
		"RecipientName":     input.Recipient.Name,
		"NotificationTitle": input.NotificationTitle,
		"ContentBody":       input.ContentBody,
		"ActionURL":         input.ActionURL,
	}

	subject, html, text, err := s.renderer.Render(TemplateCommunityNotification, data)
	if err != nil {
		s.logger.Error("Template rendering failed", zap.String("email_type", string(EmailCommunityNotification)), zap.Error(err))
		s.metricsHook.RecordEmailFailure(string(EmailCommunityNotification), string(s.config.Provider), "render_failure")
		return SendResult{}, err
	}

	msg := EmailMessage{
		MessageID:      uuid.New().String(),
		IdempotencyKey: fmt.Sprintf("community-notification/%s", input.NotificationID),
		EmailType:      EmailCommunityNotification,
		To:             []EmailRecipient{input.Recipient},
		From:           EmailRecipient{Name: s.config.FromName, Address: s.config.From},
		Subject:        subject,
		HTMLBody:       html,
		PlainText:      text,
	}

	if s.config.ReplyTo != "" {
		msg.ReplyTo = &EmailRecipient{Name: s.config.FromName, Address: s.config.ReplyTo}
	}

	return s.sendAndLog(ctx, msg, "notification", "v1")
}

// AdminAnnouncementInput represents admin broadcasts parameters.
type AdminAnnouncementInput struct {
	AnnouncementID string
	Recipient      EmailRecipient
	Title          string
	ContentBody    string
	ActionURL      string
}

// SendAdminAnnouncement renders and delivers an admin-level broadcast message.
func (s *TransactionalEmailService) SendAdminAnnouncement(ctx context.Context, input AdminAnnouncementInput) (SendResult, error) {
	if !isSafeIdentifier(input.AnnouncementID) {
		return SendResult{}, errors.New("unsafe announcement identifier")
	}

	data := map[string]interface{}{
		"RecipientName": input.Recipient.Name,
		"Title":         input.Title,
		"ContentBody":   input.ContentBody,
		"ActionURL":     input.ActionURL,
	}

	subject, html, text, err := s.renderer.Render(TemplateAdminAnnouncement, data)
	if err != nil {
		s.logger.Error("Template rendering failed", zap.String("email_type", string(EmailAdminAnnouncement)), zap.Error(err))
		s.metricsHook.RecordEmailFailure(string(EmailAdminAnnouncement), string(s.config.Provider), "render_failure")
		return SendResult{}, err
	}

	msg := EmailMessage{
		MessageID:      uuid.New().String(),
		IdempotencyKey: fmt.Sprintf("admin-announcement/%s", input.AnnouncementID),
		EmailType:      EmailAdminAnnouncement,
		To:             []EmailRecipient{input.Recipient},
		From:           EmailRecipient{Name: s.config.FromName, Address: s.config.From},
		Subject:        subject,
		HTMLBody:       html,
		PlainText:      text,
	}

	if s.config.ReplyTo != "" {
		msg.ReplyTo = &EmailRecipient{Name: s.config.FromName, Address: s.config.ReplyTo}
	}

	return s.sendAndLog(ctx, msg, "admin", "v1")
}

// SecurityAlertInput represents security event parameters.
type SecurityAlertInput struct {
	SecurityEventID string
	Recipient       EmailRecipient
	AlertDetails    string
	ActionTaken     string
	Timestamp       time.Time
	SettingsURL     string
}

// SendSecurityAlert renders and delivers a critical mandatory security alert.
func (s *TransactionalEmailService) SendSecurityAlert(ctx context.Context, input SecurityAlertInput) (SendResult, error) {
	if !isSafeIdentifier(input.SecurityEventID) {
		return SendResult{}, errors.New("unsafe security event identifier")
	}

	data := map[string]interface{}{
		"RecipientName": input.Recipient.Name,
		"AlertDetails":  input.AlertDetails,
		"ActionTaken":   input.ActionTaken,
		"Timestamp":     input.Timestamp.Format(time.RFC1123),
		"SettingsURL":   input.SettingsURL,
	}

	subject, html, text, err := s.renderer.Render(TemplateSecurityAlert, data)
	if err != nil {
		s.logger.Error("Template rendering failed", zap.String("email_type", string(EmailSecurityAlert)), zap.Error(err))
		s.metricsHook.RecordEmailFailure(string(EmailSecurityAlert), string(s.config.Provider), "render_failure")
		return SendResult{}, err
	}

	msg := EmailMessage{
		MessageID:      uuid.New().String(),
		IdempotencyKey: fmt.Sprintf("security-alert/%s", input.SecurityEventID),
		EmailType:      EmailSecurityAlert,
		To:             []EmailRecipient{input.Recipient},
		From:           EmailRecipient{Name: s.config.FromName, Address: s.config.From},
		Subject:        subject,
		HTMLBody:       html,
		PlainText:      text,
	}

	if s.config.ReplyTo != "" {
		msg.ReplyTo = &EmailRecipient{Name: s.config.FromName, Address: s.config.ReplyTo}
	}

	return s.sendAndLog(ctx, msg, "security_alert", "v1")
}

func (s *TransactionalEmailService) sendAndLog(
	ctx context.Context,
	msg EmailMessage,
	templateName string,
	templateVersion string,
) (SendResult, error) {
	if err := ValidateMessage(msg); err != nil {
		s.logger.Error("Message validation failed",
			zap.String("message_id", msg.MessageID),
			zap.String("idempotency_key_hash", HashIdempotencyKey(msg.IdempotencyKey)),
			zap.Error(err),
		)
		s.metricsHook.RecordEmailFailure(string(msg.EmailType), string(s.config.Provider), "validation_failure")
		return SendResult{}, err
	}

	recipientDomain := ""
	if len(msg.To) > 0 {
		parsed, _ := mail.ParseAddress(msg.To[0].Address)
		if parsed != nil {
			parts := strings.Split(parsed.Address, "@")
			if len(parts) == 2 {
				recipientDomain = parts[1]
			}
		}
	}

	startTime := s.clock.Now()
	result, err := s.provider.Send(ctx, msg)
	latency := s.clock.Now().Sub(startTime)

	if err != nil {
		var pErr *ProviderError
		errKind := "unknown"
		httpStatus := 0
		if errors.As(err, &pErr) {
			errKind = string(pErr.Kind)
			httpStatus = pErr.StatusCode
		}

		s.logger.Error("Email dispatch failed",
			zap.String("message_id", msg.MessageID),
			zap.String("idempotency_key_hash", HashIdempotencyKey(msg.IdempotencyKey)),
			zap.String("email_type", string(msg.EmailType)),
			zap.String("provider", string(s.config.Provider)),
			zap.String("recipient_domain", recipientDomain),
			zap.String("template_name", templateName),
			zap.String("template_version", templateVersion),
			zap.String("status", "failed"),
			zap.Int64("latency_ms", latency.Milliseconds()),
			zap.String("error_kind", errKind),
			zap.Int("http_status", httpStatus),
			zap.Error(err), // Logs wrapped underlying error internally
		)

		s.metricsHook.RecordEmailFailure(string(msg.EmailType), string(s.config.Provider), errKind)
		return SendResult{}, err
	}

	s.logger.Info("Email accepted by provider",
		zap.String("message_id", msg.MessageID),
		zap.String("idempotency_key_hash", HashIdempotencyKey(msg.IdempotencyKey)),
		zap.String("email_type", string(msg.EmailType)),
		zap.String("provider", string(s.config.Provider)),
		zap.String("provider_message_id", result.ProviderMessageID),
		zap.String("recipient_domain", recipientDomain),
		zap.String("template_name", templateName),
		zap.String("template_version", templateVersion),
		zap.String("status", "accepted"),
		zap.Int64("latency_ms", latency.Milliseconds()),
	)

	s.metricsHook.RecordEmailAccepted(string(msg.EmailType), string(s.config.Provider), latency)
	return result, nil
}
