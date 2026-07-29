# GK Circle Transactional Email Operations Manual

This document provides operational and deployment runbooks for the GK Circle Transactional Email system.

---

## 1. Local Development Configuration (Mailpit)

In local development, all emails (both Ory Kratos authentication and Go API transactional) are routed to a local Mailpit SMTP relay to avoid costs and prevent spamming external mailboxes.

### Configuration (`.env` or local system environment)
```bash
# Go Transactional Email Configuration
EMAIL_PROVIDER=smtp
EMAIL_FROM_NAME="GK Circle Dev"
EMAIL_FROM=notifications@gkcircle.local
SMTP_HOST=127.0.0.1
SMTP_PORT=1025
SMTP_USER=""
SMTP_PASSWORD=""
SMTP_FROM=""
SMTP_DISABLE_STARTTLS=true
SMTP_INSECURE_SKIP_VERIFY=true

# Ory Kratos Courier SMTP Configuration
COURIER_SMTP_CONNECTION_URI=smtp://127.0.0.1:1025/?disable_starttls=true
COURIER_SMTP_FROM_ADDRESS=account@gkcircle.local
```

---

## 2. Production Deployment (NUC Server & Resend)

For production, Ory Kratos uses Resend SMTP relay on port 465, and the Go API uses the Resend HTTP REST API. Use separate API keys for isolation.

### Configuration Checklist (`/home/randhir/apps/gk-circle-v2/.env`)

1. **Verify DNS Status**:
   - Ensure the sending domain (e.g. `mail.gkcircle.com`) has passed verification in the Resend dashboard.
   - Configure SPF, DKIM, and DMARC TXT records in Cloudflare.
2. **Kratos Configuration**:
   ```bash
   COURIER_SMTP_CONNECTION_URI=smtps://resend:<KRATOS_RESEND_KEY>@smtp.resend.com:465/
   COURIER_SMTP_FROM_ADDRESS=account@mail.gkcircle.com
   ```
3. **Go API Transactional Configuration**:
   ```bash
   EMAIL_PROVIDER=resend
   EMAIL_FROM_NAME="GK Circle"
   EMAIL_FROM=notifications@mail.gkcircle.com
   EMAIL_REPLY_TO=support@gkcircle.com
   RESEND_API_KEY=<TRANSACTIONAL_RESEND_KEY>
   EMAIL_HTTP_TIMEOUT=10s
   EMAIL_MAX_ATTEMPTS=3
   EMAIL_RETRY_BASE_DELAY=250ms
   EMAIL_RETRY_MAX_DELAY=2s
   ```

---

## 3. Post-Deployment Verification Runbook

After deploying to the NUC server, perform the following validation steps:

1. **Check Service Status**:
   ```bash
   docker compose -f docker-compose.nuc.yml ps
   ```
2. **Tail Application Logs**:
   - Monitor Kratos courier logs:
     ```bash
     docker compose -f docker-compose.nuc.yml logs --tail=200 -f kratos
     ```
   - Monitor Go API logs:
     ```bash
     docker compose -f docker-compose.nuc.yml logs --tail=200 -f api
     ```
3. **Test Deliverability**:
   - **Authentication Email**: Register a new user with an external Gmail/Outlook address and trigger a verification email.
   - **Transactional Email**: Use the frontend or call the `/v1/shared_quizzes/:quiz_id` endpoint to invite a collaborator, sending the invitation email to an external account.
4. **Header Validation**:
   - Open the received email in Gmail or Outlook, view the raw source, and verify:
     - `SPF: PASS`
     - `DKIM: PASS`
     - `DMARC: PASS`
5. **Log Integrity check**:
   - Confirm that no raw email addresses or API keys are written to application output. Confirm that the logged idempotency key hash is truncated.
