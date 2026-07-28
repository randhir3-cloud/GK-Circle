# Exam Platform — Product Roadmap

Audience: product / aspirants / stakeholders.

## North star

```text
Create PCS Course → Subjects/Topics → MCQs → Topic Test → Attempt (scored)
  → Results → Learning Analytics → Daily Revision Queue
```

## Product phases

| Phase | Name | Product outcome | Status |
|---:|---|---|---|
| 1 | Course Builder and Publication | Build course → subjects → topics; publish; learner outline | VERIFIED |
| 2 | Question Bank, Versioning and Security | Versioned MCQs + PYQ fields; answer-key endpoints secured | VERIFIED |
| 3 | Bulk MCQ Import | CSV import of thousands of PCS questions | VERIFIED |
| 4 | Collections and Visual Test Builder | STATIC/DYNAMIC collections; visual test builder; inline +Add Question | VERIFIED |
| 5 | Attempt and Scoring Engine | Server start/autosave/submit/score/negative marks | VERIFIED |
| 6 | Student Test Player | Instructions, palette, timer, mark-for-review | IN_PROGRESS |
| 7 | Results and Answer Review | Results + explanations per release policy | NOT_STARTED |
| 8 | Learning Analytics | Weak subjects/topics; repeated wrongs; slow questions; revise-today | NOT_STARTED |
| 9 | Revision System | Daily Revision Queue + bookmarks + incorrect notebook | NOT_STARTED |
| 10 | PYQ, Coverage and Release Verification | PYQ filters; content/learner/mastery coverage; E2E verify | NOT_STARTED |

## Product rules (frozen intent)

- Question Bank remains the single store; Test Builder may create questions inline into the bank.
- Collections are STATIC and/or DYNAMIC; attempts freeze resolved snapshots.
- Coverage metrics stay distinct: content coverage ≠ learner coverage ≠ mastery.
- Scorer uses snapshot answer keys only.
- CSV import first; XLSX later with separate acceptance.
- No social feeds, monetisation, native mobile, or live video classes in this programme.
