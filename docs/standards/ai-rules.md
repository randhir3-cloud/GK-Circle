# GK Circle AI Rules

Version: 1.1

Status: Mandatory

**v1.1 (2026-07-12):** Root Cause Resolution Chain, Bug Bash Rule, Agent Stop Conditions, strengthened Evidence Rule. Supplements v1.0; AGENTS.md wins on conflict.

---

# Purpose

These rules govern all AI-assisted development inside GK Circle.

The purpose of these rules is to prevent:

* Hallucinations
* Assumptions
* Fake implementations
* Mocked functionality
* Incomplete verification
* Premature completion claims

All AI agents must comply with these standards.

---

# Truth-First Engineering Rule

Truth is more important than speed.

Reality is more important than implementation.

Evidence is more important than assumptions.

Never optimize for appearing successful.

Always optimize for being correct.

---

# CRITICAL ANTI-MOCK RULE

Mock authentication, mock APIs, mock database records, mock email delivery, placeholder implementations, simulated success, and bypass mechanisms are forbidden for feature completion.

Real systems must be implemented, verified, and proven to work end-to-end.

---

# Evidence-Based Engineering Rule

Implementation is not proof.

Configuration is not verification.

Success messages are not confirmation.

Code existence is not evidence.

A feature works only when evidence proves it works.

---

# Never Assume Rule

The AI must NEVER assume functionality works because:

❌ Code exists

❌ Tests exist

❌ Build succeeds

❌ API returns 200

❌ UI shows success

❌ No error is visible

❌ Environment variables exist

❌ Provider configured

❌ Service implemented

❌ Container running

❌ Migration completed

Verification is required.

---

# Root Cause Rule

Fix causes.

Do not hide symptoms.

Do not bypass symptoms.

Do not suppress symptoms.

Do not work around symptoms.

Investigate.

Trace.

Measure.

Verify.

Fix.

Retest.

---

# Root Cause Resolution Chain

Every defect fix must follow this sequence:

```
Root Cause
    ↓
Evidence
    ↓
Minimal Fix
    ↓
Regression Tests
    ↓
Documentation
```

Symptom-only fixes (timeouts, retries, skipped tests, disabled validation) are not acceptable unless the root cause is documented and deterministic alternatives are proven impractical (see `testing-rules.md` §Deterministic Synchronization Rule).

---

# Bug Bash Rule

During **Feature Freeze**, **Bug Bash**, or **RC certification** modes:

**Allowed:**

- Reproduce reported issues
- Identify root cause
- Apply minimal fix
- Add or update tests
- Collect evidence (screenshots, logs, Playwright output)
- Update affected documentation

**Forbidden without explicit approval:**

- New features
- Redesigns
- Refactors unrelated to the defect
- New dependencies
- Parallel implementations

Treat bug bash as verification and stabilization, not scope expansion.

---

# Agent Stop Conditions Rule

Stop immediately when encountering: migration required without approval, architecture conflict with accepted ADR, standards conflict, security conflict, destructive operation without approval, or unclear requirements.

Do not guess. See **`AGENTS.md` §Agent Stop Conditions** for the full table.

---

# No Shortcut Rule

Never solve problems using:

* Hardcoded values
* Fake data
* Placeholder logic
* Disabled validation
* Ignored failures
* Temporary bypasses

If a system is broken:

Repair it.

Do not bypass it.

---

# No Fake Success Rule

Forbidden:

Return success response while operation failed.

Display success message while operation failed.

Skip failing validation.

Ignore failing dependencies.

Mark incomplete work as complete.

---

# Verification Chain Rule

Every feature must follow:

Implementation

↓

Execution

↓

Observation

↓

Verification

↓

Evidence Collection

↓

Completion

---

# Trust Nothing Rule

Until verified:

Assume the feature may be broken.

Verify:

Authentication

Authorization

Database

Email

Payments

OCR

RAG

AI

Analytics

Notifications

Webhooks

Queues

File Uploads

Third-Party Integrations

---

# Completion Rule

The AI may NOT claim:

Fixed

Resolved

Working

Complete

Production Ready

Verified

until evidence exists.

---

# Evidence Rule

Acceptable evidence:

* Typecheck output (pass)
* Build output (pass)
* Test Results
* Playwright Results
* Screenshots
* Database Records
* API Responses
* Logs
* Monitoring Data
* User Workflow Verification

Evidence must be objective.

Never claim **Fixed**, **Done**, **Working**, or **Production Ready** without applicable evidence from the list above.

---

# Playwright Rule

Every feature requires:

Real execution

↓

Real user workflow

↓

Screenshots

↓

Verification

Do not rely solely on:

Unit tests

Integration tests

Manual inspection

---

# AI Self-Verification Rule

Before reporting completion:

Verify your own work.

Do not delegate verification to the user.

The user should not discover issues that automated verification could have found.

---

# AI Feature Rule

AI features must support:

* Error handling
* Rate limiting
* Observability
* Monitoring
* Source attribution
* Security controls

AI output must be treated as untrusted input until validated.

---

# RAG Rule

Retrieval-Augmented Generation systems must:

* Cite sources
* Track source provenance
* Prevent hallucinated references
* Handle missing context gracefully

No RAG response may fabricate sources.

---

# AI Tutor Rule

AI Tutor responses must:

* Explain reasoning
* Cite sources when available
* Handle ambiguity safely
* Avoid fabricated facts

Educational correctness takes priority over response speed.

---

# AI Agent Rule

Before implementation:

Read standards

↓

Read context

↓

Load skills

↓

Create plan

↓

Get approval

↓

Implement

Skipping steps is prohibited.

---

# Scope Control Rule

Implement only the approved scope.

Do not rewrite unrelated systems.

Do not refactor stable systems while fixing isolated defects.

Additional improvements require approval.

---

# Hallucination Prevention Rule

If uncertain:

Investigate.

Search.

Read source code.

Read documentation.

Ask for clarification.

Never invent behavior.

Never invent architecture.

Never invent requirements.

Never invent functionality.

---

# Production Readiness Rule

Temporary code is not complete.

Scaffolding is not complete.

Placeholder logic is not complete.

TODO items are not complete.

Mock implementations are not complete.

Only production-grade implementations may be marked complete.

---

# User Intent Rule

Implement the user's requested solution.

Do not replace it with:

* Easier solutions
* Simpler solutions
* Alternative solutions

without approval.

---

# Assumption Elimination Rule

The AI must never convert assumptions into conclusions.

The following are not evidence:

- code exists
- configuration exists
- service exists
- provider exists
- build succeeded
- tests were written
- success message displayed
- API returned 200
- error not visible

Every claim must be backed by evidence.

Evidence must be observable, reproducible, and verifiable.

If evidence is missing:

The feature is not proven.

Repository Verification Rule

A file may only be declared missing after:

1. Directory verification.
2. File existence verification.
3. Path verification.

Search failure is not evidence of absence.

Failed glob results are not evidence of absence.

The AI must verify the repository structure before declaring any file missing.

# Research Agent Boundary (v0.5 Addendum, 2026-06-05)

A research agent operating inside GK Circle is a recommendation engine and nothing more. This rule supplements, never replaces, the v1.0 rules above. Where this section conflicts with AGENTS.md, AGENTS.md wins. Where this section conflicts with another standards file, the more specific rule wins. If ambiguity exists, document it in an ADR before the research run.

Reference: [ADR 0003](../adrs/0003-autoresearch-read-only-research-subsystem.md), [.agent/skills/autoresearch/SKILL.md](../../.agent/skills/autoresearch/SKILL.md).

A research agent MUST NOT:

- Write, edit, or delete source code under `backend/src/`, `frontend/src/`, `mobile/`, `infrastructure/`
- Modify the Prisma schema or run any migration command (`prisma migrate dev`, `prisma migrate deploy`, `prisma migrate reset`, `prisma db push` are all forbidden)
- Modify any file in `docs/standards/` or any architecture document (`docs/architecture.md`, `docs/database-design.md`, `docs/ai-features.md`, etc.)
- Modify or supersede an existing ADR; a research finding that contradicts an ADR is a finding *about* the ADR, not a change to it. To change a decision, a new ADR that supersedes the old one must be written.
- Execute `git commit`, `git push`, or open a pull request
- Deploy, restart, or scale any environment
- Bypass authentication, authorization, RLS, secret rotation, or any other security control
- Substitute parametric knowledge for retrieved, cited evidence
- Invent citations, fabricate sources, or report a finding as "verified" without evidence
- Edit `node_modules`, `.env*`, `docker-compose*.yml`, or any deployment or infrastructure configuration

A research agent MUST:

- Cite every factual claim with a URL or a `file:line` reference
- Run a Faithfulness Self-Check (per [rag-rules.md](rag-rules.md) §Faithfulness Verifier pattern, L230) before emitting a report
- Refuse to act on its own findings; emit a HANDOFF-TO-HUMAN block instead
- Validate every recommendation against the relevant standard files; if a recommendation violates a standard, REJECT the recommendation and report the violation
- When unable to verify a claim, mark it `[UNSOURCED — confidence: low]`, never state it as fact
- Operate only with read tools: `Read`, `Grep`, `Glob`, `WebSearch`, `WebFetch`, `Skill`, `Agent` (read-only subagent types: `Explore`, `Plan`, `claude-code-guide`). `Bash` is permitted only for read-only commands (`ls`, `cat`, `grep`, `git log`, `git diff`, `git status`); never `git commit`, `git push`, `npm install`, `prisma migrate`, `docker`, `kubectl`, `terraform`, `aws`, `gcloud`, or any state-mutating CLI
- Produce a `Recommended Actions` section that is HUMAN-ONLY — no auto-execution, no script that would mutate state

The Truth-First, Anti-Mock, Evidence-Based, Never Assume, Hallucination Prevention, Completion, Scope Control, and Assumption Elimination rules in this file all apply to research agents. A research agent that fabricates a citation is a violation of the Anti-Mock Rule. A research agent that reports a finding as "verified" without a retrievable source is a violation of the Evidence-Based Engineering Rule and the Never Assume Rule.

# Final Directive

Never replace evidence with confidence.

Never replace verification with assumptions.

Never replace implementation with proof.

Never replace reality with optimism.

If evidence does not exist:

The feature is not proven.

If proof does not exist:

The task is not complete.
