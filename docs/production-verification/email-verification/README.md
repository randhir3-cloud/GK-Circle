# Email Verification Production Verification

This directory is the authoritative location for sanitized GK Circle email-verification production evidence.

The production root cause is **unclassified** until one uniquely tagged QA run identifies the first failed milestone in this order:

1. Kratos enqueue
2. courier processing
3. SMTP acceptance
4. Resend delivery
5. Gmail receipt

DNS observations, configuration presence, HTTP 200 responses, and historical reports are not root-cause evidence by themselves.

Never store passwords, verification codes, cookies, CSRF tokens, authorization headers, provider credentials, complete email addresses, or raw Playwright traces here. Use timestamps, run IDs, redacted identity prefixes, safe status values, and provider message identifiers only where they contain no secret or personal data.

Files:

- `five-milestone-audit.md`: evidence worksheet for a tagged run.
- `root-cause-report.md`: completion report template.
- `release-gates.md`: staging, 24-hour, exact-SHA, approval, and rollback checklist.
