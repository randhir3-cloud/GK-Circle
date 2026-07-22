STANDARD-ID: DOC-001

Title:
Documentation Governance Standard

Status:
ACTIVE

Effective Date:
Immediately

Purpose:
Maintain a scalable and organized documentation structure.

Rules:

1. Repository Root Policy
   Allowed markdown files:
   - README.md
   - AGENTS.md
   - CLAUDE.md

   No other markdown files allowed.

2. Documentation Placement Policy
   All documentation must reside under docs/.

3. Feature Ownership Policy
   Every document belongs to exactly one feature.

4. No Duplicate Documentation Policy
   A report may only have one authoritative location.

5. Evidence Storage Policy
   Evidence must be stored under:
   docs/features/<feature>/evidence/

6. Production Verification Policy
   Production verification reports must be stored under:
   docs/production-verification/<feature>/

7. Architecture Policy
   Architecture documents belong under:
   docs/architecture/<feature>/

8. Audit Policy
   Audits belong under:
   docs/reports/audits/

Compliance:
Violations are considered repository defects.

---

## Evidence Folder Convention (Bug Bash / Release)

Never scatter verification evidence across the repository root, `screenshots/`, or random feature folders.

Use this layout:

```
docs/
  bug-bash/
    <YYYY-MM-DD-or-sprint-id>/
      reports/
      screenshots/
      traces/
      logs/
  releases/
    <release-id>/
      reports/
      screenshots/
      traces/
      logs/
  evidence/
    <initiative>/
      screenshots/
      traces/
      logs/
      reports/
```

Rules:

- Feature-specific evidence **also** belongs under `docs/features/<feature>/evidence/` per §Evidence Storage Policy
- Production verification reports: `docs/production-verification/<feature>/`
- Do not commit Playwright MCP debug dumps; use `.gitignore` for tool clutter
- Link evidence paths in completion reports and `docs/standards/CHANGELOG.md` when governance changes

Cross-reference: `testing-rules.md` §Production Verification Rule, `testing-rules.md` §Screenshot Evidence Rule.

---


RULE AC-001

Every new markdown file created during Phase 3.1.3B must be stored under:

docs/features/anti-cheating/

Allowed:
docs/features/anti-cheating/implementation/
docs/features/anti-cheating/evidence/
docs/features/anti-cheating/reports/

Forbidden:
repository root
backend root
frontend root