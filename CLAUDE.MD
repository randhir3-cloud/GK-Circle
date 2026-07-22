# CLAUDE.md

# GK Circle Product Context

Version: 3.1

This document provides long-term product context for all AI coding agents working on GK Circle.

Read this file before generating any code.

**Governance:** Start with **`AGENTS.md`** §GK Circle AI Maintainer Constitution and §Implementation Gate. Standards history: **`docs/standards/CHANGELOG.md`**.

---

Documentation rules are defined in:

docs/standards/documentation-governance.md

Agents must follow this standard when generating reports, evidence, plans, audits, architecture documents, and verification artifacts.

# GK Circle Enterprise Master Agent Rules (from AGENTS.md)

## CRITICAL TRUTH RULE

Never bypass an error using dummy data, fake responses, hardcoded values, disabled validation, skipped tests, mocked success, or placeholder implementations. Always identify, fix, verify, and retest the real root cause before proceeding.

## CRITICAL ANTI-MOCK RULE

Mock authentication, mock APIs, mock database records, mock email delivery, placeholder implementations, simulated success, and bypass mechanisms are forbidden for feature completion. Real systems must be implemented, verified, and proven to work end-to-end.

## CRITICAL EVIDENCE RULE

Do not assume functionality from implementation, configuration, success messages, logs, code existence, API responses, environment variables, providers, services, completed tasks, successful builds, or passing tests. Verification requires evidence.

## CRITICAL COMPLETION RULE

A task is not complete because code was written. A task is complete only when:

- Functionality works end-to-end
- Typecheck passes
- Build passes
- Tests pass
- Playwright passes (when UI/workflow affected)
- Manual verification complete (when applicable)
- Screenshots and evidence exist (when required)
- Security checks pass
- Documentation updated (when applicable)
- Acceptance criteria are satisfied

# Account Rule

Exactly **four** immutable QA identities exist (Student, Creator, Admin, Super Admin). Agents must never create duplicate QA accounts, timestamp users, random emails, or temporary testing users outside registration specs.

Authoritative detail: **`docs/testing/qa-account-governance.md`**

Approved credential helpers only: `getQaUser()`, `loginAs()`, `loginQaViaApi()`, `verify-qa-accounts.ts`, `cleanup-ephemeral-users.ts`. Additional QA users require explicit approval before creation.

## CRITICAL SECURITY RULE

Security is mandatory. No feature may bypass authentication, authorization, validation, access control, secret management, security auditing, or production security standards.

## CRITICAL NO-DESTRUCTIVE-CHANGES RULE

Never delete, reset, overwrite, truncate, drop, or destructively migrate data, APIs, schemas, or infrastructure without explicit approval and rollback planning.

---

# Mandatory Standards Loading

Before ANY task, read `docs/standards/index.md` and load EVERY standard referenced by the index, including Core Governance Standards (architecture, security, backend, frontend, testing, ai, documentation, operations) and Domain Standards (product, live-exam, rag, creator-economy, admin-panel, mobile, devops). Identify which standards are relevant and enforce them. No work may begin until standards are loaded. Ignoring a relevant standard is considered a rule violation.

---

# Mandatory Context Loading

Before ANY task read:

- CLAUDE.md
- docs/product-vision.md
- docs/design-system.md
- docs/architecture.md
- docs/database-design.md
- docs/ai-features.md
- docs/active-context/latest-session.md
- docs/active-context/active-plan.md
- docs/active-context/active-tasks.md

---

# Required Execution Workflow

1. Load Standards
2. Load Project Context
3. Load Active Context
4. Understand Requirements
5. Create Implementation Plan
6. Obtain Approval
7. Implement
8. Test
9. Run Security Validation
10. Run Playwright Verification
11. Capture Evidence
12. Update Documentation
13. Complete

Skipping steps is prohibited.

---

# Skills System

Before implementation, inspect `.agents/skills/`. Load all relevant skills. Use the most specialized skills available.

---

# Documentation First & Rule Hierarchy

Every significant feature requires Architecture, Specification, Implementation Plan, Testing Strategy, and Security Review before implementation.

Priority Order:

1. AGENTS.md
2. docs/standards/\*
3. Architecture Documents
4. Feature Specifications
5. Implementation Plans
6. User Request
   If conflicts occur, higher priority rules win.

---

# Repository Constitution

All agents must comply with the **GK Circle AI Maintainer Constitution** and **Implementation Gate** in **`AGENTS.md`**. Cross-reference `docs/standards/index.md`, `docs/standards/CHANGELOG.md`, and `docs/testing/qa-account-governance.md` for task-specific rules.

---

# What is GK Circle?

GK Circle is an AI-powered social learning ecosystem designed for competitive exam preparation and lifelong learning.

The platform combines:

- AI Tutor
- Social Learning
- Community Learning
- MCQ Practice
- Live Tests
- Real-Time Quiz Battles
- Personalized Learning
- Learning Analytics
- Creator Economy
- Knowledge Management

The goal is to create the world's most intelligent learning platform.

---

# Long-Term Vision

Most education platforms only deliver content.

GK Circle helps users:

Learn

Practice

Revise

Discuss

Compete

Teach

Create

Earn

Improve

The platform should evolve into an educational operating system.

---

# Core User Personas

## Student

Preparing for:

- UPSC
- SSC
- State PCS
- Banking
- Railways
- Defence
- Teaching Exams
- University Exams

Goals:

- Learn faster
- Improve scores
- Track progress
- Get AI guidance

---

## Aspirant

Needs:

- Daily practice
- Current affairs
- Revision plans
- Test series
- Performance analytics

---

## Teacher

Needs:

- Create courses
- Create tests
- Create notes
- Build communities

---

## Mentor

Needs:

- Manage learners
- Track student progress
- Conduct live sessions

---

## Content Creator

Needs:

- Publish content
- Build audience
- Earn revenue
- Sell courses

---

## Admin

Needs:

- Platform control
- Moderation
- Analytics
- Compliance

---

# Product Pillars

## Pillar 1

AI First Learning

Every learning workflow should be AI-enhanced.

---

## Pillar 2

Social Learning

Students learn better together.

---

## Pillar 3

Gamification

Learning should be rewarding.

---

## Pillar 4

Personalization

Every learner should have a unique journey.

---

## Pillar 5

Data Driven Improvement

Measure everything.

Improve continuously.

---

# Core Modules

## Learning Module

Features:

Courses

Lessons

Notes

Mind Maps

Flashcards

Bookmarks

Revision

---

## Testing Module

Features:

MCQs

Mock Tests

Previous Papers

Quiz Battles

Live Exams

Rankings

---

## AI Module

Features:

AI Tutor

AI Coach

AI Notes Generator

AI Quiz Generator

AI Revision Planner

AI Interview Coach

AI Doubt Solver

AI MCQ Analyzer

---

## Community Module

Features:

Posts

Comments

Groups

Study Rooms

Voice Channels

Discussion Threads

Mentorship

---

## Analytics Module

Features:

Progress Tracking

Skill Maps

Weak Areas

Strength Areas

Revision Analytics

Performance Reports

Predictions

---

## Creator Module

Features:

Course Creation

Test Creation

Content Publishing

Subscriptions

Monetization

Revenue Dashboard

---

# North Star Metrics

Track:

Daily Active Learners

Weekly Active Learners

Questions Solved

Tests Completed

Study Time

Community Engagement

Learning Retention

AI Usage

Creator Revenue

---

# Mobile App Strategy

Primary platform:

Mobile

Secondary:

Web

Design decisions must prioritize mobile users.

---

# AI Roadmap

Phase 1

AI Tutor

AI MCQ Explanation

AI Doubt Solver

---

Phase 2

AI Study Planner

AI Revision Coach

AI Interview Coach

---

Phase 3

Multi-Agent Learning System

Personal Learning Agents

Exam-Specific Agents

Research Agents

---

# Gamification Roadmap

XP

Levels

Badges

Achievements

Leaderboards

Streaks

Challenges

Seasonal Events

Learning Quests

---

# Community Roadmap

Groups

Study Circles

Voice Rooms

Mentorship

Live Events

Peer Learning

Collaborative Notes

---

# Creator Economy Roadmap

Paid Courses

Premium Communities

Subscriptions

Affiliate Programs

Revenue Sharing

Creator Marketplace

---

# Future AI Architecture

RAG System

Vector Search

Knowledge Graph

Multi-Agent Workflows

Learning Memory

Personalized Recommendations

Learning Analytics Engine

---

# Design Philosophy

The platform should feel like:

Duolingo
Reddit
Discord
LinkedIn
Notion
ChatGPT

combined into one ecosystem.

Avoid:

Generic LMS

Generic Coaching Website

Generic SaaS Dashboard

---

# Product Decision Framework

When building features ask:

1. Does it improve learning?
2. Does it increase engagement?
3. Does it improve retention?
4. Does AI make it better?
5. Does it support personalization?
6. Does it support community learning?

If not, reconsider the feature.

---

# Final Directive

Do not think like a website developer.

Think like the team building the future of education.

Do not build pages. Build systems.
Do not build features. Build reusable platforms.

Before writing any code:

1. Read ALL documentation listed at the top of this file
2. Read the relevant feature specification in docs/features/
3. Understand the database schema in backend/prisma/schema.prisma
4. Create an implementation plan
5. Get approval before proceeding

Do not stop after code generation. Continue until:

✓ Documentation read completely
✓ Feature specification followed
✓ Requirements satisfied
✓ Tests pass
✓ Screenshots verified
✓ Security audit passes
✓ No known errors remain

If evidence does not exist, the feature is not yet proven to work.
If proof does not exist, the task is not complete.

Every line of code should contribute to:

Smarter learning.

Better outcomes.

Stronger communities.

More personalized education.

Documentation-first. Always.
