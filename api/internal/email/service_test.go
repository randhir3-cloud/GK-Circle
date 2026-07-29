package email

import (
	"context"
	"testing"
	"time"

	"github.com/randhir3-cloud/GK-Circle-v2/api/config"
	"go.uber.org/zap"
)

type mockProvider struct {
	lastMessage EmailMessage
	result      SendResult
	err         error
}

func (m *mockProvider) Send(ctx context.Context, msg EmailMessage) (SendResult, error) {
	m.lastMessage = msg
	return m.result, m.err
}

type mockMetricsHook struct {
	acceptedCalls int
	failureCalls  int
	retryCalls    int
}

func (m *mockMetricsHook) RecordEmailAccepted(emailType string, provider string, latency time.Duration) {
	m.acceptedCalls++
}

func (m *mockMetricsHook) RecordEmailFailure(emailType string, provider string, kind string) {
	m.failureCalls++
}

func (m *mockMetricsHook) RecordEmailRetry(emailType string, provider string, attempt int) {
	m.retryCalls++
}

func TestTransactionalEmailService_SendQuizInvitation_Success(t *testing.T) {
	cfg := config.EmailConfig{
		Provider: config.ProviderSMTP,
		FromName: "GK Circle",
		From:     "notifications@gkcircle.com",
	}

	provider := &mockProvider{
		result: SendResult{ProviderMessageID: "smtp-success-1", AcceptedAt: time.Now()},
	}

	renderer, err := NewTemplateRenderer()
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	urlBuilder, err := NewAppURLBuilder("https://gkcircle.com", "production")
	if err != nil {
		t.Fatalf("Failed to create urlBuilder: %v", err)
	}

	mClock := mockClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	mMetrics := &mockMetricsHook{}
	logger := zap.NewNop()

	svc := NewTransactionalEmailService(cfg, provider, renderer, urlBuilder, logger, mClock, mMetrics)

	input := QuizInvitationInput{
		InvitationID: "inv-123",
		Recipient:    EmailRecipient{Name: "Candidate", Address: "candidate@gkcircle.com"},
		InviterName:  "Inviter",
		QuizTitle:    "General Knowledge",
		QuizID:       "quiz-456",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	result, err := svc.SendQuizInvitation(context.Background(), input)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.ProviderMessageID != "smtp-success-1" {
		t.Errorf("Expected 'smtp-success-1', got: %s", result.ProviderMessageID)
	}

	if provider.lastMessage.EmailType != EmailQuizInvitation {
		t.Errorf("Expected EmailQuizInvitation type, got %s", provider.lastMessage.EmailType)
	}

	if provider.lastMessage.IdempotencyKey != "quiz-invitation/inv-123" {
		t.Errorf("Expected idempotency key 'quiz-invitation/inv-123', got %s", provider.lastMessage.IdempotencyKey)
	}

	if mMetrics.acceptedCalls != 1 {
		t.Errorf("Expected 1 accepted metrics call, got %d", mMetrics.acceptedCalls)
	}
}

func TestTransactionalEmailService_SendQuizInvitation_UnsafeIdentifiers(t *testing.T) {
	cfg := config.EmailConfig{
		Provider: config.ProviderSMTP,
		FromName: "GK Circle",
		From:     "notifications@gkcircle.com",
	}

	svc := NewTransactionalEmailService(cfg, &mockProvider{}, &TemplateRenderer{}, &AppURLBuilder{}, zap.NewNop(), mockClock{}, nil)

	input := QuizInvitationInput{
		InvitationID: "../inv-123", // unsafe traversal
		Recipient:    EmailRecipient{Name: "Candidate", Address: "candidate@gkcircle.com"},
		InviterName:  "Inviter",
		QuizTitle:    "General Knowledge",
		QuizID:       "quiz-456",
		ExpiresAt:    time.Now().Add(24 * time.Hour),
	}

	_, err := svc.SendQuizInvitation(context.Background(), input)
	if err == nil {
		t.Error("Expected error for unsafe identifier invitation ID, got nil")
	}

	input.InvitationID = "inv-123"
	input.QuizID = "quiz-456; DROP TABLE users" // unsafe injection
	_, err = svc.SendQuizInvitation(context.Background(), input)
	if err == nil {
		t.Error("Expected error for unsafe identifier quiz ID, got nil")
	}
}

func TestTransactionalEmailService_RemainingFlows(t *testing.T) {
	cfg := config.EmailConfig{
		Provider: config.ProviderSMTP,
		FromName: "GK Circle",
		From:     "notifications@gkcircle.com",
	}

	provider := &mockProvider{
		result: SendResult{ProviderMessageID: "success-id", AcceptedAt: time.Now()},
	}

	renderer, _ := NewTemplateRenderer()
	urlBuilder, _ := NewAppURLBuilder("https://gkcircle.com", "production")
	mClock := mockClock{now: time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)}
	mMetrics := &mockMetricsHook{}
	logger := zap.NewNop()

	svc := NewTransactionalEmailService(cfg, provider, renderer, urlBuilder, logger, mClock, mMetrics)
	ctx := context.Background()
	recipient := EmailRecipient{Name: "Student", Address: "student@gkcircle.com"}

	t.Run("achievement", func(t *testing.T) {
		input := AchievementInput{
			AchievementID:    "ach-1",
			Recipient:        recipient,
			AchievementTitle: "Quiz Master",
			Description:      "Scored 100% in 5 quizzes",
			EarnedAt:         time.Now(),
		}
		_, err := svc.SendAchievement(ctx, input)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if provider.lastMessage.EmailType != EmailAchievement {
			t.Errorf("Expected EmailAchievement, got %s", provider.lastMessage.EmailType)
		}
	})

	t.Run("certificate", func(t *testing.T) {
		input := CertificateInput{
			CertificateID:  "cert-1",
			Recipient:      recipient,
			CourseTitle:    "Modern History",
			CertificateURL: "https://gkcircle.com/certs/cert-1",
			IssuedAt:       time.Now(),
		}
		_, err := svc.SendCertificate(ctx, input)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if provider.lastMessage.EmailType != EmailCertificate {
			t.Errorf("Expected EmailCertificate, got %s", provider.lastMessage.EmailType)
		}
	})

	t.Run("weekly_progress", func(t *testing.T) {
		input := WeeklyProgressInput{
			Recipient:      recipient,
			CompletedCount: 4,
			AverageScore:   85,
			RankProgress:   "Up 2 places",
			ReportURL:      "https://gkcircle.com/reports/weekly",
			WeekStartDate:  time.Now(),
		}
		_, err := svc.SendWeeklyProgress(ctx, input)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if provider.lastMessage.EmailType != EmailWeeklyReport {
			t.Errorf("Expected EmailWeeklyReport, got %s", provider.lastMessage.EmailType)
		}
	})

	t.Run("community_notification", func(t *testing.T) {
		input := CommunityNotificationInput{
			NotificationID:    "notif-1",
			Recipient:         recipient,
			NotificationTitle: "New Quiz Released",
			ContentBody:       "A new state PCS prep quiz has been posted.",
			ActionURL:         "https://gkcircle.com/quizzes/new",
		}
		_, err := svc.SendCommunityNotification(ctx, input)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if provider.lastMessage.EmailType != EmailCommunityNotification {
			t.Errorf("Expected EmailCommunityNotification, got %s", provider.lastMessage.EmailType)
		}
	})

	t.Run("admin_announcement", func(t *testing.T) {
		input := AdminAnnouncementInput{
			AnnouncementID: "ann-1",
			Recipient:      recipient,
			Title:          "Scheduled Maintenance",
			ContentBody:    "The site will be down for 2 hours.",
			ActionURL:      "https://gkcircle.com/status",
		}
		_, err := svc.SendAdminAnnouncement(ctx, input)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if provider.lastMessage.EmailType != EmailAdminAnnouncement {
			t.Errorf("Expected EmailAdminAnnouncement, got %s", provider.lastMessage.EmailType)
		}
	})

	t.Run("security_alert", func(t *testing.T) {
		input := SecurityAlertInput{
			SecurityEventID: "sec-1",
			Recipient:       recipient,
			AlertDetails:    "Login from new IP address",
			ActionTaken:     "Temporary session lock",
			Timestamp:       time.Now(),
			SettingsURL:     "https://gkcircle.com/settings",
		}
		_, err := svc.SendSecurityAlert(ctx, input)
		if err != nil {
			t.Fatalf("Expected no error, got: %v", err)
		}
		if provider.lastMessage.EmailType != EmailSecurityAlert {
			t.Errorf("Expected EmailSecurityAlert, got %s", provider.lastMessage.EmailType)
		}
	})
}
