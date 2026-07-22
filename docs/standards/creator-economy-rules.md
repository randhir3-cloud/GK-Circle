# GK Circle Creator Economy Rules

Version: 1.0

Status: Mandatory

---

# Purpose

Govern creators, educators, mentors, institutions, and publishers.

---

# Creator Types

Educator

Mentor

Institution

Publisher

Coach

Content Creator

---

# Monetization Models

Course Sales

Subscriptions

Memberships

Mentorship

Live Sessions

Revenue Sharing

Affiliate Programs

---

# Revenue Rule

Revenue must be traceable.

Every payout requires:

Revenue Source

↓

Order

↓

Settlement

↓

Audit Trail

---

# Creator Dashboard Rule

Provide:

Revenue

Courses

Students

Performance

Analytics

Payouts

---

# Trust Rule

Prevent:

Fake Reviews

Fake Students

Fake Enrollments

Artificial Ratings

---

# Completion Rule

Creator systems must support:

Creation

Sales

Analytics

Payouts

Auditability

---

# v0.5 Addendum — Architecture-Intent Standard

The v1.0 section above defines the principles (Creator Types, Monetization Models, Revenue traceability, Creator Dashboard, Trust). This addendum makes those principles enforceable as code, configuration, and process. The structure mirrors the other v0.5 standards.

This is **architecture-intent**: the creator economy is not yet fully implemented. The rules below are the contract it MUST satisfy when built. The next revision (v1.0 of this file) will merge these rules into the body.

The creator economy is the second-highest-trust surface in GK Circle (after the admin panel). A creator-economy compromise is a financial compromise — creators lose money, learners lose trust, and the platform's reputation is damaged. The standards below are tuned for that risk profile.

---

# Dependency Rule (Mandatory)

This document does not replace AGENTS.md.

It supplements AGENTS.md.

If this document conflicts with AGENTS.md, AGENTS.md wins.

If this document conflicts with another standards file, the more specific standard wins.

If ambiguity exists, document the ambiguity in an ADR before implementation.

# Verification Requirement (Mandatory)

A rule is not considered satisfied because code, configuration, tests, or documentation exist. A rule is satisfied only when evidence exists. For the creator economy, evidence must include:

- A double-entry ledger reconciliation test demonstrating that every transaction balances
- A payout test demonstrating that a creator can withdraw funds end-to-end
- A refund test demonstrating that a refund correctly reverses the revenue split
- A penetration test confirming that a creator cannot forge a sale or inflate their metrics

---

# Revenue Ownership (Mandatory)

The cardinal rule: every rupee of revenue has exactly one owner at any point in time. The system MUST be able to answer "who owns this money?" for any transaction, at any point in its lifecycle.

## Revenue Flow (Mandatory)

```
Customer Payment
    ↓
Platform Receives Funds
    ↓
Order Created
    ↓
Revenue Allocated (split: creator share, platform fee, taxes)
    ↓
Funds Held in Platform Account
    ↓
Creator Balance Updated (Available, Pending, Reserved)
    ↓
[On Payout Cycle]
Creator Requests Payout
    ↓
Payout Approved
    ↓
Funds Transferred to Creator
    ↓
Payout Reconciled with Bank
```

Each step is a database transaction. Each step is logged. There is no step where money "disappears."

## Double-Entry Ledger (Mandatory)

The creator economy uses a **double-entry ledger**. Every transaction creates at least two ledger entries (debit one account, credit another). The sum of all entries is always zero.

### Account Types

The ledger has the following account types:

| Account Type | Owner | Purpose |
|---|---|---|
| `PLATFORM_REVENUE` | Platform | Platform's share of revenue |
| `CREATOR_REVENUE_<creatorId>` | Creator | Creator's share of revenue (pending) |
| `CREATOR_AVAILABLE_<creatorId>` | Creator | Creator's available-for-payout balance |
| `CUSTOMER_CREDIT` | Customer | Customer's account credit (for refunds, gift cards) |
| `TAX_PAYABLE` | Platform | GST and other taxes collected |
| `PROCESSING_FEE` | Platform | Payment gateway fees |
| `REFUND_LIABILITY` | Platform | Funds held for pending refunds |
| `CHARGEBACK_LIABILITY` | Platform | Funds held for chargebacks |
| `PLATFORM_CASH` | Platform | Platform's cash account |
| `CREATOR_CASH_<creatorId>` | Creator | Creator's paid-out balance |

### Ledger Entry

```typescript
interface LedgerEntry {
  id: string;
  txnId: string;
  account: string;
  entryType: 'DEBIT' | 'CREDIT';
  amount: number;
  currency: 'INR';
  referenceType: 'ORDER' | 'REFUND' | 'PAYOUT' | 'ADJUSTMENT' | 'CHARGEBACK' | 'FEE';
  referenceId: string;
  description: string;
  metadata: Record<string, any>;
  createdAt: string;
  createdBy: string;
  reversedBy?: string;
  reversalOf?: string;
}
```

### Ledger Invariants (Mandatory)

The ledger is checked for the following invariants daily:

1. Conservation: SUM(all DEBITs) == SUM(all CREDITs)
2. No negative balances (except in allowed types)
3. Referential integrity
4. No orphan reversals
5. No double-reversal
6. Amount integrity (positive integer paise, no floats, no zeros unless explicitly a fee waiver)

A violation of any invariant pages the on-call.

## No Direct Manipulation of Payouts (Mandatory)

The cardinal financial rule: **no creator, no admin, and no service may directly set, edit, or override a creator's payout amount, available balance, pending balance, or settlement record.** Payouts and balances are DERIVED. They are computed from the underlying source-of-truth records:

```
creator_balance_available  = SUM(ledger.credits - ledger.debits) WHERE account LIKE 'CREATOR_AVAILABLE_%'
creator_balance_pending    = SUM(ledger.credits - ledger.debits) WHERE account LIKE 'CREATOR_REVENUE_%'
creator_payout_amount      = creator_balance_available AT payout_request_time - outstanding_chargebacks - outstanding_negative_balances
```

Every figure shown to a creator in their dashboard, in a payout notification, or in a tax form is a computed projection of the ledger. There is no "payout override" field on the creator model. There is no admin endpoint that accepts a payout amount as input. An admin action that needs to adjust a balance is recorded as an `ADJUSTMENT` ledger entry (debit one account, credit another) with a required reason and an `AuditLog` entry. The adjustment appears in the ledger and in the creator's history like any other transaction.

The only "writes" allowed against the ledger are:

1. Order → creates revenue allocation entries
2. Refund → creates reversal entries
3. Chargeback → creates reversal entries
4. Payout → creates debit against the creator's available balance and credit against the creator's cash account
5. Tax event → creates a tax payable entry
6. Admin adjustment → creates a paired debit/credit entry with a reason

Any code path that attempts to write to the ledger outside these six is a security review incident.

## Revenue Split (Mandatory)

The platform supports configurable revenue splits. Default:

| Party | Share |
|---|---|
| Creator | 70% |
| Platform | 20% |
| Taxes (GST) | 10% (passed through) |

Enterprise creators may negotiate different splits (e.g., 80/10/10). The split is recorded in the `CreatorRevenueAgreement`.

The split is calculated at the time of order completion, not at payout. A retroactive change to the split does NOT affect already-allocated revenue.

## Settlement (Mandatory)

Settlement is the process of moving funds from "pending" to "available" for payout. The default settlement period is **7 days** after order completion (to allow for refunds and chargebacks). Verified creators may request a 3-day settlement; enterprise creators may request a 1-day (or instant, with a fee) settlement.

Settlement is a daily cron job that:

1. Identifies all orders completed ≥ 7 days ago (or ≥ 3 days, etc.) that are not yet settled
2. Moves the creator's share from `CREATOR_REVENUE_<id>` to `CREATOR_AVAILABLE_<id>`
3. Creates a settlement record

The settlement is idempotent.

## Refund Impact on Revenue (Mandatory)

A refund reverses the original revenue allocation:

- Creator's share is debited from the creator's available balance
- Platform's share is debited from the platform's revenue
- Taxes are reversed (a credit note is issued)
- Processing fees are NOT reversed (the payment gateway keeps them)

If the creator does not have enough available balance, the platform covers the refund and creates a `NEGATIVE_BALANCE` on the creator. The negative balance is recovered from the creator's future earnings.

## Chargeback Impact (Mandatory)

A chargeback is treated like a refund, with additional consequences:

- The order is marked `CHARGEBACKED`
- The creator's `chargebackRate` is updated
- If `chargebackRate` exceeds 1% (over 30 days), payouts are paused and the creator is reviewed
- The creator is notified of every chargeback
- The creator can dispute a chargeback via the support flow

---

# Enrollment Ownership (Mandatory)

When a customer buys a Course, they receive an `Enrollment`. The enrollment is a record of the customer's right to access the Course's content.

## Enrollment Lifecycle (Mandatory)

```
PENDING → ACTIVE → [CANCELLED | EXPIRED | REFUNDED | REVOKED]
```

| State | Description |
|---|---|
| `PENDING` | Payment is in progress; enrollment is being created |
| `ACTIVE` | Customer has access to the Course |
| `CANCELLED` | Subscription cancelled; access continues until period end |
| `EXPIRED` | Subscription period ended; access revoked |
| `REFUNDED` | Enrollment refunded; access revoked immediately |
| `REVOKED` | Access revoked by admin (e.g., Course taken down) |

## Enrollment Rules (Mandatory)

- An enrollment is bound to exactly one user and exactly one Course
- A user can have at most one `ACTIVE` enrollment per Course (re-purchase is forbidden; renewal is allowed)
- A refund transitions the enrollment to `REFUNDED` and revokes access immediately
- A `REVOKED` enrollment is not eligible for refund

## Bundle Enrollments (Mandatory)

A Course can be a "bundle" containing multiple sub-Courses. An enrollment in the bundle implies enrollment in all sub-Courses. A refund of the bundle is a refund of all sub-Courses.

## Cohort Enrollments (Mandatory)

A cohort Course (e.g., a 12-week live Course) has a `cohortId` and a `cohortStartDate`. The enrollment is valid only for the cohort period. After the cohort ends, the enrollment transitions to `EXPIRED` (with a 30-day grace period to view recordings).

## Gift Enrollments (Mandatory)

A user MAY gift an enrollment. The gifted enrollment is owned by the recipient, but the original purchase is attributed to the gifter (for revenue share and analytics). The recipient receives an email with a redemption code.

---

# Refunds (Mandatory)

Refunds are a customer right. The system MUST support refunds that reverse the revenue allocation correctly.

## Refund Eligibility (Mandatory)

| Reason | Eligible Within | Notes |
|---|---|---|
| `NOT_AS_DESCRIBED` | 30 days | Course does not match its description |
| `TECHNICAL_ISSUE` | 30 days | Course has a technical issue that the platform could not resolve |
| `DUPLICATE_PURCHASE` | 7 days | Customer accidentally bought twice |
| `UNAUTHORIZED` | 60 days | Customer did not authorize the purchase |
| `OTHER` | 14 days | Case-by-case; admin reviews |

## Refund Initiation (Mandatory)

A refund is initiated by:

1. The user (self-service) via the order page
2. Customer support (assisted) via the admin panel
3. A payment gateway (gateway-initiated, e.g., dispute resolution)
4. The platform (auto-refund, e.g., for a Course taken down)

## Refund Approval (Mandatory)

| Source | Approval |
|---|---|
| Self-service within 14 days, amount ≤ ₹5,000 | Auto-approved |
| Self-service within 14 days, amount > ₹5,000 | Admin review (FINANCE role) |
| Self-service after 14 days | Admin review |
| Customer support (assisted) | Admin review |
| Gateway-initiated | Auto-processed |
| Platform-initiated | Auto-processed |

Admin review is logged in `AuditLog`.

## Refund Processing (Mandatory)

The refund is processed by the payment gateway. The platform waits for the gateway's confirmation (typically 5–7 business days). The refund is marked `COMPLETED` only after the gateway confirms.

If the gateway fails, the refund is marked `FAILED` and the user is notified. The admin can retry.

## Refund to Credit (Mandatory)

A user may choose to receive a refund as platform credit (instead of a bank refund). The credit is added to the user's `CUSTOMER_CREDIT` balance, can be used for future purchases, and does not expire (unless the user's account is deleted).

## Refund Notification (Mandatory)

The user, the creator, and the platform are notified of every refund. The notification includes:

- Order ID
- Refund amount
- Reason
- Expected processing time
- For the creator: the deduction from their balance

## Refund Rate Monitoring (Mandatory)

Refund rate = refunds in last 30 days / orders in last 30 days. A rate above 5% is flagged. A rate above 10% triggers an admin review. A rate above 20% hides the creator's Courses pending review.

## Reverse the Revenue Share (Mandatory)

When a refund is processed, the original revenue allocation is reversed:

- Creator's share is debited from the creator's available balance
- Platform's share is debited from the platform's revenue
- Taxes are reversed (credit note)
- Processing fees are NOT reversed

If the creator's available balance is insufficient, the platform covers the refund and creates a `NEGATIVE_BALANCE` on the creator. The negative balance is recovered from future earnings.

---

# Coupons and Discounts (Mandatory)

The platform supports coupons and discounts. Coupons have to prevent abuse.

## Coupon Types (Mandatory)

| Type | Description | Example |
|---|---|---|
| `PERCENT_OFF` | Percentage off the order total | 20% off |
| `FIXED_OFF` | Fixed amount off | ₹500 off |
| `BOGO` | Buy one, get one (free or discounted) | Buy a Course, get a test series free |
| `FIRST_PURCHASE` | Discount on the user's first purchase | 50% off first purchase |
| `BUNDLE` | Discount when buying a bundle | 30% off a 3-Course bundle |
| `REFERRAL` | Discount for both the referrer and the referee | ₹200 off for both |
| `CREATOR` | Creator-issued discount (creator absorbs the cost) | Creator gives 10% off to their students |
| `PLATFORM` | Platform-issued discount (platform absorbs the cost) | New Year sale |

## Coupon Field Constraints (Mandatory)

```typescript
interface Coupon {
  id: string;
  code: string;
  type: CouponType;
  value: number;
  minOrderValue?: number;
  maxDiscount?: number;
  validFrom: string;
  validUntil: string;
  usageLimit?: number;
  usageLimitPerUser?: number;
  applicableCourses?: string[];
  applicableCreators?: string[];
  firstPurchaseOnly?: boolean;
  newUserOnly?: boolean;
  combinableWithOtherCoupons?: boolean;
  costBearer: 'PLATFORM' | 'CREATOR';
  costBearerId?: string;
  status: 'DRAFT' | 'ACTIVE' | 'PAUSED' | 'EXPIRED' | 'EXHAUSTED';
  createdBy: string;
  createdAt: string;
  totalUsages: number;
}
```

## Coupon Cost Bearer (Mandatory)

- `PLATFORM`: The platform's revenue share is reduced; the creator's share is unaffected
- `CREATOR`: The creator's share is reduced; the platform's share is unaffected

A creator-issued coupon can be created only by the creator themselves (or an admin on their behalf). The creator's available balance covers the cost.

## Coupon Validation (Mandatory)

A coupon is validated at checkout time:

1. Coupon exists and is `ACTIVE`
2. Coupon is within its `validFrom` / `validUntil`
3. Coupon's `usageLimit` is not exhausted
4. User has not exceeded `usageLimitPerUser`
5. Coupon applies to the Courses in the cart
6. Cart total meets `minOrderValue`
7. User meets `firstPurchaseOnly` / `newUserOnly` constraints
8. If `!combinableWithOtherCoupons`, no other coupon is in the cart

A failure of any check rejects the coupon with a specific error message.

## Coupon Stacking (Mandatory)

Only one coupon is allowed per order, unless `combinableWithOtherCoupons` is true. The number of combinable coupons is capped at 2 (one `PLATFORM`, one `CREATOR`) to prevent abuse.

## Coupon Abuse Prevention (Mandatory)

- A user is limited to `usageLimitPerUser` (default: 1)
- Coupon codes are bound to a creator or platform; a creator cannot use another creator's coupon
- Self-referral (using your own referral code) is detected and the discount is reversed
- Coupon creation requires a minimum tenure (creators need ≥ 30 days on the platform)
- High-value coupons (≥ 50% off, or ≥ ₹5,000 off) require admin approval

## Coupon Audit (Mandatory)

Every coupon creation, update, deletion, and usage is logged. The audit includes the coupon's before/after state (for updates) and the order, user, and discount applied (for usages).

---

# Payouts (Mandatory)

Creators receive their earnings via payouts. Payouts are a financial primitive with strict audit and reconciliation.

## Payout Schedule (Mandatory)

The default payout schedule is:

- Cycle: Monthly, on the 5th of the month
- Coverage: Earnings from the 1st to the last day of the previous month
- Minimum payout: ₹500 (below this, the balance rolls over)
- Maximum payout: ₹10,00,000 (above this, requires manual approval)

Verified creators may request weekly payouts. Enterprise creators may request on-demand payouts.

## Payout Methods (Mandatory)

A creator's payout method is one of:

- Bank transfer (NEFT/IMPS/RTGS) — the default
- UPI — supported but not recommended for high-value payouts
- PayPal/Wise — for international creators (future)

A creator MUST verify their payout method before the first payout. Verification includes:

- For bank transfer: a small test deposit (₹1) and confirmation of the deposit amount
- For UPI: a UPI handle verification

## Payout Request Flow (Mandatory)

A payout is created in one of two ways:

1. Automatic: the platform creates a payout on the 5th of each month for the previous month's earnings
2. Manual: the creator requests a payout via the dashboard

The payout goes through these states:

```
REQUESTED → APPROVED → PROCESSING → COMPLETED
                ↓
              REJECTED
                ↓
              ON_HOLD
```

| State | Description | Who Can Transition |
|---|---|---|
| `REQUESTED` | Payout created, awaiting approval | Creator (manual), System (auto) |
| `APPROVED` | Approved by admin/system, ready to process | Admin (with `payout:approve`) |
| `PROCESSING` | Sent to payment gateway | System |
| `COMPLETED` | Gateway confirmed transfer | System |
| `REJECTED` | Payout rejected (e.g., bank details invalid) | Admin |
| `ON_HOLD` | Payout held for review | Admin |

## Payout Approval Rules (Mandatory)

| Condition | Approval |
|---|---|
| Auto-scheduled, creator is `verified` | Auto-approved |
| Auto-scheduled, creator is not `verified` | Auto-approved up to ₹10,000; admin review above |
| Manual, ≤ ₹10,000, creator is `verified` | Auto-approved |
| Manual, > ₹10,000, creator is `verified` | Admin review |
| Manual, creator is not `verified` | Admin review |
| Creator's chargeback rate > 1% in the last 30 days | Admin review |
| Creator's refund rate > 10% in the last 30 days | Admin review |
| First payout ever | Admin review |
| Bank details changed in the last 7 days | Admin review |
| Payout destination country ≠ creator's KYC country | Admin review |

## Payout Reconciliation (Mandatory)

The platform reconciles payouts daily:

1. Payouts marked `COMPLETED` by the gateway are matched with bank statements
2. A `PayoutReconciliation` record is created for each match
3. Unmatched payouts are flagged for review

A reconciliation break is a Sentry alert at `error` level.

## Payout Fees (Mandatory)

Payout fees are borne by the creator and deducted from the payout amount:

| Method | Fee |
|---|---|
| NEFT | ₹2.50 + GST |
| IMPS | ₹5 + GST |
| RTGS | ₹25 + GST (for amounts > ₹2,00,000) |
| UPI | Free (in beta; may change) |

The fee is shown to the creator before they confirm the payout.

## Negative Balance Recovery (Mandatory)

If a creator has a `NEGATIVE_BALANCE`, the negative balance is recovered from the next payout:

```
next_payout_amount = available_balance - abs(negative_balance)
```

If the negative balance exceeds the available balance, the entire payout is consumed and the remaining negative balance rolls over. A creator is notified of every negative balance event.

## Tax (Mandatory)

Creators are responsible for their own taxes. The platform:

- Issues a Form 16A (TDS certificate) annually for TDS deducted
- Files TDS returns (Form 26Q) quarterly
- Deducts TDS at the applicable rate

A creator can download their TDS certificates from the dashboard.

---

# Creator Analytics (Mandatory)

The creator dashboard exposes metrics that creators need to make decisions.

## Required Metrics (Mandatory)

| Metric | Description | Source |
|---|---|---|
| `revenue` | Total revenue (gross) | Sum of `Order.amount` for creator's Courses |
| `net_revenue` | Revenue after platform fee and taxes | `revenue` - `platform_fee` - `taxes` |
| `available_balance` | Balance available for payout | Ledger |
| `pending_balance` | Earnings in the settlement window | Ledger |
| `enrollments` | Total active enrollments | Count of `Enrollment` with `status = ACTIVE` |
| `new_enrollments` | Enrollments in the period | Count of `Enrollment` created in the period |
| `refund_rate` | Refund rate in the period | Refund metrics |
| `chargeback_rate` | Chargeback rate in the period | Chargeback metrics |
| `student_engagement` | Avg lessons completed, tests taken per student | Activity logs |
| `completion_rate` | % of enrolled students who completed the Course | Activity logs |
| `rating` | Average Course rating | Reviews |
| `review_count` | Number of reviews | Reviews |
| `top_courses` | Best-selling Courses | Orders |
| `payout_history` | Past payouts and their status | Payouts |

## Metric Computation (Mandatory)

Metrics are computed nightly from the database and stored in a `CreatorMetric` table with a `date` and `metricName`. The dashboard reads from this table, not from a live query on the orders table.

A metric definition lives in code; the dashboard links to the docstring.

## Time-Series Visualization

Metrics are visualized as time-series charts. The default view is the last 30 days, with options for 7d, 30d, 90d, 1y, and all-time.

## Cohort Analysis (Mandatory)

The dashboard supports cohort analysis: a cohort is a group of users who enrolled in a Course in the same week. For each cohort, the dashboard shows retention (D1, D7, D30), revenue, and engagement. Cohorts are filterable by Course, by date, and by acquisition channel.

Cohort analysis is computed nightly. The data is stored in a `CreatorCohort` table.

## Real-Time vs Delayed

| Metric | Freshness |
|---|---|
| `available_balance`, `pending_balance` | Real-time (from ledger) |
| `enrollments`, `new_enrollments` | 1 hour |
| `revenue`, `net_revenue` | 1 hour |
| `student_engagement` | 4 hours |
| `completion_rate` | 4 hours |
| `refund_rate`, `chargeback_rate` | 4 hours |
| `rating`, `review_count` | 4 hours |
| Cohort analysis | 24 hours |

A metric is labeled with its freshness in the dashboard.

---

# Creator Moderation (Mandatory)

Creators are subject to moderation. See admin-panel-rules.md for the general moderation framework. This section adds creator-specific rules.

## Creator Verification Tiers (Mandatory)

| Tier | Requirements | Privileges |
|---|---|---|
| `UNVERIFIED` | Email verified, phone verified | Can create free Courses |
| `VERIFIED` | KYC completed (PAN + Aadhaar), bank account verified | Can create paid Courses, weekly payouts |
| `ENTERPRISE` | Manual review, business documents, ≥ ₹1L GMV | Custom revenue share, on-demand payouts |

KYC includes:

- PAN verification
- Aadhaar (or equivalent) verification
- Bank account verification (test deposit)
- For institutions: GST registration, business PAN, certificate of incorporation

## Creator Suspension (Mandatory)

A creator can be suspended by an admin for:

- Policy violation (e.g., plagiarism, hate speech)
- High refund rate (≥ 20%)
- High chargeback rate (≥ 2%)
- KYC failure or fraud
- Legal request

A suspension:

- Hides all the creator's Courses from public listings
- Freezes pending payouts
- Disables new enrollments
- Preserves existing enrollments (students keep access)
- Creates an `AuditLog` entry with the reason
- Notifies the creator by email

A suspension is reviewed every 30 days.

## Plagiarism Detection (Mandatory)

A creator MAY report another creator for plagiarism. The reported content is checked via:

- An LLM-based similarity check (against the rest of the platform's content)
- A hash-based check (against the same content uploaded by another creator)
- A web search (if the content is publicly available)

A confirmed plagiarism case results in:

- The duplicated content is removed
- The creator is warned (first offense) or suspended (repeat offense)
- The original creator is notified

## Quality Bar (Mandatory)

Every Course MUST meet a quality bar before being published. The bar is enforced by an LLM-based review (see AI rules). The review checks:

- Content is grammatically correct
- Content is factually accurate (sampled)
- Content is not plagiarized
- Content is not harmful (hate speech, violence, sexual content)
- Content is appropriate for the declared exam/subject/level

A Course that fails the review is returned to the creator with feedback.

## Anti-Fake-Reviews (Mandatory)

The platform detects and removes fake reviews:

- Reviews from accounts created recently with no other activity
- Reviews from accounts that have reviewed only one creator (the creator in question)
- Reviews with timing patterns (e.g., 5 reviews within 1 hour)
- Reviews with content patterns (e.g., identical text from different accounts)

A confirmed fake review is removed. The reviewer account is flagged. The creator is notified (a single time, not every removal).

## Anti-Fake-Enrollments (Mandatory)

The platform detects and removes fake enrollments:

- Enrollments from accounts with the same device fingerprint as the creator
- Enrollments from accounts with the same payment method as the creator
- Enrollments from accounts created within 24 hours of the enrollment

A confirmed fake enrollment is reversed (refunded), the review is removed, and the creator is warned.

---

# Compatibility With Other Standards

This document defers to:

- [security-rules.md](docs/standards/security-rules.md) for auth, MFA, and audit logging
- [admin-panel-rules.md](docs/standards/admin-panel-rules.md) for moderation, audit, and impersonation
- [architecture-rules.md](docs/standards/architecture-rules.md) for module boundaries and the ADR requirement
- [backend-rules.md](docs/standards/backend-rules.md) for service patterns and Sentry instrumentation
- [ai-rules.md](docs/standards/ai-rules.md) for the LLM-based quality bar and plagiarism detection
- [rag-rules.md](docs/standards/rag-rules.md) for AI Tutor interactions on creator content

---

# Sprint 1 Compliance Checklist

Since the creator economy is not yet fully implemented, this section will be expanded in v1.0. For Sprint 1:

- [ ] ADR written for: payment gateway (Razorpay vs Stripe vs Cashfree), ledger storage (PostgreSQL vs dedicated ledger service), KYC provider (Setu vs Karza vs Signzy), payout method (NEFT/IMPS/UPI)
- [ ] Double-entry ledger implemented with all account types
- [ ] Order → settlement → payout flow implemented
- [ ] Refund flow with revenue reversal
- [ ] Coupon system with `PLATFORM` and `CREATOR` cost bearers
- [ ] Payout flow with state machine and approval rules
- [ ] Creator dashboard with required metrics
- [ ] Cohort analysis computation
- [ ] KYC integration (PAN, Aadhaar, bank)
- [ ] Plagiarism detection via LLM
- [ ] Quality bar review for new Courses
- [ ] TDS calculation and Form 16A generation

---

# Final Directive

The creator economy is GK Circle's growth engine.

A creator who loses money is a creator who leaves.

A creator who leaves is a creator who takes their students with them.

A student who leaves is a platform lost.

Build a creator economy that is traceable, fair, auditable, and rewarding.

Verify all four.
