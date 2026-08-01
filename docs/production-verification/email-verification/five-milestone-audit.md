# Five-Milestone Audit Worksheet

Run ID: `<unique-run-id>`

Commit SHA: `<40-character-sha>`

Started at UTC: `<timestamp>`

| Milestone | Result | Timestamp | Sanitized evidence reference |
|---|---|---|---|
| Kratos enqueue | NOT_RUN | | |
| Courier processing | NOT_RUN | | |
| SMTP acceptance | NOT_RUN | | |
| Resend delivery | NOT_RUN | | |
| Gmail receipt | NOT_RUN | | |

The first milestone marked `FAILED` classifies the failure boundary. Do not infer later milestone results after a failure.

Additional checks:

- Identity has a verifiable email address through Kratos Admin API.
- Verification changes that address to verified.
- Session refresh preserves authentication.
- Protected route is denied before verification and accessible afterward.
- Resend throttling prevents duplicate sends.
- Test identity is removed through the internal QA cleanup command.
- No redirect exposes localhost, loopback, or Railway internal hostnames.
