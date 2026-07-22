# GK Circle Documentation Rules

Version: 1.1

Status: Mandatory

**v1.1 (2026-07-12):** Task Completion Documentation Rule. Supplements v1.0; AGENTS.md wins on conflict.

---

# Purpose

These rules govern all documentation, knowledge management, architecture records, implementation history, active context tracking, and project continuity across GK Circle.

Documentation is a production asset.

Documentation must evolve with the codebase.

---

# Documentation First Rule

Before major implementation:

Create documentation first.

Required:

Problem

↓

Requirements

↓

Architecture

↓

Data Model

↓

API Design

↓

Testing Strategy

↓

Implementation Plan

Only then begin coding.

---

# Documentation Is Code Rule

Documentation is part of the platform.

Documentation must be:

Accurate

Current

Versioned

Maintained

Outdated documentation is a bug.

---

# Single Source Of Truth Rule

Each topic must have one authoritative document.

Examples:

Architecture

↓

docs/architecture.md

Design System

↓

docs/design-system.md

Database

↓

docs/database-design.md

Avoid duplicate documentation.

---

# Active Context Rule

The project must maintain active context.

Directory:

docs/active-context/

Purpose:

Allow future agents to continue work without losing knowledge.

---

# Required Active Context Files

Maintain:

docs/active-context/latest-session.md

docs/active-context/active-plan.md

docs/active-context/active-tasks.md

docs/active-context/blockers.md

docs/active-context/decisions.md

---

# Latest Session Rule

latest-session.md must contain:

Date

Objectives

Completed Work

Current State

Open Issues

Next Steps

This file should allow another engineer to continue immediately.

---

# Active Plan Rule

active-plan.md contains:

Current Goal

Current Phase

Implementation Strategy

Dependencies

Milestones

Expected Outcomes

Update when plans change.

---

# Active Tasks Rule

active-tasks.md contains:

Pending Tasks

In Progress Tasks

Completed Tasks

Blocked Tasks

Task owners if applicable.

---

# Blockers Rule

blockers.md contains:

Issue

Impact

Root Cause

Current Status

Resolution Strategy

Never hide blockers.

---

# Task Completion Documentation Rule

Every completed task must update documentation **where applicable**. Update only documents affected by the change — do not perform unrelated doc churn.

Examples of documents that may require updates:

- `docs/active-context/active-tasks.md` and `latest-session.md`
- Feature specs, architecture notes, and testing docs under `docs/features/`
- `docs/testing/` governance and certification artifacts
- Release notes, CHANGELOG, or deployment checklists when shipping
- Standards or ADRs when governance or architecture decisions change

Documentation updates are part of the Definition of Done (`testing-rules.md`). A task marked complete with outdated or missing docs is not complete.

---

# Decisions Rule

decisions.md contains:

Decision

Reason

Alternatives Considered

Tradeoffs

Date

Owner

This becomes the project memory.

---

# Architecture Decision Record Rule

Major architectural decisions require ADRs.

Location:

docs/adrs/

Format:

Problem

↓

Options

↓

Tradeoffs

↓

Decision

↓

Consequences

---

# Feature Specification Rule

Every major feature requires:

docs/features/<feature-name>.md

Include:

Overview

Goals

Requirements

User Flows

Data Models

API Contracts

Testing Strategy

Acceptance Criteria

---

# Change Log Rule

Maintain:

CHANGELOG.md

Track:

Added

Changed

Fixed

Removed

Deprecated

---

# Migration Documentation Rule

Database changes require documentation.

Document:

Schema Changes

Migration Strategy

Rollback Plan

Data Impact

---

# API Documentation Rule

Every API must document:

Purpose

Request

Response

Authentication

Permissions

Validation

Errors

Examples

---

# Integration Documentation Rule

Third-party integrations require documentation.

Examples:

Payments

Email

AI Providers

Storage

Analytics

Document:

Purpose

Credentials

Configuration

Failure Modes

Recovery Procedures

---

# Deployment Documentation Rule

Document:

Deployment Process

Rollback Process

Environment Variables

Infrastructure Requirements

Monitoring

Recovery

---

# Security Documentation Rule

Document:

Authentication

Authorization

Secrets Management

RLS Policies

Rate Limits

Security Controls

Audit Procedures

---

# Testing Documentation Rule

Document:

Testing Strategy

Test Coverage

Playwright Workflows

Evidence Requirements

Definition of Done

---

# AI Documentation Rule

Document:

Prompts

RAG Architecture

Agent Behaviors

AI Workflows

Model Selection

Fallback Logic

Cost Controls

Safety Controls

---

# Knowledge Preservation Rule

Never allow knowledge to exist only in:

* Chat conversations
* Slack messages
* Memory
* Verbal discussions

Important knowledge must be documented.

---

# Session Handoff Rule

Before ending significant work:

Update:

latest-session.md

active-plan.md

active-tasks.md

decisions.md

This allows seamless continuation.

---

# Documentation Review Rule

When implementation changes:

Review documentation.

If documentation becomes inaccurate:

Update it immediately.

---

# Documentation Completion Rule

A feature is not complete until:

✓ Implementation Complete

✓ Tests Pass

✓ Documentation Updated

✓ Active Context Updated

✓ Decisions Recorded

---

# AI Context Loading Rule

Before implementation:

Read:

AGENTS.md

↓

docs/standards/index.md

↓

All standards

↓

Active Context Files

↓

Relevant Feature Documents

Skipping documentation review is prohibited.

---

# Decision Capture Rule

Whenever a non-trivial technical decision is made:

Create or update:

docs/active-context/decisions.md

Include:

- decision
- reason
- alternatives considered
- tradeoffs
- date

Future agents should understand why a decision was made, not only what was implemented.

# Final Directive

Code without documentation creates future failures.

Documentation without maintenance becomes misinformation.

Maintain documentation with the same discipline as code.
