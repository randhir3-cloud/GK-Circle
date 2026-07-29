# GK Circle Email Styling Architecture & Template Guide

This document outlines the design aesthetics, template architecture, compilation verification, and responsive guidelines for email communications in GK Circle.

---

## 1. Design Aesthetics & Visual Identity

All templates conform to the premium visual aesthetic of GK Circle:
* **Typography**: Clean sans-serif font stack (`System-UI, -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif`) to guarantee readability without depending on web-font permissions.
* **Palette**: Curated premium palette:
  * **Primary (Neo-Brutalist Highlight)**: `#FF5B5B` (Coral Accent)
  * **Background**: `#F3F4F6` (Light grey envelope)
  * **Card Container**: `#FFFFFF` (Solid white)
  * **Text/Borders**: `#1A1A1A` (Deep charcoal, ensuring extremely high accessibility contrast)
* **Shadows & Accents**: 4px solid borders with `#1A1A1A` 4px solid offset shadows (`box-shadow: 4px 4px 0px #1A1A1A`) on primary buttons to align with the frontend branding guidelines.
* **Accessibility**: Contrast ratios exceed 7:1 for all textual contents.

---

## 2. Template Structure & Version Comments

To allow validation, every template contains structured version identifiers:
* **HTML**: Standardized version comment block at the beginning of the `<body>` element.
* **Plaintext**: The first non-empty line of the text template contains the version descriptor.

The template renderer prepends the metadata block automatically during execution to prevent HTML parsers from stripping comments:
```html
<!-- GK Circle Email Template: <name> <version> -->
```

### Supported Templates & Files
1. **Quiz Invitation** (`invitation/`)
   * Subject: `You are invited to participate in the quiz: {{.QuizTitle}}`
2. **Achievement Earned** (`achievement/`)
   * Subject: `Congratulations! You've earned a new achievement: {{.AchievementTitle}}`
3. **Certificate Issued** (`certificate/`)
   * Subject: `Your Certificate of Completion is ready: {{.CourseTitle}}`
4. **Weekly Progress** (`weekly_report/`)
   * Subject: `GK Circle Weekly Progress Report`
5. **Community Update** (`notification/`)
   * Subject: `Community Update: {{.NotificationTitle}}`
6. **Admin Announcement** (`admin/`)
   * Subject: `Important Announcement: {{.Title}}`
7. **Security Alert** (`security_alert/`)
   * Subject: `Critical Security Alert`

---

## 3. Operations & Code Requirements

* **Zero Map Interface**: Every template is matched to a concrete Go typed struct input parameters:
  * `QuizInvitationInput`
  * `AchievementInput`
  * `CertificateInput`
  * `WeeklyProgressInput`
  * `CommunityNotificationInput`
  * `AdminAnnouncementInput`
  * `SecurityAlertInput`
* **Startup Compilation**: The renderer utilizes `go:embed` to read template files from the disk and parses them during service instantiation. Any syntax errors inside template structures trigger immediate startup panic, acting as a gate against deploying broken email templates to production.
* **No Inline CSS Engines**: Styles are inline directly in the raw template sources to ensure compatibility with Gmail, Outlook, Apple Mail, and mobile email clients without depending on performance-heavy runtime CSS inline engines.
