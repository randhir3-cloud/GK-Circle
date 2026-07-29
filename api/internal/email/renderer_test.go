package email

import (
	"strings"
	"testing"
	"time"
)

func TestNewTemplateRenderer(t *testing.T) {
	renderer, err := NewTemplateRenderer()
	if err != nil {
		t.Fatalf("Failed to initialize TemplateRenderer: %v", err)
	}

	if len(renderer.templates) != len(SupportedTemplates) {
		t.Errorf("Expected %d compiled templates, got %d", len(SupportedTemplates), len(renderer.templates))
	}
}

func TestRenderQuizInvitation(t *testing.T) {
	renderer, err := NewTemplateRenderer()
	if err != nil {
		t.Fatalf("Failed to initialize TemplateRenderer: %v", err)
	}

	data := map[string]interface{}{
		"RecipientName": "John Doe",
		"InviterName":   "Jane <script>alert('xss')</script> Doe", // unsafe input
		"QuizTitle":     "History & Politics",
		"InvitationURL": "https://gkcircle.com/join/123",
		"ExpiresAt":     time.Now().Add(24 * time.Hour).Format(time.RFC1123),
	}

	subject, html, text, err := renderer.Render(TemplateQuizInvitation, data)
	if err != nil {
		t.Fatalf("Failed to render invitation: %v", err)
	}

	if !strings.Contains(subject, "History & Politics") {
		t.Errorf("Subject missing quiz title, got: %s", subject)
	}

	if !strings.Contains(html, "<!-- GK Circle Email Template: invitation v1 -->") {
		t.Error("HTML missing version marker")
	}
	if !strings.Contains(text, "GK Circle Email Template: invitation v1") {
		t.Error("Text missing version marker")
	}

	if strings.Contains(html, "<script>") {
		t.Errorf("HTML should be escaped, got raw script tag: %s", html)
	}
	if !strings.Contains(html, "&lt;script&gt;") {
		t.Error("Unsafe characters not escaped in HTML")
	}

	if !strings.Contains(text, "<script>") {
		t.Error("Text should remain raw and unescaped")
	}

	if strings.Contains(html, "{{") || strings.Contains(html, "}}") {
		t.Error("HTML contains unresolved template markers")
	}
	if strings.Contains(text, "{{") || strings.Contains(text, "}}") {
		t.Error("Text contains unresolved template markers")
	}
}

func TestRenderAllRegisteredTemplates(t *testing.T) {
	renderer, err := NewTemplateRenderer()
	if err != nil {
		t.Fatalf("Failed to initialize TemplateRenderer: %v", err)
	}

	fixtures := map[TemplateName]map[string]interface{}{
		TemplateQuizInvitation: {
			"RecipientName": "A", "InviterName": "B", "QuizTitle": "C", "InvitationURL": "D", "ExpiresAt": "E",
		},
		TemplateAchievement: {
			"RecipientName": "A", "AchievementTitle": "B", "Description": "C", "EarnedAt": "D",
		},
		TemplateCertificate: {
			"RecipientName": "A", "CourseTitle": "B", "CertificateURL": "C", "IssuedAt": "D",
		},
		TemplateCommunityNotification: {
			"RecipientName": "A", "NotificationTitle": "B", "ContentBody": "C", "ActionURL": "D",
		},
		TemplateWeeklyReport: {
			"RecipientName": "A", "CompletedCount": 5, "AverageScore": 85, "RankProgress": "Up", "ReportURL": "E",
		},
		TemplateAdminAnnouncement: {
			"RecipientName": "A", "Title": "B", "ContentBody": "C", "ActionURL": "D",
		},
		TemplateSecurityAlert: {
			"RecipientName": "A", "AlertDetails": "B", "ActionTaken": "C", "Timestamp": "D", "SettingsURL": "E",
		},
	}

	for _, def := range SupportedTemplates {
		t.Run(string(def.Name), func(t *testing.T) {
			data := fixtures[def.Name]
			sub, html, txt, err := renderer.Render(def.Name, data)
			if err != nil {
				t.Fatalf("Failed to render: %v", err)
			}
			if len(sub) == 0 || len(html) == 0 || len(txt) == 0 {
				t.Error("Rendered outputs cannot be empty")
			}

			// Ensure version header matches
			htmlVersionMarker := "<!-- GK Circle Email Template: " + string(def.Name) + " " + def.Version + " -->"
			txtVersionMarker := "GK Circle Email Template: " + string(def.Name) + " " + def.Version

			if !strings.Contains(html, htmlVersionMarker) {
				t.Errorf("HTML missing expected version marker '%s'", htmlVersionMarker)
			}
			if !strings.Contains(txt, txtVersionMarker) {
				t.Errorf("Text missing expected version marker '%s'", txtVersionMarker)
			}
		})
	}
}

func TestNewTemplateRenderer_DuplicateTemplate(t *testing.T) {
	old := SupportedTemplates
	defer func() { SupportedTemplates = old }()

	SupportedTemplates = []TemplateDefinition{
		{Name: TemplateQuizInvitation, Version: "v1"},
		{Name: TemplateQuizInvitation, Version: "v1"},
	}

	_, err := NewTemplateRenderer()
	if err == nil {
		t.Fatal("Expected error when duplicate template registered, got nil")
	}
	if !strings.Contains(err.Error(), "duplicate template name registered") {
		t.Errorf("Expected duplicate template error, got: %v", err)
	}
}

func TestNewTemplateRenderer_UnregisteredTemplate(t *testing.T) {
	old := SupportedTemplates
	defer func() { SupportedTemplates = old }()

	SupportedTemplates = []TemplateDefinition{
		{Name: TemplateQuizInvitation, Version: "v1"},
		{Name: TemplateAchievement, Version: "v1"},
		{Name: TemplateCertificate, Version: "v1"},
		{Name: TemplateCommunityNotification, Version: "v1"},
		{Name: TemplateWeeklyReport, Version: "v1"},
		{Name: TemplateAdminAnnouncement, Version: "v1"},
	}

	_, err := NewTemplateRenderer()
	if err == nil {
		t.Fatal("Expected error when unregistered template folder exists on disk, got nil")
	}
	if !strings.Contains(err.Error(), "unregistered template directory found on disk") {
		t.Errorf("Expected unregistered template folder error, got: %v", err)
	}
}

