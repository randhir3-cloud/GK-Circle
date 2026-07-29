package email

import (
	"bytes"
	"embed"
	"fmt"
	htmltemplate "html/template"
	"io"
	"io/fs"
	"path"
	"strings"
	texttemplate "text/template"
)

//go:embed templates
var emailTemplatesFS embed.FS

type TemplateName string

const (
	TemplateQuizInvitation        TemplateName = "invitation"
	TemplateAchievement           TemplateName = "achievement"
	TemplateCertificate           TemplateName = "certificate"
	TemplateCommunityNotification TemplateName = "notification"
	TemplateWeeklyReport          TemplateName = "weekly_report"
	TemplateAdminAnnouncement     TemplateName = "admin"
	TemplateSecurityAlert         TemplateName = "security_alert"
)

// TemplateDefinition defines template metadata and version.
type TemplateDefinition struct {
	Name    TemplateName
	Version string
}

// SupportedTemplates lists all template names and versions registered in the application.
var SupportedTemplates = []TemplateDefinition{
	{Name: TemplateQuizInvitation, Version: "v1"},
	{Name: TemplateAchievement, Version: "v1"},
	{Name: TemplateCertificate, Version: "v1"},
	{Name: TemplateCommunityNotification, Version: "v1"},
	{Name: TemplateWeeklyReport, Version: "v1"},
	{Name: TemplateAdminAnnouncement, Version: "v1"},
	{Name: TemplateSecurityAlert, Version: "v1"},
}

// CompiledEmailTemplate stores pre-compiled templates for a single flow.
type CompiledEmailTemplate struct {
	Subject   *texttemplate.Template
	HTML      *htmltemplate.Template
	PlainText *texttemplate.Template
	Version   string
}

// TemplateRenderer manages the startup compilation and execution of transactional email templates.
type TemplateRenderer struct {
	templates map[TemplateName]CompiledEmailTemplate
}

// NewTemplateRenderer reads, compiles, and validates all registered templates.
func NewTemplateRenderer() (*TemplateRenderer, error) {
	registry := make(map[TemplateName]CompiledEmailTemplate)

	// 1. Verify that there are no duplicate template names registered
	seen := make(map[TemplateName]bool)
	for _, def := range SupportedTemplates {
		if seen[def.Name] {
			return nil, fmt.Errorf("duplicate template name registered: %s", def.Name)
		}
		seen[def.Name] = true
	}

	// 2. Walk the embedded filesystem and ensure no unregistered template directories exist on disk
	err := fs.WalkDir(emailTemplatesFS, "templates", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// If it is a direct subdirectory of templates (e.g. templates/invitation)
		if d.IsDir() && p != "templates" && path.Dir(p) == "templates" {
			dirName := path.Base(p)
			found := false
			for _, def := range SupportedTemplates {
				if string(def.Name) == dirName {
					found = true
					break
				}
			}
			if !found {
				return fmt.Errorf("unregistered template directory found on disk: %s", p)
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed walking embedded templates directory: %w", err)
	}

	for _, def := range SupportedTemplates {
		subjectPath := path.Join("templates", string(def.Name), "email.subject.tmpl")
		htmlPath := path.Join("templates", string(def.Name), "email.body.html")
		txtPath := path.Join("templates", string(def.Name), "email.body.txt")

		// Compile Subject Template
		subBytes, err := readEmbedFile(subjectPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read subject template for %s at %s: %w", def.Name, subjectPath, err)
		}
		subTmpl, err := texttemplate.New(string(def.Name) + "_subject").Parse(string(subBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to parse subject template for %s: %w", def.Name, err)
		}

		// Compile HTML Body Template
		htmlBytes, err := readEmbedFile(htmlPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read html template for %s at %s: %w", def.Name, htmlPath, err)
		}
		htmlTmpl, err := htmltemplate.New(string(def.Name) + "_html").Parse(string(htmlBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to parse html template for %s: %w", def.Name, err)
		}

		// Compile PlainText Body Template
		txtBytes, err := readEmbedFile(txtPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read plaintext template for %s at %s: %w", def.Name, txtPath, err)
		}
		txtTmpl, err := texttemplate.New(string(def.Name) + "_txt").Parse(string(txtBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to parse plaintext template for %s: %w", def.Name, err)
		}

		registry[def.Name] = CompiledEmailTemplate{
			Subject:   subTmpl,
			HTML:      htmlTmpl,
			PlainText: txtTmpl,
			Version:   def.Version,
		}
	}

	return &TemplateRenderer{templates: registry}, nil
}

// Render executes subject, HTML, and plaintext templates with the provided data payload.
func (r *TemplateRenderer) Render(name TemplateName, data interface{}) (subject, htmlBody, plainText string, err error) {
	tmpl, ok := r.templates[name]
	if !ok {
		return "", "", "", fmt.Errorf("template %s not found in registry", name)
	}

	var subBuf bytes.Buffer
	if err := tmpl.Subject.Execute(&subBuf, data); err != nil {
		return "", "", "", fmt.Errorf("failed to execute subject template for %s: %w", name, err)
	}
	subjectStr := strings.TrimSpace(subBuf.String())

	var htmlBuf bytes.Buffer
	if err := tmpl.HTML.Execute(&htmlBuf, data); err != nil {
		return "", "", "", fmt.Errorf("failed to execute html template for %s: %w", name, err)
	}
	htmlOut := fmt.Sprintf("<!-- GK Circle Email Template: %s %s -->\n%s", name, tmpl.Version, htmlBuf.String())

	var txtBuf bytes.Buffer
	if err := tmpl.PlainText.Execute(&txtBuf, data); err != nil {
		return "", "", "", fmt.Errorf("failed to execute plaintext template for %s: %w", name, err)
	}
	txtOut := fmt.Sprintf("GK Circle Email Template: %s %s\n\n%s", name, tmpl.Version, txtBuf.String())

	return subjectStr, htmlOut, txtOut, nil
}

func readEmbedFile(filePath string) ([]byte, error) {
	f, err := emailTemplatesFS.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return io.ReadAll(f)
}
