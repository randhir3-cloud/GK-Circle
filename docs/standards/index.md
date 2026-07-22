# GK Circle Standards Index

Version: 2.4

Status: Mandatory

---

# Purpose

This directory contains all mandatory engineering, architecture, security, testing, AI, frontend, backend, Course, mobile, creator economy, RAG, live exam, admin, documentation, and operational standards for GK Circle.

Every AI agent must load and understand these standards before:

* Creating plans
* Generating code
* Modifying architecture
* Changing infrastructure
* Running migrations
* Implementing features
* Refactoring systems
* Deploying changes

**Repository Constitution:** Non-negotiable agent principles are defined in **`AGENTS.md`** §GK Circle AI Maintainer Constitution. Standards in this directory provide authoritative detail; do not duplicate or weaken them in feature docs.

**Standards changelog:** **`docs/standards/CHANGELOG.md`** — record every governance change (version, date, files, reason, breaking, migration).

**Architecture decisions:** **`docs/architecture/ADR/`** — new ADRs required for architectural changes.

---

# Mandatory Reading Order

The following files must be read in the exact order listed below.

---

# Core Governance Standards

## 1. Architecture Rules

docs/standards/architecture-rules.md

Defines:

* Project architecture
* Scope boundaries
* System design rules
* Migration strategy
* Course architecture
* Domain ownership

---

## 2. Security Rules

docs/standards/security-rules.md

Defines:

* Authentication
* Authorization
* Security auditing
* Secret management
* OWASP requirements
* RLS requirements
* API security
* Infrastructure security

---

## 3. Backend Rules

docs/standards/backend-rules.md

Defines:

* NestJS standards
* PostgreSQL standards
* Prisma standards
* API standards
* Service architecture
* Observability
* Event systems

---

## 4. Frontend Rules

docs/standards/frontend-rules.md

Defines:

* Next.js standards
* Design system rules
* Component architecture
* Responsive design
* Accessibility
* Performance requirements

---

## 5. Testing Rules

docs/standards/testing-rules.md

Defines:

* Unit testing
* Integration testing
* Playwright verification
* QA account governance (cross-ref: docs/testing/qa-account-governance.md)
* Evidence collection
* Acceptance criteria
* Definition of Done
* Repository cleanliness

---

## 6. AI Rules

docs/standards/ai-rules.md

Defines:

* Truth-first engineering
* Evidence-first development
* Anti-mock policies
* Root cause analysis
* AI verification requirements
* Completion requirements

---

## 7. Documentation Rules

docs/standards/documentation-rules.md

Defines:

* Documentation requirements
* Active context maintenance
* Session continuity
* Architecture documentation
* Change tracking

---

# 7.5 Documentation Folder Rules

docs/standards/documentation-folder-rules.md

Defines:

* Documentation folder structure
* Documentation organization
* Documentation maintenance
* Documentation naming conventions
* Documentation versioning

---

## 8. Operations Rules

docs/standards/operations-rules.md

Defines:

* Deployment standards
* Production readiness
* Monitoring
* Feature flags
* Backup procedures
* Recovery procedures

---

# GK Circle Domain Standards

## 9. Course Rules

docs/standards/course-rules.md

Defines:

* Course architecture
* Test Series architecture
* Current Affairs architecture
* Mentorship architecture
* Course lifecycle
* Course ownership

---

## 10. Live Exam Rules

docs/standards/live-exam-rules.md

Defines:

* Real-time exams
* Live participation
* Test sessions
* Anti-cheating systems
* Leaderboards
* Real-time analytics

---

## 11. RAG Rules

docs/standards/rag-rules.md

Defines:

* Retrieval-Augmented Generation
* Source validation
* Citations
* Knowledge architecture
* Hallucination prevention
* UPSC knowledge systems

---

## 12. Creator Economy Rules

docs/standards/creator-economy-rules.md

Defines:

* Creator onboarding
* Revenue sharing
* Course monetization
* Payout systems
* Mentorship systems
* Creator analytics

---

## 13. Admin Panel Rules

docs/standards/admin-panel-rules.md

Defines:

* Super Admin controls
* Moderation systems
* Audit logs
* Platform management
* Platform analytics
* Administrative workflows

---

## 14. Mobile Rules

docs/standards/mobile-rules.md

Defines:

* React Native standards
* Offline learning
* Mobile performance
* Mobile accessibility
* Mobile notifications
* Mobile-first experiences

---

## 15. DevOps Rules

docs/standards/devops-rules.md

Defines:

* CI/CD
* Infrastructure
* Monitoring
* Cost management
* Backups
* Disaster recovery
* Production operations

---

# Mandatory Context Files

After standards are loaded, read:

CLAUDE.md

docs/Course-vision.md

docs/design-system.md

docs/architecture.md

docs/database-design.md

docs/ai-features.md

docs/project-roadmap.md

docs/feature-roadmap.md

docs/ui-component-library.md

docs/admin-panel.md

docs/monetization-strategy.md

docs/data-analytics-blueprint.md

docs/rag-architecture.md

docs/social-learning.md

docs/live-exams.md

docs/creator-economy.md

docs/mobile-roadmap.md

docs/microservices-roadmap.md

---

# Active Context Files

Always load:

docs/active-context/latest-session.md

docs/active-context/active-plan.md

docs/active-context/active-tasks.md

docs/active-context/blockers.md

docs/active-context/decisions.md

---

# Skills Loading

Before implementation:

Inspect:

.agents/skills/

Load all relevant skills.

Use the most specialized skills available.

General-purpose implementations are discouraged when specialized skills exist.

---

# Mandatory Standards Verification

Before creating an implementation plan, the AI must explicitly verify:

✓ All files referenced in this index have been loaded.

✓ Relevant standards have been identified.

✓ Relevant domain standards have been identified.

Examples:

* Course work → Course-rules.md
* Test Series work → live-exam-rules.md
* AI work → rag-rules.md + ai-rules.md
* Creator work → creator-economy-rules.md
* Admin work → admin-panel-rules.md
* Mobile work → mobile-rules.md
* Infrastructure work → devops-rules.md

Ignoring a relevant standard is a rule violation.

---

# Execution Pipeline

Every task must follow:

Load Standards

↓

Load Context

↓

Load Active Context

↓

Load Relevant Skills

↓

Understand Requirements

↓

Create Implementation Plan

↓

Approval

↓

Implementation

↓

Testing

↓

Security Validation

↓

Playwright Verification

↓

Evidence Collection

↓

Documentation Update

↓

Completion

---

# Enforcement Rule

A task cannot begin until:

✓ Standards Loaded

✓ Context Loaded

✓ Active Context Loaded

✓ Relevant Skills Loaded

✓ Requirements Understood

✓ Plan Created

If any item is missing:

STOP.

Load the missing information first.

---

# Final Directive

Standards are mandatory.

Standards override assumptions.

Standards override convenience.

Standards override shortcuts.

Standards override implementation preferences.

Standards override agent opinions.

Every implementation must comply with all applicable standards.

If a relevant standard exists:

It must be loaded.

It must be followed.

It must be enforced.
