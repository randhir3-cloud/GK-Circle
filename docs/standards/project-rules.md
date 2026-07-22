## DOC-002 — Documentation Location Governance

All markdown documentation generated during development, implementation, verification, auditing, deployment, architecture review, testing, or production verification must be stored under the `docs/` directory.

### Approved Locations

* docs/features/<feature-name>/
* docs/features/<feature-name>/implementation/
* docs/features/<feature-name>/evidence/
* docs/features/<feature-name>/reports/
* docs/architecture/
* docs/active-context/
* docs/standards/
* docs/reports/

### Forbidden Locations

Markdown files must NOT be created in:

* Repository root
* backend/
* frontend/
* scripts/
* test/
* screenshots/

Exceptions:

* README.md
* AGENTS.md
* CLAUDE.md

### Enforcement

Before creating any new markdown file, the author must determine the owning feature and place the document within the corresponding docs subtree.

Example:

Phase 3.1.3B Anti-Cheating:

docs/features/anti-cheating/implementation/
docs/features/anti-cheating/evidence/
docs/features/anti-cheating/reports/
