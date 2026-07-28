# EXAM PLATFORM — FULL BROWSER-ONLY COURSE CREATION AND LEARNER ACCEPTANCE TEST

## PRIMARY OBJECTIVE

Perform a genuine end-to-end user-experience and operational-flow test of the GK Circle EXAM Platform using Playwright with a visible browser.

The test must simulate exactly how a real administrator and learner use the platform.

The test must:

1. Log in as an administrator through the login page.
2. Create a complete UPSC course through visible browser forms.
3. Create subjects through the browser.
4. Create topics through the browser.
5. Enter or import MCQs through visible production UI workflows.
6. Create quizzes and mock tests through the browser.
7. Publish the course and tests through the browser.
8. Log out through the browser.
9. Log in as a learner using a separate browser context.
10. Discover or enrol in the course through the browser.
11. Open and complete every required test through the browser.
12. Submit answers through visible answer controls.
13. Review learner results and analytics through the browser.
14. Log in again as administrator or instructor.
15. Review learner performance through the browser.
16. Generate CSV, XLSX and PDF reports through the browser.
17. Create and validate scheduled reports through the browser.
18. document every broken workflow, usability problem, validation issue, navigation issue and data-chain failure.

This is not a database population test.

This is not a backend API test.

This is not a seed-data verification test.

This is a complete operator and learner user-experience acceptance test.

---

# ABSOLUTE PROHIBITIONS

The following actions are strictly prohibited.

## Database restrictions

Do not:

- insert records directly into PostgreSQL
- run SQL INSERT statements
- use Prisma, GORM or another ORM to create test records
- use database seed scripts
- load fixture data directly into the database
- modify database tables manually
- use database restoration to create the test dataset
- copy existing course records inside the database
- create attempts, answers, reports or analytics directly in database tables

Database access may only be used for passive troubleshooting after a visible browser failure, and only when repository testing rules explicitly permit it.

Database access must never be used to complete a workflow.

## API restrictions

Do not:

- call internal APIs to create courses
- call internal APIs to create subjects
- call internal APIs to create topics
- call internal APIs to create MCQs
- call internal APIs to create quizzes
- call internal APIs to enrol learners
- call internal APIs to create attempts
- call internal APIs to submit answers
- call internal APIs to generate reports
- call internal APIs to create schedules

Playwright may observe network requests and responses for assertions.

Playwright must not use `request.post`, `fetch`, Axios, curl or any hidden API client to create application data.

Read-only API access is permitted only after the browser has completed the corresponding workflow and only for data reconciliation.

## Browser bypass restrictions

Do not:

- inject application state into localStorage
- inject authentication cookies manually
- set session tokens manually
- bypass login using storage state generated outside the browser login flow
- directly navigate to undocumented development routes
- invoke frontend component methods
- execute application JavaScript to create records
- submit hidden forms
- manipulate Vue or React internal state
- mark records complete by editing DOM attributes
- use browser evaluation to call backend services

All important actions must be performed using visible controls available to a real user.

---

# REQUIRED BROWSER MODE

The primary acceptance run must use Chromium in headed mode.

Required command:

npx playwright test \
  tests/e2e/exam-full-browser-operator-learner-readiness.spec.ts \
  --project=chromium \
  --headed \
  --trace=retain-on-failure

The browser window must remain visible during local execution.

CI may use headless mode only as a secondary verification run.

The final EXAM READY decision must be based on the full browser-driven run.

---

# REQUIRED TEST FILES

Create:

app/tests/e2e/exam-full-browser-operator-learner-readiness.spec.ts

Create reusable helpers under:

app/tests/e2e/helpers/browser-acceptance/

Required helpers:

- browser-auth.helper.ts
- browser-navigation.helper.ts
- browser-course-builder.helper.ts
- browser-subject-builder.helper.ts
- browser-topic-builder.helper.ts
- browser-mcq-builder.helper.ts
- browser-quiz-builder.helper.ts
- browser-learner-attempt.helper.ts
- browser-analytics.helper.ts
- browser-reporting.helper.ts
- browser-accessibility.helper.ts
- browser-evidence.helper.ts
- browser-console-monitor.helper.ts

Use strict TypeScript.

Do not use `any`.

Do not place business logic directly inside the main test file when it belongs in a helper.

---

# ENVIRONMENT VARIABLES

Read all environment values dynamically.

Required:

- E2E_BASE_URL
- E2E_ADMIN_EMAIL
- E2E_ADMIN_PASSWORD
- E2E_LEARNER_EMAIL
- E2E_LEARNER_PASSWORD

Optional:

- E2E_INSTRUCTOR_EMAIL
- E2E_INSTRUCTOR_PASSWORD
- E2E_MAILPIT_URL
- E2E_TEST_TIMEOUT_MS
- E2E_RUN_ID
- E2E_HEADLESS
- E2E_SLOW_MO_MS
- E2E_CLEANUP_MODE

The test must fail immediately with a clear error when mandatory values are missing.

Never print passwords, tokens, cookies, signed URLs or session values.

---

# REQUIRED RUN IDENTIFIER

Generate:

UPSC-BROWSER-E2E-<UTC_TIMESTAMP>

Use this run identifier in every test-created entity.

Example:

UPSC Civil Services Foundation — UPSC-BROWSER-E2E-20260728T120000Z

The run identifier must be visible in:

- course title
- course description
- quiz titles
- mock-test title
- scheduled-report title
- report filters or descriptions where supported

This allows all records created through the UI to be traced back to the browser run.

---

# USER-EXPERIENCE TESTING PRINCIPLE

Every creation action must test the complete visible workflow:

1. Navigate to the correct page.
2. Locate the user-visible action.
3. Open the form or dialog.
4. Inspect required fields.
5. Enter realistic data.
6. Verify inline validation.
7. Submit through the visible button.
8. Wait for the visible success state.
9. Verify navigation after save.
10. Verify the new record appears in the UI.
11. Reopen the record.
12. Verify persisted values.
13. Verify breadcrumb and parent-child relationship.
14. Verify browser refresh does not lose the saved record.
15. Record performance and usability observations.

A save operation is not considered passed merely because the server returned 200.

It passes only when the administrator can see, reopen and continue using the created data through the UI.

---

# SELECTOR POLICY

Prefer selectors in this order:

1. `getByRole`
2. `getByLabel`
3. `getByPlaceholder`
4. `getByText`
5. `getByTestId`

Do not use brittle CSS chains such as:

- `.container > div:nth-child(3) > button`
- DOM position-based selectors
- generated framework class names
- XPath based on layout

When accessible selectors are unavailable, document the accessibility gap.

Do not add test IDs merely to hide poor accessibility unless there is no stable semantic alternative.

---

# COURSE TO CREATE

Create one course through the browser.

## Course details

Title:

UPSC Civil Services General Studies Foundation — <RUN_ID>

Description:

A comprehensive UPSC Civil Services General Studies course created entirely through browser-based operator acceptance testing. It covers Preliminary and Main General Studies areas, CSAT, ethics, governance, economy, history, geography, environment, science, society, security and current affairs.

Exam category:

UPSC Civil Services Examination

Language:

English

Status:

Active or Published

Visibility:

Visible to eligible learners

Verify through the UI that:

- the course is saved
- the title persists
- the description persists
- the course status persists
- the course appears in the administration list
- the course can be reopened
- the course is visible to the learner after publication
- browser refresh does not remove or corrupt it

---

# REQUIRED SUBJECTS

Create all 12 subjects through the browser.

1. Indian History and National Movement
2. Indian Polity, Constitution and Governance
3. Indian and World Geography
4. Indian Economy and Social Development
5. Environment, Ecology and Biodiversity
6. General Science and Technology
7. Current Affairs and International Relations
8. Art and Culture
9. Ethics, Integrity and Aptitude
10. Internal Security and Disaster Management
11. Society and Social Justice
12. CSAT — Comprehension, Reasoning and Numeracy

Each subject must:

- be created through a visible form
- be attached to the created course
- have a meaningful description
- be active or published
- appear in the correct course
- be reopenable from the UI
- display the correct parent course
- persist after refresh

After creating every subject, record:

- subject title
- save duration
- success message
- resulting URL
- whether the subject appeared in the list
- whether the subject could be reopened
- any UI or validation issue

---

# REQUIRED TOPICS

Create at least five topics under every subject through browser forms.

## Indian History and National Movement

- Ancient India
- Medieval India
- Modern India
- Indian National Movement
- Post-Independence Consolidation

## Indian Polity, Constitution and Governance

- Constitutional Framework
- Fundamental Rights and Duties
- Parliament and State Legislatures
- Judiciary and Constitutional Bodies
- Governance, Transparency and Accountability

## Indian and World Geography

- Physical Geography
- Indian Geography
- World Geography
- Resources and Industries
- Human and Economic Geography

## Indian Economy and Social Development

- Basic Economic Concepts
- Indian Fiscal and Monetary Policy
- Agriculture and Rural Economy
- Infrastructure and Industry
- Poverty, Inclusion and Human Development

## Environment, Ecology and Biodiversity

- Ecology Fundamentals
- Biodiversity and Conservation
- Climate Change
- Environmental Pollution
- Environmental Laws and Institutions

## General Science and Technology

- Physics in Everyday Life
- Chemistry in Everyday Life
- Biology and Human Health
- Space and Defence Technology
- Digital, Biotechnology and Emerging Technology

## Current Affairs and International Relations

- National Current Affairs
- International Events
- International Organisations
- India's Foreign Policy
- Global Economic and Strategic Issues

## Art and Culture

- Indian Architecture
- Sculpture and Painting
- Music and Dance
- Literature and Philosophy
- Religion and Cultural Traditions

## Ethics, Integrity and Aptitude

- Ethics and Human Interface
- Attitude and Emotional Intelligence
- Integrity and Public Service Values
- Probity in Governance
- Ethical Case Studies

## Internal Security and Disaster Management

- Internal Security Challenges
- Cybersecurity
- Border and Coastal Security
- Terrorism and Organised Crime
- Disaster Preparedness and Response

## Society and Social Justice

- Indian Society and Diversity
- Women and Vulnerable Sections
- Education and Health
- Welfare Schemes and Institutions
- Urbanisation, Migration and Globalisation

## CSAT — Comprehension, Reasoning and Numeracy

- Reading Comprehension
- Logical Reasoning
- Analytical Ability
- Basic Numeracy
- Data Interpretation and Decision-Making

Expected minimum:

- 12 subjects
- 60 topics

Every topic must be entered through the visible browser form.

For every topic:

1. Open the correct subject.
2. Click the visible Add Topic action.
3. Enter the topic title.
4. Enter a meaningful description.
5. Save.
6. Verify success feedback.
7. Verify it appears under the correct subject.
8. Reopen it.
9. Verify its parent subject.
10. Refresh the page.
11. Verify persistence.

Record any break in the parent-child creation chain.

Examples of failures that must be reported:

- topic saved under the wrong subject
- topic disappears after refresh
- topic title is truncated
- form loses entered data
- save button remains disabled without explanation
- save appears successful but no topic is created
- duplicate topic is created after one click
- navigation returns to the wrong page
- breadcrumb points to the wrong course or subject

---

# MCQ CREATION REQUIREMENTS

Create at least 15 MCQs under every topic.

Expected minimum:

60 topics × 15 MCQs = 900 MCQs

Every MCQ must be created through a production-supported visible browser workflow.

Allowed workflows:

## Preferred workflow: visible UI CSV import

Use this only when the application provides an official MCQ import page or dialog.

The browser test must:

1. Open the visible import page.
2. Download or inspect the official template when available.
3. Generate a valid CSV file locally.
4. Click the visible file chooser.
5. Select the CSV file through Playwright.
6. Verify the selected filename is displayed.
7. Start import through the visible Import button.
8. Verify import preview or validation.
9. Verify valid row count.
10. Verify invalid row count is zero.
11. Confirm the import through the visible UI.
12. Wait for visible completion.
13. Verify all 15 questions appear under the correct topic.
14. Reopen representative questions.
15. Verify options, answer, explanation and difficulty.
16. Refresh and verify persistence.

Generating the CSV file locally is permitted.

Uploading the CSV using the visible browser file input is permitted.

Calling an import API directly is prohibited.

## Fallback workflow: manual browser entry

When UI import is unavailable or broken, manually create MCQs through the browser form.

For every MCQ:

1. Open the correct topic.
2. Click Add MCQ.
3. Enter the complete question.
4. Enter four meaningful options.
5. select exactly one correct answer.
6. Enter a meaningful explanation.
7. Select difficulty.
8. Enter positive marks.
9. Enter negative marks where supported.
10. Save using the visible button.
11. Verify success.
12. Verify the question appears in the topic.
13. Reopen a representative sample.
14. Verify persistence.

Do not switch silently from import to manual entry.

Record:

- selected MCQ creation mode
- reason for fallback
- time required
- imported or manually created count
- rejected count
- UI defects

---

# MCQ CONTENT RULES

Each topic must contain:

- 5 Easy MCQs
- 6 Medium MCQs
- 4 Hard MCQs

Every MCQ must include:

- meaningful question stem
- exactly four meaningful options
- exactly one correct answer
- explanation
- difficulty
- topic association
- marks
- negative marks where supported
- published or active status

Avoid:

- placeholder questions
- duplicated questions
- meaningless options
- “Question 1”
- “Option A”
- lorem ipsum
- repeated correct-answer patterns
- copyrighted coaching-centre content

Use original, syllabus-grounded content.

Maintain an in-memory answer ledger containing:

- question identifier or stable visible text
- topic
- correct option
- difficulty
- marks
- negative marks

The answer ledger may be used later by Playwright to choose visible answers during learner attempts.

It must not be used to submit answers through APIs.

---

# MCQ ENTRY UX VALIDATION

While entering MCQs, test:

- tab order
- keyboard navigation
- required-field validation
- duplicate-option handling
- multiple-correct-answer prevention
- rich-text behaviour where supported
- option-add and option-remove behaviour
- explanation persistence
- difficulty persistence
- marks validation
- negative-mark validation
- save button state
- cancel behaviour
- unsaved-change warning
- page refresh handling
- browser back-button behaviour
- duplicate submission protection

At least once, intentionally submit an invalid MCQ form and verify that:

- the user remains on the form
- inline errors are visible
- entered valid data is not lost
- error fields are clearly identified
- no incomplete MCQ is created

---

# QUIZ CREATION

Create at least 13 quizzes through visible browser UI.

## Subject quizzes

Create one quiz for each of the 12 subjects.

Each subject quiz must:

- contain 25 questions
- include 5 questions from each of the subject's five topics
- include mixed difficulty
- be timed
- support negative marking where available
- be active or published
- be visible to the learner
- include instructions
- display the correct total marks
- display the correct duration
- display the correct question count

## Comprehensive mock test

Create one course-level comprehensive mock test.

Requirements:

- minimum 60 questions
- questions from all 12 subjects
- mixed difficulty
- timed
- negative marking enabled where supported
- published
- visible to learner
- result release mode configured

All quiz configuration must be completed using visible browser controls.

Do not build a quiz by inserting question relationships directly into the database.

Do not call quiz APIs directly.

---

# QUIZ CREATION CHAIN VALIDATION

For every quiz:

1. Open the visible quiz creation workflow.
2. Enter title and instructions.
3. Select the created course.
4. Select the correct subject or course scope.
5. Add questions through visible selectors.
6. Configure marks.
7. Configure negative marks.
8. Configure duration.
9. Configure attempt rules.
10. Configure result release.
11. Save draft.
12. Reopen draft.
13. Verify all values persist.
14. Publish through the visible button.
15. Confirm publication.
16. Verify the quiz appears in the course.
17. Verify it is learner-visible.

Report any break in the creation chain, including:

- questions missing after save
- wrong question count
- wrong topic association
- lost instructions
- lost duration
- publication state not updated
- quiz visible before publication
- quiz unavailable after publication
- duplicate quiz created after one click
- page redirects incorrectly
- filters reset unexpectedly
- selected questions disappear when changing pages

---

# ADMIN EXPERIENCE REVIEW

During the administrator workflow, actively evaluate:

- ease of finding course management
- clarity of navigation
- form labelling
- field descriptions
- required-field indicators
- error-message clarity
- success-message clarity
- ability to continue from course to subject to topic to MCQ
- whether the operator knows what to do next
- search usability
- pagination usability
- bulk-action usability
- import usability
- loading-state clarity
- accidental duplicate prevention
- unsaved-change protection
- responsiveness
- keyboard accessibility
- mobile usability where administration is supported

For each issue, record:

- severity
- page
- exact action
- expected behaviour
- actual behaviour
- screenshot
- trace reference
- whether workflow was blocked
- workaround, when one exists

Severity:

- BLOCKER
- CRITICAL
- MAJOR
- MINOR
- COSMETIC

---

# LEARNER SESSION ISOLATION

Use a completely separate browser context for the learner.

Do not reuse:

- administrator cookies
- administrator localStorage
- administrator session storage
- administrator page
- administrator browser context

The learner must log in through the visible login page.

Verify:

- learner dashboard loads
- learner cannot open admin routes
- learner cannot access course-management controls
- learner cannot access instructor-only reports
- learner sees the published course
- learner can enrol through the visible UI when required

---

# LEARNER COURSE EXPERIENCE

The learner must complete the real visible course-discovery and test-taking flow.

1. Log in through the browser.
2. Navigate to course discovery or assigned courses.
3. Find the created UPSC course.
4. Open the course.
5. Verify course title and description.
6. Verify subject list.
7. Verify topic structure where learner-visible.
8. Enrol using the visible Enrol button if required.
9. Verify enrolment success.
10. Verify all published tests are visible.
11. Verify unpublished content is not visible.

Record usability issues such as:

- course cannot be found
- course appears under wrong category
- learner sees admin controls
- enrol button is unclear
- success feedback is missing
- published quiz does not appear
- duplicated quiz cards
- incorrect test metadata
- inaccessible navigation

---

# LEARNER TEST ATTEMPTS

The learner must take all 13 tests through the open browser.

For each quiz:

1. Open the quiz card.
2. Read visible instructions.
3. Verify duration.
4. Verify marks.
5. Verify negative-marking information.
6. Verify number of questions.
7. Click Start Test.
8. Confirm the start when required.
9. Wait for the first question to render.
10. Select answers using visible option controls.
11. Navigate using visible Next, Previous or question-palette controls.
12. Verify selected answers persist.
13. Answer all required questions.
14. Leave some questions unanswered in selected attempts.
15. submit through the visible Submit button.
16. Confirm final submission through the visible dialog.
17. Verify only one submission occurs.
18. Verify the result or pending-result screen.
19. Record displayed score and statistics.

Do not submit attempts through API calls.

Do not insert attempt answers into the database.

Do not call frontend internals.

Every answer must be selected using Playwright locators on visible answer controls.

---

# CONTROLLED LEARNER PERFORMANCE

Use the generated answer ledger to create predictable results through visible selections.

## Strong performance

For four subject quizzes:

- 80–90% correct
- some incorrect answers
- at least one skipped answer across the group

## Medium performance

For four subject quizzes:

- 55–70% correct
- mixed correct and incorrect answers

## Weak performance

For four subject quizzes:

- 25–45% correct
- include incorrect and skipped answers

## Comprehensive mock test

Produce:

- strong performance in at least two subjects
- medium performance in several subjects
- weak performance in at least two subjects
- some skipped questions
- some negative-marked answers

The goal is to produce meaningful analytics.

---

# TEST-TAKING EXPERIENCE VALIDATION

During attempts, verify:

- timer starts correctly
- timer does not reset on navigation
- question number updates
- selected option remains selected
- previous and next controls work
- question palette reflects answered state
- marked-for-review behaviour works where supported
- refresh restores the active attempt where supported
- accidental back navigation produces appropriate warning
- duplicate Submit clicks create only one submission
- final confirmation is clear
- unanswered-question warning is accurate
- negative-marking information is visible
- keyboard navigation works
- visible focus is present
- mobile layout has no horizontal overflow
- submission error does not lose answers
- results are not shown before allowed release

At least one attempt must include:

- browser refresh
- leaving and returning
- question navigation backwards
- one skipped question
- repeated submit click
- mobile viewport validation

---

# RESULT RELEASE TESTING

Test all supported result-release modes through visible UI.

At minimum:

- immediate release
- manual release
- scheduled release where supported

For manual release:

1. Learner completes the quiz.
2. Verify result is pending or unavailable.
3. Return to admin or instructor context.
4. Open the visible result-release interface.
5. Release results using the visible button.
6. Return to learner context.
7. Refresh through the browser.
8. Verify results are now visible.

Do not change release state through APIs or database updates.

---

# LEARNER ANALYTICS

After completing all tests, open learner analytics through the browser.

Verify:

- total tests attempted
- completed tests
- average score
- overall accuracy
- correct answers
- incorrect answers
- skipped answers
- time spent
- score trend
- subject performance
- topic performance
- strengths
- weaknesses
- recent attempts
- negative-mark impact where shown

Compare visible analytics to the deterministic answer ledger.

Counts must be exact.

Percentages may use documented UI rounding tolerance.

Fail when analytics contain:

- blank widgets
- NaN
- Infinity
- negative percentages
- incorrect totals
- missing attempts
- duplicated attempts
- wrong subject attribution
- wrong topic attribution

---

# ADMIN OR INSTRUCTOR ANALYTICS

Log in as the administrator or instructor through the browser.

Open analytics using normal navigation.

Verify:

- course count
- quiz count
- learner count
- attempt count
- average performance
- subject-level performance
- topic-level performance
- learner-level performance
- question-level metrics
- option selection distribution
- completion counts
- time-spent metrics
- result-release status

Do not open analytics using a hidden URL unless that URL is part of normal application navigation.

Verify access boundaries:

- learner cannot access instructor analytics
- unrelated users cannot access protected quiz analytics
- administrator or authorised instructor can access required analytics

---

# REPORT GENERATION THROUGH BROWSER

Generate reports only through visible reporting UI.

Required formats:

- CSV
- XLSX
- PDF

Required report scopes:

- course or portfolio overview
- quiz summary
- learner performance
- question metrics
- engagement
- full dashboard where supported

For each report:

1. Navigate to reporting through the UI.
2. Click the visible Export or Generate Report action.
3. Select report type.
4. Select format.
5. Apply visible filters.
6. Click Generate.
7. Verify visible queued or processing state.
8. Poll only through visible status controls.
9. Wait for completed status.
10. Click the visible Download action.
11. Capture the Playwright download.
12. Validate the downloaded file.
13. Verify report appears in history.
14. Verify download count or state updates where supported.

Do not call export APIs directly.

Do not insert report jobs into the database.

---

# SCHEDULED REPORTS THROUGH BROWSER

Create scheduled reports through the visible schedule form.

Test:

- create schedule
- select report type
- select format
- select recipients
- select cadence
- select timezone
- save
- verify schedule appears
- reopen schedule
- edit schedule
- disable schedule
- enable schedule
- delete schedule

Trigger execution only through an official UI action, supported QA scheduler mechanism or normal scheduler operation.

Do not create schedule records through SQL or APIs.

---

# EMAIL DELIVERY VALIDATION

Use Mailpit or the configured QA inbox.

The browser workflow must generate the email naturally.

Verify:

- email arrives
- subject is correct
- report title is correct
- recipient is correct
- attachment or download-link behaviour is correct
- expiry information is shown
- no signed URL is logged in evidence
- no duplicate email is created

Email verification may read Mailpit through its supported test interface after the browser has created the schedule.

The schedule itself must still be created through the application browser UI.

---

# CONSOLE AND NETWORK MONITORING

Attach monitoring before every major workflow.

Fail on unexpected:

- console errors
- page errors
- unhandled promise rejections
- hydration errors
- chunk-load errors
- 5xx responses
- failed save requests
- repeated submissions
- endless polling
- aborted navigation without user feedback

Record all relevant request URLs without sensitive query parameters.

A backend 200 response does not override a broken browser experience.

A UI workflow still fails when:

- the page does not update
- feedback is missing
- the record is not visible
- navigation is incorrect
- fields lose data
- the user cannot continue

---

# RESPONSIVE TESTING

Test representative workflows at:

- 1440 × 900
- 768 × 1024
- 360 × 800

Required mobile checks:

- login works
- course page works
- quiz page works
- options remain selectable
- navigation controls remain visible
- submit button remains reachable
- no horizontal overflow
- dialogs fit the viewport
- analytics remain readable
- report history remains usable

---

# ACCESSIBILITY TESTING

Use `@axe-core/playwright` or the repository-approved equivalent.

Run checks on:

- login page
- admin dashboard
- course creation form
- subject form
- topic form
- MCQ form or import page
- quiz creation page
- learner course page
- learner attempt page
- learner result page
- learner analytics
- instructor analytics
- report page

Fail on:

- critical violations
- serious violations without approved exception

Also manually validate through Playwright:

- keyboard-only navigation
- visible focus
- accessible form labels
- field error association
- dialog focus trap
- answer-option accessible names
- button accessible names
- table headers

---

# CHECKPOINTING

The suite may save progress only for recovery.

Checkpointing must not replace browser verification.

Checkpoint path:

artifacts/exam-readiness/<RUN_ID>/checkpoint.json

Checkpoint example:

{
  "schemaVersion": 1,
  "runId": "",
  "baseUrl": "",
  "lastCompletedPhase": "",
  "course": {
    "title": "",
    "url": "",
    "verifiedInBrowser": false
  },
  "subjects": {},
  "topics": {},
  "mcqs": {},
  "quizzes": {},
  "attempts": {},
  "reports": {}
}

On resume:

1. Log in through the browser again.
2. Navigate to every checkpointed entity.
3. Verify it exists.
4. Verify it belongs to the current run.
5. Continue only after browser validation.

Do not trust a checkpoint without visible confirmation.

---

# CLEANUP

Supported cleanup modes:

- NONE
- ARCHIVE
- DELETE

Cleanup must use visible application controls.

Do not delete records directly from the database.

In DELETE mode:

1. Open created schedules.
2. Delete them through the UI.
3. Open quizzes.
4. Delete or archive them through the UI.
5. Open topics and subjects only where UI deletion is safely supported.
6. Delete or archive the course through the UI.
7. Verify the records no longer appear.

Never delete records not associated with the current run.

When the UI does not support deletion, use ARCHIVE or leave the run data in place and report the limitation.

---

# REQUIRED EVIDENCE

Create:

artifacts/exam-readiness/<RUN_ID>/

Include:

- acceptance-results.json
- browser-user-experience-report.md
- admin-workflow-report.md
- learner-workflow-report.md
- data-chain-report.md
- accessibility-report.json
- console-errors.json
- network-failures.json
- performance-summary.json
- checkpoint.json
- screenshots/
- downloads/
- traces/
- videos/ when enabled

Required screenshots:

1. Admin login
2. Admin dashboard
3. Empty or initial course list
4. Course creation form
5. Created course
6. Subject creation
7. Subject list
8. Topic creation
9. Topic list
10. MCQ form or import dialog
11. Import validation
12. Topic showing 15 MCQs
13. Quiz builder
14. Published quiz
15. Learner login
16. Learner course discovery
17. Learner course page
18. Active test
19. Question navigation
20. Final submission
21. Learner result
22. Learner analytics
23. Instructor analytics
24. Report-generation dialog
25. Report history
26. Completed CSV report
27. Completed XLSX report
28. Completed PDF report
29. Scheduled-report form
30. Mailpit delivery
31. Mobile learner attempt
32. Any blocker or broken workflow

---

# DATA-CHAIN REPORT

Create:

data-chain-report.md

It must document whether the following full chain works without hidden intervention:

Admin login
→ Course creation
→ Subject creation
→ Topic creation
→ MCQ creation
→ Quiz creation
→ Quiz publication
→ Learner course discovery
→ Learner enrolment
→ Learner attempt
→ Answer persistence
→ Submission
→ Result release
→ Learner analytics
→ Instructor analytics
→ Report generation
→ Report download
→ Scheduled report
→ Email delivery

For every link, record:

- PASS or FAIL
- page
- action performed
- expected result
- observed result
- error message
- screenshot
- severity
- workaround

Any broken mandatory link means the full chain has failed.

---

# MACHINE-READABLE RESULTS

Create:

acceptance-results.json

Use:

{
  "runId": "",
  "startedAt": "",
  "completedAt": "",
  "browserDriven": true,
  "databaseSeedingUsed": false,
  "creationApiUsed": false,
  "headedBrowserRun": true,
  "adminWorkflow": {
    "status": "PASSED|FAILED",
    "courseCreatedThroughBrowser": false,
    "subjectsCreatedThroughBrowser": 0,
    "topicsCreatedThroughBrowser": 0,
    "mcqsCreatedThroughBrowser": 0,
    "quizzesCreatedThroughBrowser": 0
  },
  "learnerWorkflow": {
    "status": "PASSED|FAILED",
    "loggedInThroughBrowser": false,
    "enrolledThroughBrowser": false,
    "testsStartedThroughBrowser": 0,
    "answersSelectedThroughBrowser": 0,
    "testsSubmittedThroughBrowser": 0
  },
  "analytics": {
    "learnerAnalyticsVerified": false,
    "instructorAnalyticsVerified": false,
    "totalsReconciled": false
  },
  "reports": {
    "csvGeneratedThroughBrowser": false,
    "xlsxGeneratedThroughBrowser": false,
    "pdfGeneratedThroughBrowser": false,
    "scheduledReportCreatedThroughBrowser": false,
    "emailVerified": false
  },
  "userExperience": {
    "blockers": [],
    "criticalIssues": [],
    "majorIssues": [],
    "minorIssues": []
  },
  "dataChain": {
    "status": "PASSED|FAILED",
    "brokenLinks": []
  },
  "finalDecision": "EXAM READY|EXAM NOT READY"
}

The following must remain exactly false:

- databaseSeedingUsed
- creationApiUsed

When either becomes true, the run is invalid and cannot produce EXAM READY.

---

# REQUIRED ASSERTIONS

At the end, assert:

- exactly one run-specific course was created through browser UI
- at least 12 subjects were created through browser UI
- at least 60 topics were created through browser UI
- at least 900 MCQs were created or imported through visible browser UI
- at least 13 quizzes were created through browser UI
- at least 13 quizzes were published through browser UI
- learner logged in through browser
- learner found the course through browser
- learner enrolled through browser where required
- learner opened every quiz through browser
- learner selected every recorded answer through visible controls
- learner submitted all 13 attempts through visible controls
- analytics reflect the attempts
- reports were requested through visible controls
- files were downloaded through visible controls
- scheduled report was created through visible controls
- no database seeding occurred
- no creation APIs were called directly
- no mandatory chain link failed

---

# INVALID RUN CONDITIONS

The complete run must be marked INVALID rather than passed when:

- any required data was created by database seed
- any required data was inserted using SQL
- any required entity was created by direct API call
- attempt answers were inserted directly
- reports were created directly
- authentication was bypassed
- learner session reused the administrator session
- browser workflow was skipped because seeded data already existed
- the suite only verified pre-existing data
- the browser never visibly entered or uploaded the data
- the test claims EXAM READY without completing the user-facing workflow

Invalid-run output:

TEST RUN INVALID — NON-BROWSER DATA CREATION DETECTED

---

# EXAM READY CRITERIA

Return:

EXAM READY

only when:

1. All course data was entered or imported through visible production browser UI.
2. No SQL, seed or direct creation API was used.
3. The administrator completed the full creation workflow.
4. All parent-child relationships remained correct.
5. All 900 MCQs persisted.
6. All quizzes were built and published through the UI.
7. The learner completed every test through a separate visible browser session.
8. Answers were selected and submitted through visible controls.
9. Analytics reconciled with visible learner activity.
10. CSV, XLSX and PDF reports were generated through visible UI.
11. Scheduled reporting worked through visible UI.
12. No mandatory link in the data chain failed.
13. No blocker or critical user-experience defect remained.
14. The headed Playwright execution evidence was produced.

Otherwise return:

EXAM NOT READY

List every blocking defect and broken chain link.

---

# FINAL REPORT FORMAT

At completion, output:

============================================================
EXAM PLATFORM — FULL BROWSER USER-EXPERIENCE ACCEPTANCE TEST
============================================================

Run ID:
Browser:
Headed mode:
Database seeding used: NO
Direct creation API used: NO

ADMIN FLOW
Course created through browser:
Subjects created through browser:
Topics created through browser:
MCQs entered/imported through browser:
Quizzes created through browser:
Quizzes published through browser:

LEARNER FLOW
Learner login through browser:
Course discovered:
Learner enrolled:
Tests started:
Tests completed:
Answers selected through browser:
Submissions completed:

ANALYTICS
Learner analytics:
Instructor analytics:
Data reconciled:

REPORTING
CSV:
XLSX:
PDF:
Scheduled reports:
Email delivery:

USER EXPERIENCE
Blockers:
Critical issues:
Major issues:
Minor issues:

DATA CHAIN
Course → Subject:
Subject → Topic:
Topic → MCQ:
MCQ → Quiz:
Quiz → Learner:
Learner → Attempt:
Attempt → Analytics:
Analytics → Report:
Report → Email:

FINAL DECISION:
EXAM READY

or

EXAM NOT READY

---

# STOP CONDITION

Stop after completing this browser-only operator-to-learner acceptance test.

Do not implement unrelated features.

Do not seed data to make the test pass.

Do not repair failures by inserting records manually.

When a workflow fails, capture the evidence and report the failure.

The purpose of this test is to reveal the real administrator and learner experience, including every break in the data-entry and test-taking chain.

Add a "Human Operator Review Mode"

I would add one section:

HUMAN OPERATOR MODE

While executing every browser interaction,
behave like a first-time administrator.

Do not merely automate clicks.

Evaluate the usability.

After every major workflow record:

• Was it obvious what to do?
• Was navigation intuitive?
• Were required fields obvious?
• Was validation understandable?
• Was the save confirmation obvious?
• Did the page return somewhere unexpected?
• Did the operator lose context?
• Were there too many clicks?
• Could this workflow be simplified?
• Was the terminology understandable?
• Was the workflow fast?
• Would a non-technical employee understand it?

Rate every workflow:

★★★★★ Excellent
★★★★ Good
★★★ Acceptable
★★ Poor
★ Unusable

For every workflow calculate:

Number of clicks
Number of page transitions
Number of dialogs
Typing effort
Estimated operator time

Suggest UX improvements.

Never automatically classify a workflow as good.

Only rate it after observing the actual browser behaviour.
Also add "Friction Detection"

This is something most automation never measures.

Example:

Operator clicks

↓

Unexpected dialog

↓

Extra confirmation

↓

Extra page

↓

Lost navigation

↓

Must search again

The AI should detect this automatically.

Example report:

MCQ Creation

Clicks: 14

Typing:

210 characters

Page loads:

3

Dialogs:

2

Navigation resets:

1

UX Rating:

★★☆☆☆

Recommendation:

Keep operator on Topic page after Save.

Automatically reopen Add MCQ dialog.

Reduce workflow from 14 clicks to 6.

That becomes incredibly valuable.

Measure Operator Productivity

At the end generate something like:

Operator Productivity

Course Creation

6 min

★★★★★

Subject Creation

12 min

★★★★☆

Topic Creation

10 min

★★★☆☆

MCQ Creation

48 min

★★☆☆☆

Recommendation

Bulk Entry Required

Quiz Creation

8 min

★★★★★

Learner Flow

7 min

★★★★★

Analytics

3 min

★★★★★

Overall UX

4.4 / 5

That tells you exactly where the software needs improvement.

Add "AI UX Suggestions"

Instead of only saying

MCQ creation is slow.

the report should say

Suggestion:

Allow keyboard shortcut Ctrl+Enter to save.

Automatically focus Question textbox after Save.

Keep dialog open.

Preselect previous Topic.

Remember previous Difficulty.

Expected productivity improvement:

45%.

That makes the suite a product-improvement tool as well as a test suite.

## HUMAN OBSERVATION MODE (MANDATORY)

This is NOT a background automation test.

The primary purpose of this test is to allow the project owner to WATCH the entire administrator and learner workflow in real time.

Therefore the browser MUST remain visible throughout execution.

--------------------------------------------------
BROWSER EXECUTION REQUIREMENTS
--------------------------------------------------

The Playwright browser MUST run in headed mode.

Never run headless.

Required configuration:

headless: false

slowMo: 300

viewport:
1920 × 1080

Launch Chromium maximized.

Do NOT minimise the browser.

Do NOT hide the browser behind other windows.

Do NOT close the browser automatically when the test finishes unless the user explicitly requests it.

--------------------------------------------------
LIVE EXECUTION
--------------------------------------------------

The browser should visibly perform every action.

Examples:

✓ Open login page

✓ Type username character by character

✓ Type password

✓ Click Login

✓ Wait for dashboard

✓ Navigate to Course Management

✓ Click "Create Course"

✓ Fill every field

✓ Save

✓ Wait for success message

✓ Open created course

✓ Create Subject

✓ Create Topic

✓ Upload MCQs

✓ Create Quiz

✓ Publish Quiz

✓ Logout

✓ Login as Learner

✓ Open Course

✓ Start Quiz

✓ Select answers

✓ Submit

✓ Open Analytics

✓ Generate Reports

The user should be able to watch every interaction.

--------------------------------------------------
DO NOT EXECUTE TOO FAST
--------------------------------------------------

Do NOT click as fast as possible.

Add realistic delays between actions.

Examples:

Pause after navigation:
500–1000 ms

Pause after success messages:
1000 ms

Pause after opening dialogs:
500 ms

Pause after Save:
wait until success message is visible

Pause before closing dialogs:
500 ms

The goal is to let the user observe the workflow.

--------------------------------------------------
LIVE COMMENTARY
--------------------------------------------------

Print progress in the terminal.

Example:

==================================================
STEP 1 / 12

Logging in as Administrator
==================================================

✓ Login page opened

Typing username...

Typing password...

Clicking Login...

Dashboard loaded

Duration: 3.2 s

--------------------------------------------------

STEP 2 / 12

Creating Course

--------------------------------------------------

Opening Course Management

Clicking Create Course

Entering Title

Entering Description

Selecting Category

Saving Course

Waiting for confirmation

✓ Course Created

--------------------------------------------------

Continue for every workflow.

--------------------------------------------------
ERROR REPORTING
--------------------------------------------------

If any step fails:

STOP immediately.

Do NOT continue.

Keep the browser open.

Highlight the failing control.

Take a screenshot.

Capture Playwright trace.

Print:

Current Page

Current URL

Action being attempted

Expected Result

Actual Result

Visible Validation Message

Browser Console Errors

Failed Network Request

DOM Screenshot

Then WAIT.

Do not automatically recover.

Allow the user to inspect the browser.

--------------------------------------------------
BREAKPOINT MODE
--------------------------------------------------

When a major workflow completes:

Pause.

Display:

Press ENTER to continue...

or

Wait 10 seconds before proceeding.

Major checkpoints include:

✓ Course created

✓ Subject completed

✓ Topic completed

✓ MCQs imported

✓ Quiz published

✓ Learner logged in

✓ Quiz submitted

✓ Reports generated

This allows the user to inspect the application before continuing.

--------------------------------------------------
SCREEN RECORDING
--------------------------------------------------

Enable Playwright video recording.

Record every browser session.

Save videos under:

artifacts/exam-readiness/<RUN_ID>/videos/

--------------------------------------------------
SCREENSHOTS
--------------------------------------------------

Take screenshots:

Before every major action

After every successful action

After every failure

Store under:

artifacts/exam-readiness/<RUN_ID>/screenshots/

--------------------------------------------------
LIVE UX ANALYSIS
--------------------------------------------------

While operating the browser, behave like a human tester.

Comment on usability.

Examples:

"This Create Course button is difficult to discover."

"The Save confirmation disappears too quickly."

"The Topic form resets unexpectedly."

"Too many clicks are required to create one MCQ."

"Validation message is unclear."

"Navigation after save is confusing."

Record every UX observation.

--------------------------------------------------
DO NOT HIDE FAILURES
--------------------------------------------------

Never repair a failed browser workflow using:

• SQL
• Database edits
• API calls
• JavaScript injection
• Manual state changes

The objective is to expose flaws, not hide them.

--------------------------------------------------
SUCCESS CONDITION
--------------------------------------------------

A test is considered successful only if the entire workflow is completed visibly in the browser and the user can observe every action from start to finish.


Also change your Playwright configuration

Tell the agent to use settings like these:

use: {
  browserName: "chromium",
  headless: false,
  launchOptions: {
    slowMo: 300,
    args: ["--start-maximized"],
  },
  viewport: null,
  video: "on",
  screenshot: "on",
  trace: "retain-on-failure",
},

For debugging, you can even increase:

slowMo: 500

or

slowMo: 1000

so you can clearly watch every click and understand where the workflow succeeds or fails.

This approach turns the test into an interactive demonstration of the complete administrator and learner experience, making it much easier to spot broken navigation, validation issues, missing feedback, or any break in the course-creation chain.

A few Playwright-specific details must be implemented carefully so the browser genuinely stays visible and failures remain inspectable.

1. Configure the browser at runner level

Do not rely only on settings inside the test specification. Put the observation configuration in a dedicated Playwright project or configuration file.

import { defineConfig, devices } from "@playwright/test";

export default defineConfig({
  testDir: "./tests/e2e",
  timeout: 4 * 60 * 60 * 1000,
  expect: {
    timeout: 15_000,
  },
  fullyParallel: false,
  workers: 1,
  retries: 0,
  reporter: [["line"]],
  use: {
    ...devices["Desktop Chrome"],
    headless: false,
    viewport: {
      width: 1920,
      height: 1080,
    },
    launchOptions: {
      slowMo: 300,
    },
    video: "on",
    screenshot: "on",
    trace: "retain-on-failure",
  },
  projects: [
    {
      name: "chromium-observation",
      use: {
        browserName: "chromium",
        headless: false,
        launchOptions: {
          slowMo: 300,
        },
      },
    },
  ],
});

Run it with:

npx playwright test tests/e2e/exam-full-browser-operator-learner-readiness.spec.ts \
  --project=chromium-observation \
  --headed \
  --workers=1

workers=1 is important. Parallel browser windows would make the demonstration difficult to follow.

2. Do not mix a fixed viewport with “maximised window” assumptions

A fixed 1920 × 1080 viewport provides deterministic screenshots and visual checks.

Using viewport: null with --start-maximized instead uses the actual available screen size.

Choose one observation mode:

viewport: {
  width: 1920,
  height: 1080,
}

is preferable for consistent evidence.

Do not claim that this guarantees a physically maximised operating-system window. It guarantees the browser content viewport.

3. “Keep the browser open on failure” needs an explicit pause

The Playwright runner normally closes its browser context after a test fails. Logging an error alone will not keep the window open.

Add an observation pause controlled by an environment variable:

import { Page } from "@playwright/test";
import { createInterface } from "node:readline/promises";
import { stdin, stdout } from "node:process";

export async function pauseForHumanInspection(
  page: Page,
  message: string,
): Promise<void> {
  if (process.env.E2E_HUMAN_OBSERVATION !== "true") {
    return;
  }

  console.error("\n==================================================");
  console.error("TEST EXECUTION PAUSED");
  console.error(message);
  console.error(`Current URL: ${page.url()}`);
  console.error("The browser will remain open for inspection.");
  console.error("Press ENTER in this terminal to continue teardown.");
  console.error("==================================================\n");

  const readline = createInterface({ input: stdin, output: stdout });

  try {
    await readline.question("");
  } finally {
    readline.close();
  }
}

Use it inside the failure handler before rethrowing:

try {
  await createCourseThroughBrowser(page, course);
} catch (error: unknown) {
  await captureFailureEvidence(page, "course-creation", error);
  await pauseForHumanInspection(
    page,
    "Course creation failed. Inspect the visible form and error state.",
  );

  throw error;
}

This is the mechanism that actually keeps the browser visible.

4. Add two distinct operating modes

Use an explicit environment variable:

E2E_HUMAN_OBSERVATION=true

Recommended modes:

Observation mode
- headed
- slowMo enabled
- one worker
- pauses at major checkpoints
- pauses indefinitely on failure
- full terminal commentary

Automated mode
- may run headless
- no interactive prompts
- evidence still captured
- suitable for CI

The release-observation command should be:

E2E_HUMAN_OBSERVATION=true \
E2E_PAUSE_AT_CHECKPOINTS=true \
E2E_SLOW_MO_MS=300 \
npx playwright test \
  tests/e2e/exam-full-browser-operator-learner-readiness.spec.ts \
  --project=chromium-observation \
  --headed \
  --workers=1

On Windows PowerShell:

$env:E2E_HUMAN_OBSERVATION = "true"
$env:E2E_PAUSE_AT_CHECKPOINTS = "true"
$env:E2E_SLOW_MO_MS = "300"

npx playwright test `
  tests/e2e/exam-full-browser-operator-learner-readiness.spec.ts `
  --project=chromium-observation `
  --headed `
  --workers=1
5. Avoid typing the password character by character

Typing normal form content visibly is useful.

Typing credentials character by character is unnecessary and creates a shoulder-surfing risk. The password field should remain masked, but the terminal must also never print the password.

Recommended behaviour:

console.log("Entering administrator email...");
await page.getByLabel(/email/i).fill(credentials.email);

console.log("Entering administrator password securely...");
await page.getByLabel(/password/i).fill(credentials.password);

For ordinary content fields, visible typing can be used:

await titleInput.pressSequentially(course.title, {
  delay: 30,
});

Use slower character entry selectively. Typing 900 MCQs character by character would make the suite impractically long.

6. Use meaningful pauses, not delays everywhere

Do not add arbitrary sleeps after every locator action. That will make the suite slow and unreliable.

Use Playwright’s state-based waits first:

await expect(page.getByRole("dialog")).toBeVisible();
await expect(page.getByText(/course created successfully/i)).toBeVisible();
await expect(page).toHaveURL(/courses\/[^/]+$/);

Then add a short observation pause:

await observationDelay(800);

A suitable helper:

export async function observationDelay(
  milliseconds = 700,
): Promise<void> {
  if (process.env.E2E_HUMAN_OBSERVATION !== "true") {
    return;
  }

  await new Promise<void>((resolve) => {
    setTimeout(resolve, milliseconds);
  });
}

This keeps correctness based on visible application state rather than fixed timing.

7. Add visible stage overlays

Terminal logs are useful, but the browser itself should show which phase is running.

Inject a temporary non-interactive observation banner:

import { Page } from "@playwright/test";

export async function showObservationBanner(
  page: Page,
  title: string,
  detail: string,
): Promise<void> {
  if (process.env.E2E_HUMAN_OBSERVATION !== "true") {
    return;
  }

  await page.evaluate(
    ({ titleText, detailText }) => {
      const existing = document.getElementById(
        "gk-e2e-observation-banner",
      );

      existing?.remove();

      const banner = document.createElement("aside");
      banner.id = "gk-e2e-observation-banner";
      banner.setAttribute("aria-live", "polite");
      banner.style.position = "fixed";
      banner.style.top = "16px";
      banner.style.right = "16px";
      banner.style.zIndex = "2147483647";
      banner.style.maxWidth = "420px";
      banner.style.padding = "14px 18px";
      banner.style.borderRadius = "10px";
      banner.style.background = "rgba(15, 23, 42, 0.96)";
      banner.style.color = "white";
      banner.style.fontFamily = "system-ui, sans-serif";
      banner.style.boxShadow = "0 10px 30px rgba(0, 0, 0, 0.35)";
      banner.style.pointerEvents = "none";

      const heading = document.createElement("strong");
      heading.textContent = titleText;
      heading.style.display = "block";
      heading.style.marginBottom = "6px";

      const description = document.createElement("span");
      description.textContent = detailText;

      banner.append(heading, description);
      document.body.appendChild(banner);
    },
    {
      titleText: title,
      detailText: detail,
    },
  );
}

Example:

await showObservationBanner(
  page,
  "Phase 2: Course Creation",
  "Entering course information through the visible administrator form.",
);

The banner must not be used to manipulate application state. It is only an observation aid.

8. Highlight the control before interacting with it

To make each action easier to follow, briefly outline the target element before clicking or typing.

import { Locator } from "@playwright/test";

export async function highlightForObservation(
  locator: Locator,
): Promise<void> {
  if (process.env.E2E_HUMAN_OBSERVATION !== "true") {
    return;
  }

  await locator.evaluate((element) => {
    const htmlElement = element as HTMLElement;
    const originalOutline = htmlElement.style.outline;
    const originalOffset = htmlElement.style.outlineOffset;

    htmlElement.style.outline = "4px solid #f59e0b";
    htmlElement.style.outlineOffset = "3px";

    window.setTimeout(() => {
      htmlElement.style.outline = originalOutline;
      htmlElement.style.outlineOffset = originalOffset;
    }, 1_200);
  });

  await observationDelay(600);
}

Usage:

const createCourseButton = page.getByRole("button", {
  name: /create course/i,
});

await highlightForObservation(createCourseButton);
await createCourseButton.click();

This makes it clear which UI element the automation is using.

9. Failure output must include exact context

On failure, print:

Phase
Step
Role: Administrator or Learner
Current page title
Current URL
Target locator description
Action attempted
Expected visible result
Observed result
Visible validation messages
Console errors
Page errors
Failed requests
HTTP 5xx responses
Screenshot path
Trace status
Checkpoint path

Do not print:

passwords
session cookies
access tokens
signed download URLs
complete request headers
sensitive learner data
10. Capture evidence before pausing

The failure order should be:

Failure detected
→ Stop further workflow actions
→ Capture screenshot
→ Capture DOM snapshot or relevant visible text
→ Flush console/network evidence
→ Write checkpoint
→ Print failure summary
→ Pause for human inspection
→ Rethrow after user presses Enter

Do not pause before writing evidence, because the browser or terminal might later be closed manually.

11. Store videos using Playwright’s output lifecycle

Playwright initially writes video files into its configured output directory. The final filename may only be available after the page or context closes.

Configure:

use: {
  video: "on",
}

Then copy or attach the resulting video into:

artifacts/exam-readiness/<RUN_ID>/videos/

during teardown.

Do not assume that setting video: "on" alone writes directly to the custom evidence path.

12. Add mandatory checkpoints

Pause after:

administrator login
course creation and reopen verification
each subject group
every five completed topics
every MCQ import batch
each published quiz
learner login
first learner test start
first learner submission
learner analytics
instructor analytics
each export format
scheduled report creation
final decision

Because 60 topics and 900 MCQs are involved, pausing after every single record would make the run excessively long. Use grouped checkpoints while still showing every browser action.

Required final implementation statement

Add this to the agent instructions:

The acceptance suite must not be considered executed merely because the
Playwright process completed.

The primary certification run is valid only when:

1. Chromium opened visibly on the project owner's desktop.
2. The browser remained headed for the entire workflow.
3. Exactly one Playwright worker executed the workflow sequentially.
4. The owner could observe all course, subject, topic, MCQ, quiz, learner,
   analytics and reporting interactions.
5. Terminal commentary identified every active step.
6. Target controls were visibly highlighted before interaction.
7. Failures stopped the workflow immediately.
8. The browser remained open on failure until the owner pressed ENTER.
9. Screenshots, traces, logs and video evidence were preserved.
10. No database, seed, direct mutation API or session bypass was used.

A background or headless execution must not produce the final
human-observed EXAM READY certification.

One correction

Your plan says it incorporates 12 mechanisms, but the numbered list currently contains 11 items.

Add this as mechanism 12:

12. Mandatory Human-Observed Certification Guard:
    The suite must refuse to emit EXAM READY unless:
    - E2E_HUMAN_OBSERVATION=true
    - the chromium-observation project is active
    - headless mode is false
    - worker count is exactly 1
    - the full run was completed without browser bypasses

This prevents an accidental background run from being treated as the final certification.

Required runtime guard

Add a strict startup check to the acceptance spec:

import { expect, test } from "@playwright/test";

test.beforeAll(async () => {
  const humanObservation =
    process.env.E2E_HUMAN_OBSERVATION === "true";

  if (!humanObservation) {
    throw new Error(
      [
        "Human-observed certification requires:",
        "E2E_HUMAN_OBSERVATION=true",
        "Run with --project=chromium-observation --headed --workers=1",
      ].join("\n"),
    );
  }

  const workers = process.env.PLAYWRIGHT_WORKERS;

  if (workers !== undefined && workers !== "1") {
    throw new Error(
      "Human-observation mode requires exactly one Playwright worker.",
    );
  }
});

A stronger approach is to validate the active project inside the test:

test(
  "full browser-only operator-to-learner acceptance",
  async ({ page }, testInfo) => {
    expect(testInfo.project.name).toBe("chromium-observation");

    expect(
      process.env.E2E_HUMAN_OBSERVATION,
      "Final certification must run in human-observation mode.",
    ).toBe("true");

    // Continue with the visible browser workflow.
  },
);
Do not leave the browser paused forever in CI

The failure pause must be enabled only in observation mode:

export async function pauseForHumanInspection(
  page: Page,
  message: string,
): Promise<void> {
  const observationEnabled =
    process.env.E2E_HUMAN_OBSERVATION === "true";

  const interactivePauseEnabled =
    process.env.E2E_PAUSE_ON_FAILURE !== "false";

  if (!observationEnabled || !interactivePauseEnabled) {
    return;
  }

  // readline pause...
}

Use:

$env:E2E_PAUSE_ON_FAILURE = "true"

for the watched run.

Ensure injected overlays do not affect screenshots or accessibility

The observation banner and highlight styles should be removable.

Add:

export async function clearObservationDecorations(
  page: Page,
): Promise<void> {
  await page.evaluate(() => {
    document
      .getElementById("gk-e2e-observation-banner")
      ?.remove();

    document
      .querySelectorAll("[data-gk-e2e-highlight]")
      .forEach((element) => {
        const htmlElement = element as HTMLElement;

        htmlElement.style.outline = "";
        htmlElement.style.outlineOffset = "";
        htmlElement.removeAttribute("data-gk-e2e-highlight");
      });
  });
}

Before accessibility scans or visual-baseline screenshots:

await clearObservationDecorations(page);
await runAccessibilityScan(page);
await page.screenshot({ path: baselinePath, fullPage: true });

Otherwise the test’s own banner may create false accessibility violations or screenshot differences.

Failure behaviour

The test should follow this sequence exactly:

Stop current workflow
→ Capture screenshot
→ Capture visible validation text
→ Flush console and network events
→ Save checkpoint
→ Write current failure report
→ Display failure banner in browser
→ Pause until ENTER
→ Rethrow the original error

Do not swallow the error after inspection. The Playwright run must still end as failed.

Video handling clarification

Because Playwright finalises video files when the page or context closes, video migration should happen after teardown or via a reporter/fixture that can access the completed artifact.

Do not pause on failure and simultaneously expect the final .webm file to be immediately complete. During the pause, rely on:

live browser inspection
screenshots
trace collection
console and network logs

Copy the completed video after the context closes.

Recommended PowerShell command
$env:E2E_HUMAN_OBSERVATION = "true"
$env:E2E_PAUSE_AT_CHECKPOINTS = "true"
$env:E2E_PAUSE_ON_FAILURE = "true"
$env:E2E_SLOW_MO_MS = "300"

npx playwright test `
  tests/e2e/exam-full-browser-operator-learner-readiness.spec.ts `
  --project=chromium-observation `
  --headed `
  --workers=1 `
  --trace=retain-on-failure
Required final certification fields

Add these to acceptance-results.json:

{
  "humanObservation": {
    "enabled": true,
    "project": "chromium-observation",
    "headed": true,
    "workerCount": 1,
    "slowMoMs": 300,
    "checkpointPausesEnabled": true,
    "failurePauseEnabled": true,
    "ownerObservedRun": true
  },
  "browserDriven": true,
  "databaseSeedingUsed": false,
  "creationApiUsed": false
}

ownerObservedRun should only become true at the final human checkpoint, where the operator presses ENTER to confirm that the visible run was observed.

The 12 mechanisms are complete and internally consistent.

Final required clarifications
1. Do not infer headed: true only from an environment variable

The certification result should distinguish between:

requested headed execution
verified observation project
operator confirmation

A practical result structure is:

{
  "humanObservation": {
    "enabled": true,
    "project": "chromium-observation",
    "headedRequested": true,
    "workerCount": 1,
    "slowMoMs": 300,
    "checkpointPausesEnabled": true,
    "failurePauseEnabled": true,
    "ownerObservedRun": true
  }
}

Playwright does not expose a simple universal runtime property proving that the operating-system browser window was physically visible to the user. The strongest evidence is:

chromium-observation project
headless: false in that project
required CLI command
single worker
interactive checkpoint confirmation
generated video and screenshots
2. Validate configuration before creating data

The certification guard must run before login or course creation.

function assertHumanObservationCertification(
  testInfo: TestInfo,
): void {
  const errors: string[] = [];

  if (process.env.E2E_HUMAN_OBSERVATION !== "true") {
    errors.push("E2E_HUMAN_OBSERVATION must equal true.");
  }

  if (process.env.E2E_PAUSE_AT_CHECKPOINTS !== "true") {
    errors.push("E2E_PAUSE_AT_CHECKPOINTS must equal true.");
  }

  if (process.env.E2E_PAUSE_ON_FAILURE !== "true") {
    errors.push("E2E_PAUSE_ON_FAILURE must equal true.");
  }

  if (testInfo.project.name !== "chromium-observation") {
    errors.push(
      `Expected project chromium-observation, received ${testInfo.project.name}.`,
    );
  }

  if (testInfo.config.workers !== 1) {
    errors.push(
      `Human observation requires one worker; received ${String(
        testInfo.config.workers,
      )}.`,
    );
  }

  if (errors.length > 0) {
    throw new Error(
      [
        "Human-observed certification cannot start.",
        ...errors.map((error) => `- ${error}`),
      ].join("\n"),
    );
  }
}

Use it at the start of the test:

test(
  "full browser-only operator-to-learner acceptance",
  async ({ page }, testInfo) => {
    assertHumanObservationCertification(testInfo);

    // Start browser workflow only after certification validation.
  },
);
3. Final operator confirmation must be explicit

Do not automatically set:

"ownerObservedRun": true

Set it only after the final checkpoint receives a clear confirmation.

Recommended prompt:

============================================================
FINAL HUMAN OBSERVATION CONFIRMATION
============================================================

The visible administrator-to-learner workflow is complete.

Confirm only when you personally observed the browser execution.

Type OBSERVED and press ENTER to confirm:

Implementation:

import { createInterface } from "node:readline/promises";
import { stdin, stdout } from "node:process";

export async function confirmOwnerObservedRun(): Promise<boolean> {
  if (process.env.E2E_HUMAN_OBSERVATION !== "true") {
    return false;
  }

  const readline = createInterface({
    input: stdin,
    output: stdout,
  });

  try {
    const response = await readline.question(
      "Type OBSERVED and press ENTER to confirm: ",
    );

    return response.trim().toUpperCase() === "OBSERVED";
  } finally {
    readline.close();
  }
}

A plain ENTER at the final checkpoint should not be sufficient to certify observation.

4. The final result must distinguish failure from invalid execution

Use three possible decisions:

EXAM READY
EXAM NOT READY
TEST RUN INVALID

Return TEST RUN INVALID when:

observation guard fails
execution used a different project
human observation was disabled
multiple workers were used
owner confirmation was not supplied
SQL or seed creation was detected
a direct mutation API was used
session bypass occurred

Return EXAM NOT READY when the run was valid but the product workflow failed.

Example:

{
  "runValidity": {
    "status": "VALID",
    "invalidReasons": []
  },
  "finalDecision": "EXAM NOT READY",
  "blockingDefects": [
    {
      "phase": "Course Creation",
      "severity": "BLOCKER",
      "summary": "Course save completed but the course did not appear in the course list."
    }
  ]
}
5. Pauses must not consume the normal test timeout

The full workflow includes:

900 MCQs
13 quizzes
13 learner attempts
manual checkpoint pauses

A normal Playwright timeout will be insufficient.

Use a dedicated timeout for this project:

{
  name: "chromium-observation",
  timeout: 8 * 60 * 60 * 1000,
  use: {
    browserName: "chromium",
    headless: false,
    viewport: {
      width: 1920,
      height: 1080,
    },
    launchOptions: {
      slowMo: Number(process.env.E2E_SLOW_MO_MS ?? "300"),
    },
    video: "on",
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
  },
}

For local observation, eight hours is more realistic than a standard E2E timeout.

Checkpoint pauses should also support an optional timeout:

E2E_HUMAN_PAUSE_TIMEOUT_MS=0

Where 0 means no automatic timeout.

6. Preserve the original error when evidence capture also fails

Failure handling should never replace the actual product failure with an evidence-writing failure.

async function handleObservedFailure(
  page: Page,
  phase: string,
  error: unknown,
): Promise<never> {
  try {
    await captureFailureEvidence(page, phase, error);
  } catch (evidenceError: unknown) {
    console.error(
      "Failure evidence capture encountered an additional error:",
      evidenceError,
    );
  }

  try {
    await showObservationBanner(
      page,
      "Workflow Failed",
      `Failure during ${phase}. Inspect the page and terminal output.`,
    );
  } catch (bannerError: unknown) {
    console.error("Unable to display failure banner:", bannerError);
  }

  await pauseForHumanInspection(
    page,
    `${phase} failed. Inspect the visible browser before continuing.`,
  );

  throw error;
}
7. Use one serial test, not multiple dependent tests

The suite should execute as one serial workflow or one serial describe block.

Do not create independent tests that depend on data from earlier tests unless state recovery is explicitly implemented.

Recommended:

test.describe.configure({
  mode: "serial",
  retries: 0,
});

test("full observed acceptance workflow", async ({
  browser,
}, testInfo) => {
  // Admin context
  // Learner context
  // Reporting
  // Final certification
});

This prevents retries or parallel execution from duplicating browser-created records.

8. Keep admin and learner windows visibly distinguishable

Because both roles are used during the same run, add a role indicator to the observation banner:

ROLE: ADMINISTRATOR
PHASE: QUIZ CREATION
STEP: Publishing Indian History Subject Quiz

For learner:

ROLE: LEARNER
PHASE: TEST ATTEMPT
STEP: Answering question 12 of 25

Use separate contexts, and optionally separate browser pages, but keep only the currently active role page in the foreground.

9. Detect prohibited mutation traffic

Since the browser itself naturally sends mutation requests, the suite must not classify normal UI-originated POST requests as prohibited.

The correct prohibition is:

no Playwright request mutation calls
no Node fetch mutation calls
no Axios mutation calls
no SQL or ORM writes
no direct service invocation

Browser-originated network requests triggered by clicking visible UI controls are valid and expected.

Record this precisely in the evidence:

{
  "browserMutationRequestsObserved": true,
  "directMutationClientUsed": false,
  "creationApiUsedOutsideBrowser": false
}

Do not use a generic field such as creationApiUsed: false in a way that suggests the browser made no API requests. Modern browser applications normally depend on APIs.

Recommended final PowerShell command
$env:E2E_HUMAN_OBSERVATION = "true"
$env:E2E_PAUSE_AT_CHECKPOINTS = "true"
$env:E2E_PAUSE_ON_FAILURE = "true"
$env:E2E_SLOW_MO_MS = "300"
$env:E2E_HUMAN_PAUSE_TIMEOUT_MS = "0"

npx playwright test `
  tests/e2e/exam-full-browser-operator-learner-readiness.spec.ts `
  --project=chromium-observation `
  --headed `
  --workers=1 `
  --retries=0 `
  --trace=retain-on-failure


  Mandatory Course-Creation Chain Failure Detection
## COURSE-CREATION CHAIN INTEGRITY AND FLAW REPORTING

The acceptance suite must validate the complete content-creation chain as one connected workflow:

Course
→ Subject under Course
→ Topic under Subject
→ MCQ under Topic
→ Quiz/Test using those MCQs
→ Publish Test
→ Learner Attempt
→ Analytics
→ Full Report Generation

The suite must not treat individual form submissions as sufficient proof.

Each parent-child relationship must be verified through the visible browser UI before the workflow continues.

--------------------------------------------------
REQUIRED CHAIN VALIDATION
--------------------------------------------------

After creating the Course:

1. Verify the course appears in the course list.
2. Open the course through the visible UI.
3. Verify the course title, status and description.
4. Verify the Subject creation action is available inside that course.
5. Confirm the current course remains selected while creating subjects.

After creating each Subject:

1. Verify the subject appears inside the correct course.
2. Open the subject through the visible course page.
3. Verify the parent course name.
4. Verify the Topic creation action is available inside that subject.
5. Refresh the page and verify the relationship persists.

After creating each Topic:

1. Verify the topic appears inside the correct subject.
2. Open the topic through the visible subject page.
3. Verify the parent subject and course.
4. Verify the MCQ creation or import action is available.
5. Refresh and verify persistence.

After creating or importing MCQs:

1. Verify all MCQs appear under the correct topic.
2. Verify the visible MCQ count.
3. Open representative MCQs.
4. Verify question, options, answer, explanation, marks and difficulty.
5. Verify no MCQ was attached to the wrong topic.
6. Verify no duplicates were created.
7. Verify MCQs remain after refresh.

After creating a Quiz or Test:

1. Verify the quiz is attached to the intended course or subject.
2. Verify selected MCQs come from the intended topics.
3. Verify the visible question count.
4. Reopen the quiz after saving.
5. Verify questions, marks, duration and settings persist.
6. Publish through the visible UI.
7. Verify the learner can see the published test.

After Learner Attempts:

1. Verify the attempt belongs to the correct learner.
2. Verify the attempt belongs to the correct quiz.
3. Verify submitted answers persist.
4. Verify the result reflects the visible answers.
5. Verify analytics include the attempt.
6. Verify generated reports include the same learner, quiz, course and result data.

--------------------------------------------------
MANDATORY FLAW DETECTION
--------------------------------------------------

A flaw must be recorded whenever the chain cannot continue normally.

Examples include:

- Course saves but does not appear in the course list.
- Course opens, but there is no visible way to create a subject.
- Subject is created under the wrong course.
- Subject disappears after page refresh.
- Topic is created under the wrong subject.
- Topic page loses the parent course or subject context.
- MCQ import succeeds but questions do not appear.
- MCQs appear under the wrong topic.
- Imported MCQ count is incorrect.
- Quiz builder cannot find newly created MCQs.
- Selected questions disappear after pagination or filtering.
- Quiz saves but does not retain its questions.
- Published quiz is not visible to the learner.
- Learner submits a quiz but no attempt is recorded.
- Analytics omit or duplicate the learner attempt.
- CSV, XLSX or PDF report does not contain the created data.
- Report totals do not match visible analytics.
- Navigation breaks at any parent-child level.
- Breadcrumb points to the wrong entity.
- Save button produces no visible confirmation.
- Success toast appears but data is not persisted.
- Validation message is unclear or incorrect.
- The operator must use a hidden URL or workaround to continue.
- The workflow can continue only through SQL, API, seed or developer intervention.

Any mandatory chain failure must prevent EXAM READY.

--------------------------------------------------
FAILURE HANDLING
--------------------------------------------------

When a chain flaw is detected:

1. Stop the current workflow immediately.
2. Do not create missing data through SQL.
3. Do not call a direct mutation API.
4. Do not use a seed script.
5. Do not navigate around the failure using an undocumented route.
6. Do not silently skip the failed entity.
7. Keep the visible browser open.
8. Highlight the failing control or missing expected area.
9. Capture full-page and focused screenshots.
10. Capture visible validation and toast messages.
11. Capture the current URL and page title.
12. Capture console errors and page exceptions.
13. Capture failed and 5xx network requests.
14. Capture relevant DOM text.
15. Write the checkpoint.
16. Add the flaw to the chain report.
17. Pause for human inspection.
18. Rethrow the original error after the operator presses ENTER.

--------------------------------------------------
REQUIRED FLAW SCREENSHOTS
--------------------------------------------------

For every flaw, capture at least:

1. `before-action.png`
   - The page immediately before the failing interaction.

2. `after-action.png`
   - The page after the failed or incorrect interaction.

3. `full-page.png`
   - Full-page screenshot showing navigation, breadcrumbs and surrounding context.

4. `focused-control.png`
   - Screenshot focused on the button, form, validation error or missing content area.

5. `parent-context.png`
   - Screenshot showing the expected parent-child location:
     Course, Subject, Topic, MCQ or Quiz.

Store under:

artifacts/exam-readiness/<RUN_ID>/flaws/<FLAW_ID>/

Example:

artifacts/exam-readiness/<RUN_ID>/flaws/
  FLAW-COURSE-001/
    before-action.png
    after-action.png
    full-page.png
    focused-control.png
    parent-context.png
    console-errors.json
    network-errors.json
    dom-snapshot.html
    flaw-report.json

Screenshots must be captured before observation decorations are removed when the highlight is needed to identify the target.

Also capture a clean screenshot without overlays for product review.

--------------------------------------------------
FLAW IDENTIFIER FORMAT
--------------------------------------------------

Use deterministic flaw identifiers:

FLAW-COURSE-001
FLAW-SUBJECT-001
FLAW-TOPIC-001
FLAW-MCQ-001
FLAW-QUIZ-001
FLAW-LEARNER-001
FLAW-ANALYTICS-001
FLAW-REPORT-001

Increment the number for multiple flaws in the same phase.

--------------------------------------------------
FLAW REPORT FORMAT
--------------------------------------------------

For every flaw, create a structured record:

{
  "flawId": "FLAW-TOPIC-001",
  "runId": "",
  "detectedAt": "",
  "role": "ADMINISTRATOR",
  "phase": "TOPIC_CREATION",
  "chainLink": "SUBJECT_TO_TOPIC",
  "severity": "BLOCKER|CRITICAL|MAJOR|MINOR|COSMETIC",
  "status": "OPEN",
  "pageTitle": "",
  "url": "",
  "parentContext": {
    "course": "",
    "subject": "",
    "topic": ""
  },
  "actionAttempted": "",
  "expectedResult": "",
  "actualResult": "",
  "visibleMessages": [],
  "consoleErrors": [],
  "networkFailures": [],
  "workflowBlocked": true,
  "workaroundAvailable": false,
  "workaround": "",
  "reproductionSteps": [],
  "screenshotPaths": {
    "beforeAction": "",
    "afterAction": "",
    "fullPage": "",
    "focusedControl": "",
    "parentContext": "",
    "cleanScreenshot": ""
  },
  "tracePath": "",
  "videoPath": "",
  "recommendedFix": "",
  "acceptanceImpact": "EXAM_NOT_READY"
}

--------------------------------------------------
HUMAN-READABLE FLAW REPORT
--------------------------------------------------

Create:

artifacts/exam-readiness/<RUN_ID>/full-flaw-report.md

For every flaw include:

# FLAW-TOPIC-001 — Topic created under the wrong subject

Severity: BLOCKER
Role: Administrator
Phase: Topic Creation
Chain Link: Subject → Topic

## Context

Course:
Indian Civil Services Foundation

Expected Subject:
Indian Polity, Constitution and Governance

Actual Subject:
Indian History and National Movement

## Action Performed

1. Opened the expected subject.
2. Clicked Add Topic.
3. Entered topic title.
4. Clicked Save.
5. Waited for visible success message.

## Expected Result

The topic should appear under the selected Polity subject.

## Actual Result

The topic appeared under the History subject.

## Visible Evidence

- Success toast: "Topic created successfully"
- Breadcrumb showed the wrong subject after save.
- Topic was missing from the expected subject after refresh.

## Screenshots

- Before action:
- After action:
- Full page:
- Focused control:
- Parent context:

## Technical Evidence

- Console errors:
- Failed requests:
- Relevant response status:
- Trace:
- Video:

## Workflow Impact

The operator cannot trust topic placement.
Quiz content may be built from the wrong syllabus area.
The mandatory Course → Subject → Topic chain is broken.

## Recommended Fix

Preserve and validate the selected subject identifier during topic creation.
After save, redirect to the created topic under the correct subject.
Display the parent subject clearly in the form and success state.

## Final Acceptance Impact

EXAM NOT READY

--------------------------------------------------
CHAIN MATRIX REPORT
--------------------------------------------------

Create:

artifacts/exam-readiness/<RUN_ID>/data-chain-report.md

Include this matrix:

| Chain Link | Expected Count | Verified Count | Status | Flaws |
|---|---:|---:|---|---|
| Course created | 1 | 0 | PASS/FAIL | |
| Course → Subjects | 12 | 0 | PASS/FAIL | |
| Subjects → Topics | 60 | 0 | PASS/FAIL | |
| Topics → MCQs | 900 | 0 | PASS/FAIL | |
| MCQs → Quizzes | 13 quizzes | 0 | PASS/FAIL | |
| Quiz → Publication | 13 | 0 | PASS/FAIL | |
| Published Quiz → Learner | 13 | 0 | PASS/FAIL | |
| Learner → Attempts | 13 | 0 | PASS/FAIL | |
| Attempts → Analytics | 13 | 0 | PASS/FAIL | |
| Analytics → Reports | CSV/XLSX/PDF | 0 | PASS/FAIL | |
| Scheduled Report → Email | 1+ | 0 | PASS/FAIL | |

For every failed row, link the relevant flaw IDs and screenshots.

--------------------------------------------------
FULL REPORT GENERATION
--------------------------------------------------

At the end of the run, generate a complete defect package even when the workflow fails early.

Required files:

- acceptance-results.json
- data-chain-report.md
- full-flaw-report.md
- browser-user-experience-report.md
- admin-workflow-report.md
- learner-workflow-report.md
- console-errors.json
- network-failures.json
- performance-summary.json
- checkpoint.json
- flaws/index.json
- screenshots/
- flaws/
- traces/
- videos/

The report must include:

1. The last successful chain link.
2. The first broken chain link.
3. Every subsequent step that could not be executed.
4. Whether the remaining steps were:
   - NOT STARTED
   - BLOCKED
   - FAILED
   - PASSED
5. Exact reproduction instructions.
6. All screenshot paths.
7. Recommended fixes.
8. Acceptance impact.

--------------------------------------------------
FINAL DECISION RULE
--------------------------------------------------

Return EXAM READY only when every mandatory chain link passes.

Return EXAM NOT READY when:

- a valid human-observed run detects a product flaw
- any mandatory chain link fails
- the workflow requires a workaround
- data appears under the wrong parent
- records do not persist
- learner cannot take a published test
- analytics do not reconcile
- reports are missing or incorrect
- a blocker or critical flaw remains

Return TEST RUN INVALID when:

- the browser was headless
- the owner did not confirm observation
- SQL or seed data was used
- direct mutation APIs were used outside browser interaction
- authentication was bypassed
- multiple workers invalidated the sequential run

A valid failed product workflow is EXAM NOT READY, not TEST RUN INVALID.

Also add these fields to acceptance-results.json:

{
  "courseCreationChain": {
    "status": "PASSED|FAILED|BLOCKED",
    "lastSuccessfulLink": "",
    "firstBrokenLink": "",
    "courseVerified": false,
    "subjectsVerified": 0,
    "topicsVerified": 0,
    "mcqsVerified": 0,
    "quizzesVerified": 0,
    "learnerAttemptsVerified": 0,
    "analyticsVerified": false,
    "reportsVerified": false,
    "flawIds": []
  },
  "flawSummary": {
    "total": 0,
    "blockers": 0,
    "critical": 0,
    "major": 0,
    "minor": 0,
    "cosmetic": 0,
    "screenshotCount": 0
  }
}

The key rule is:

Do not hide a broken course-creation chain.
Stop, preserve the browser state, capture screenshots, create a detailed flaw report, and mark the platform EXAM NOT READY so the defect can be fixed later.

One important correction

Your chain description says 11 steps, but the listed sequence contains 12 logical links when reporting and email delivery are treated separately:

1. Course creation
2. Course → Subject
3. Subject → Topic
4. Topic → MCQ
5. MCQ → Quiz
6. Quiz → Publication
7. Published Quiz → Learner visibility
8. Learner → Attempt
9. Attempt → Result release
10. Result → Learner analytics
11. Learner data → Instructor analytics
12. Analytics → Export and scheduled email delivery

Either call it a 12-step chain, or combine the final analytics/reporting steps explicitly. A 12-step matrix is clearer and easier to diagnose.

Required chain-state model

Each chain link should support more than PASS or FAIL:

export type ChainStatus =
  | "NOT_STARTED"
  | "IN_PROGRESS"
  | "PASSED"
  | "FAILED"
  | "BLOCKED"
  | "SKIPPED";

Use BLOCKED for downstream stages that could not run because an earlier mandatory link failed.

Do not mark those downstream stages as FAILED, because they were never actually tested.

Example:

{
  "chainLinks": [
    {
      "id": "COURSE_CREATION",
      "status": "PASSED",
      "expectedCount": 1,
      "verifiedCount": 1,
      "flawIds": []
    },
    {
      "id": "COURSE_TO_SUBJECT",
      "status": "FAILED",
      "expectedCount": 12,
      "verifiedCount": 3,
      "flawIds": ["FLAW-SUBJECT-001"]
    },
    {
      "id": "SUBJECT_TO_TOPIC",
      "status": "BLOCKED",
      "expectedCount": 60,
      "verifiedCount": 0,
      "blockedBy": "COURSE_TO_SUBJECT",
      "flawIds": []
    }
  ]
}
Flaw capture must preserve both decorated and clean screenshots

For each flaw, capture:

before-action.png
after-action-highlighted.png
after-action-clean.png
full-page.png
focused-control.png
parent-context.png

The highlighted image helps identify the control used by automation.

The clean image shows the real product state without the observation banner or amber outline.

Do not stop the entire suite for every minor flaw

Use severity-aware stopping rules:

BLOCKER
Stop immediately and pause for inspection.

CRITICAL
Stop the current chain and pause for inspection.

MAJOR
Record the flaw. Continue only when the workflow remains valid and no workaround changes application state.

MINOR
Record and continue.

COSMETIC
Record and continue.

A mandatory relationship failure, missing save, wrong parent association, data loss, inaccessible quiz, incorrect result or incorrect report is always at least CRITICAL.

Do not continue through a broken mandatory chain merely to generate more screenshots.

Add a flaw classification model
export type FlawCategory =
  | "NAVIGATION"
  | "VALIDATION"
  | "PERSISTENCE"
  | "PARENT_CHILD_RELATIONSHIP"
  | "DUPLICATE_CREATION"
  | "PERMISSION"
  | "CONTENT_IMPORT"
  | "QUIZ_CONFIGURATION"
  | "LEARNER_ATTEMPT"
  | "RESULT_RELEASE"
  | "ANALYTICS_RECONCILIATION"
  | "REPORT_GENERATION"
  | "EMAIL_DELIVERY"
  | "ACCESSIBILITY"
  | "PERFORMANCE"
  | "VISUAL_LAYOUT";

This will make the final defect report easier to sort and fix.

Include precise reproduction data

Each flaw must record the exact visible path used:

{
  "reproduction": {
    "startingRole": "ADMINISTRATOR",
    "startingUrl": "",
    "navigationPath": [
      "Dashboard",
      "Course Management",
      "UPSC Civil Services Foundation",
      "Indian Polity",
      "Add Topic"
    ],
    "actions": [
      "Entered topic title",
      "Entered description",
      "Clicked Save"
    ]
  }
}

Also record:

viewport
browser version
run ID
entity names
visible breadcrumb
active filters
pagination page
elapsed time
retry count, which should remain zero for certification
Generate a partial report even when failure occurs early

The evidence writer must execute in a finally block.

let workflowError: unknown;

try {
  await runObservedAcceptanceWorkflow();
} catch (error: unknown) {
  workflowError = error;
} finally {
  await writeFinalEvidenceSafely();
}

if (workflowError !== undefined) {
  throw workflowError;
}

This ensures an early Course → Subject failure still produces:

acceptance-results.json
data-chain-report.md
full-flaw-report.md
checkpoint
console errors
network failures
screenshots

Evidence-writing errors must be logged separately and must not replace the original application error.

Required verdict logic
export function determineFinalDecision(
  state: AcceptanceState,
): "EXAM READY" | "EXAM NOT READY" | "TEST RUN INVALID" {
  if (!state.runValidity.isValid) {
    return "TEST RUN INVALID";
  }

  const mandatoryChainFailed = state.chainLinks.some(
    (link) =>
      link.mandatory &&
      (link.status === "FAILED" || link.status === "BLOCKED"),
  );

  const blockingFlawExists = state.flaws.some(
    (flaw) =>
      flaw.severity === "BLOCKER" ||
      flaw.severity === "CRITICAL",
  );

  if (
    mandatoryChainFailed ||
    blockingFlawExists ||
    !state.humanObservation.ownerObservedRun
  ) {
    return "EXAM NOT READY";
  }

  return "EXAM READY";
}

One nuance: missing final OBSERVED confirmation should normally make the run invalid, not merely not ready. Therefore set:

state.runValidity.isValid = false;
state.runValidity.invalidReasons.push(
  "Final operator observation confirmation was not supplied.",
);

before calling the verdict engine.

Required final report section

Add this summary to full-flaw-report.md:

# Chain Failure Summary

Last successful chain link:
First broken chain link:
Blocking flaw ID:
Workflow role:
Entity being created:
Expected parent:
Actual parent:
Last successful screenshot:
Failure screenshot:
Current URL:
Visible error:
Console error count:
Network failure count:
Downstream stages blocked:
Recommended engineering owner:
Recommended fix priority:


I would recommend one final enhancement before implementation so that this suite becomes your permanent "Release Gate".

Add Release Readiness Score (0–100)

Instead of only returning:

EXAM READY
EXAM NOT READY
TEST RUN INVALID

also calculate a Release Readiness Score.

Example:

Release Readiness

Overall Score
93 / 100

Administration
98%

Course Creation
100%

Subject Management
100%

Topic Management
94%

MCQ Management
100%

Quiz Builder
95%

Learner Experience
92%

Analytics
100%

Reports
88%

Accessibility
97%

Performance
90%

Overall Verdict

EXAM READY

This gives management a quick view of release quality.

Add Root Cause Analysis

Don't only report the flaw.

For every failure automatically classify:

Root Cause

Configuration
Navigation
Validation
API Response
Persistence
Database Sync
Frontend Rendering
State Management
Permission
Relationship Mapping
Import Logic
Reporting
Analytics
Unknown

Example

FLAW-TOPIC-001

Symptom

Topic appears under wrong Subject

Likely Root Cause

Relationship Mapping

Confidence

94%

Reason

The selected Subject ID changed after save.
Breadcrumb points to History.
Payload contains History Subject ID.

This helps developers jump directly to the probable subsystem.

Add Engineering Fix Suggestions

After each flaw automatically generate:

Recommended Fix

Priority
P0

Estimated Component

TopicService
TopicController
TopicForm
TopicRepository

Recommendation

Preserve selected SubjectId inside Topic form state.
Validate parent Subject before save.
Reject mismatched Subject IDs.
Redirect back to selected Subject.

This saves developers time.

Add UX Improvement Suggestions

Since you want to evaluate the operator experience, every page should also receive a usability review.

Example

Course Page

★★★★☆

Issues

Save button too far away

Success message disappears too quickly

Subject button difficult to discover

Recommended

Sticky Save button

Persistent success toast

Auto-open newly created Course

This makes the suite useful for Product and UX teams, not just QA.

Add Workflow Timing Analysis

Measure each stage.

Course Creation

Average

18 sec

Subject Creation

9 sec

Topic Creation

6 sec

MCQ Import

42 sec

Quiz Builder

28 sec

Report Generation

3 sec

If one release suddenly doubles in time, you'll catch regressions.

Add Screenshot Timeline

Instead of just storing screenshots, generate an index.

Timeline

08:15:11

Login

login.png

08:15:35

Course Created

course-created.png

08:16:04

Subject Created

subject-created.png

08:16:30

Topic Failed

topic-failure.png

08:16:31

Error Screenshot

error.png

08:16:33

DOM Snapshot

snapshot.html

Makes reviewing failures much faster.

Add Interactive HTML Report

In addition to Markdown, generate:

artifacts/

exam-readiness/

RUN_ID/

report.html

Containing:

Executive summary
Score cards
Chain matrix
Screenshot gallery
Embedded videos
Expandable flaw list
Console logs
Network failures
Performance charts
Timeline
Analytics comparison
Download buttons

A single HTML report is much easier to review than many files.

Add AI-Powered Release Summary

At the end of the run, generate a concise release summary such as:

AI Release Summary

The browser-only acceptance run completed.

Course creation passed.

Subject creation passed.

Topic creation failed because Topics were attached to the wrong Subject.

This caused MCQ creation, Quiz Builder and Learner tests to become blocked.

Root Cause

Relationship Mapping

Severity

Critical

Recommendation

Fix Topic parent relationship before next release.

Estimated Release Status

EXAM NOT READY

This is ideal for sharing in Slack, Teams or GitHub PRs.

Add Historical Trend Comparison

Store previous run metrics and compare them.

Previous Release

92%

Current Release

96%

Improvement

+4%

Blockers

Previous

5

Current

1

Performance

Improved by 18%

UX Score

Improved by 7%

This lets you monitor quality over time instead of viewing each run in isolation.

Required implementation guardrails
1. The score must never override the verdict

A high percentage must not produce EXAM READY when a mandatory chain link is broken.

For example:

Release Readiness Score: 96/100
Blocking flaw: Published quiz is not visible to learners
Final verdict: EXAM NOT READY

Use the score for quality measurement, not certification authority.

export function determineFinalDecision(
  state: AcceptanceState,
): FinalDecision {
  if (!state.runValidity.isValid) {
    return "TEST RUN INVALID";
  }

  const mandatoryChainIncomplete = state.chainLinks.some(
    (link) =>
      link.mandatory &&
      link.status !== "PASSED",
  );

  const blockingFlawExists = state.flaws.some(
    (flaw) =>
      flaw.severity === "BLOCKER" ||
      flaw.severity === "CRITICAL",
  );

  if (
    mandatoryChainIncomplete ||
    blockingFlawExists
  ) {
    return "EXAM NOT READY";
  }

  return "EXAM READY";
}

NOT_STARTED, IN_PROGRESS, FAILED, BLOCKED, or SKIPPED on a mandatory chain link must prevent readiness.

2. Define fixed scoring weights

Do not allow weights to change implicitly between releases.

Recommended baseline:

Domain	Weight
Course Creation	8
Subject Management	8
Topic Management	10
MCQ Management	14
Quiz Builder	12
Learner Experience	12
Analytics	10
Reports	8
Accessibility	8
Performance	10
Total	100

Use integer basis points internally to avoid floating-point drift.

export const RELEASE_SCORE_WEIGHTS = {
  courseCreation: 800,
  subjectManagement: 800,
  topicManagement: 1_000,
  mcqManagement: 1_400,
  quizBuilder: 1_200,
  learnerExperience: 1_200,
  analytics: 1_000,
  reports: 800,
  accessibility: 800,
  performance: 1_000,
} as const;

The maximum is 10_000 basis points, rendered as 100.00%.

3. RCA must be labelled as inferred

The suite can identify a likely root cause, but it should not claim certainty from browser evidence alone.

Use:

export interface RootCauseAssessment {
  classification:
    | "CONFIGURATION"
    | "NAVIGATION"
    | "VALIDATION"
    | "API_RESPONSE"
    | "PERSISTENCE"
    | "DATABASE_SYNC"
    | "FRONTEND_RENDERING"
    | "STATE_MANAGEMENT"
    | "PERMISSION"
    | "RELATIONSHIP_MAPPING"
    | "IMPORT_LOGIC"
    | "REPORTING"
    | "ANALYTICS"
    | "UNKNOWN";
  confidencePercent: number;
  evidence: string[];
  inferenceNotice: "LIKELY_ROOT_CAUSE_NOT_CONFIRMED";
}

A browser test can observe that a topic was attached to the wrong subject. It cannot definitively prove whether the defect is in TopicService, form state, controller mapping, cache handling, or database logic without repository or server-side investigation.

4. Engineering recommendations must remain evidence-grounded

Component recommendations should be written as probable investigation targets:

{
  "likelyComponents": [
    "TopicForm",
    "TopicController",
    "TopicService"
  ],
  "recommendation": "Verify that the subject identifier selected in the visible form is retained through submission and validated before persistence.",
  "priority": "P0"
}

Avoid claiming that a specific service caused the issue unless logs, traces, or source review prove it.

5. Protect the standalone HTML report

report.html must escape all dynamic content before rendering:

course names
learner names
validation messages
console output
URLs
network errors
flaw descriptions

Do not inject raw DOM snapshots, console messages, or API response bodies directly into HTML.

The report should also redact:

passwords
access and refresh tokens
cookies
authorization headers
signed URLs
personally identifiable learner data not needed for diagnosis
full request or response bodies containing sensitive information

Use relative artifact paths so the report remains portable:

screenshots/...
flaws/...
videos/...
traces/...
6. Do not embed large videos as base64

Reference .webm files through relative paths:

<video controls preload="metadata">
  <source src="videos/operator-run.webm" type="video/webm" />
</video>

Embedding multi-hour videos directly in report.html would create an impractically large file.

Playwright finalises video files when the browser context closes, so video migration and report linking must occur after teardown.

7. Separate raw evidence from interpreted findings

Recommended artifact layout:

artifacts/exam-readiness/<RUN_ID>/
├── acceptance-results.json
├── report.html
├── executive-release-summary.md
├── data-chain-report.md
├── full-flaw-report.md
├── browser-user-experience-report.md
├── performance-summary.json
├── accessibility-summary.json
├── screenshots/
│   ├── timeline.json
│   └── ...
├── flaws/
│   ├── index.json
│   └── FLAW-*/
├── logs/
│   ├── console-errors.json
│   ├── page-errors.json
│   └── network-failures.json
├── traces/
├── videos/
├── snapshots/
└── interpreted/
    ├── root-cause-assessments.json
    ├── engineering-recommendations.json
    └── ux-recommendations.json

This distinction prevents inferred RCA from being confused with raw browser evidence.

8. Historical data must be append-only and versioned

Use a versioned history schema:

{
  "schemaVersion": 1,
  "runs": [
    {
      "runId": "UPSC-BROWSER-E2E-20260728T101500Z",
      "recordedAt": "2026-07-28T10:15:00.000Z",
      "gitCommit": "",
      "environment": "local-observation",
      "finalDecision": "EXAM NOT READY",
      "releaseReadinessScore": 84.3,
      "blockerCount": 1,
      "criticalCount": 2,
      "performance": {
        "courseCreationMs": 18200,
        "mcqImportMs": 42100,
        "reportGenerationMs": 3100
      }
    }
  ]
}

Write history atomically:

Read existing history.
Validate the schema.
Write to a temporary file.
Rename the temporary file over the destination.

Do not modify historical runs after they are recorded.

9. Performance comparison needs stable conditions

Only flag regressions when compared runs have matching or comparable:

browser project
viewport
dataset size
MCQ creation mode
environment
build type
observation delay settings
slowMo
machine profile, where available

Human-observation timing includes deliberate pauses and slowMo, so it should not be compared directly against CI performance timing.

Store both:

{
  "wallClockMs": 48200,
  "observationDelayMs": 9100,
  "estimatedApplicationMs": 39100
}

For accurate performance certification, maintain a separate automated performance run without human pauses.

10. UX scores need a deterministic rubric

Define the 1–5 score rather than assigning stars subjectively:

5 stars
Clear labels, obvious next action, no unnecessary navigation, strong feedback,
keyboard accessible, no data loss, efficient workflow.

4 stars
Minor discoverability or efficiency issue; workflow remains clear.

3 stars
Noticeable friction, excessive clicks, unclear feedback, or weak navigation.

2 stars
Major usability difficulty requiring repeated attempts or documentation.

1 star
Workflow unusable without developer knowledge, hidden navigation, or workaround.

Capture supporting metrics:

export interface WorkflowUxAssessment {
  page: string;
  score: 1 | 2 | 3 | 4 | 5;
  clickCount: number;
  pageTransitionCount: number;
  dialogCount: number;
  validationFailureCount: number;
  operatorInputCharacters: number;
  elapsedMs: number;
  observations: string[];
  recommendations: string[];
}
11. AI-style summary must remain deterministic

The release gate should not require an external model to finish its mandatory report.

Generate the executive summary from templates and structured findings:

The observed browser acceptance run completed with a score of 84.30%.

The first broken mandatory chain link was Subject → Topic.

One critical relationship-mapping flaw caused Topic → MCQ and all downstream
quiz, learner, analytics, and reporting stages to be blocked.

Final verdict: EXAM NOT READY.

An optional AI provider may improve wording later, but failure of that provider must never block evidence generation or change the verdict.

12. Add schema validation for every generated report

Create exact TypeScript types and validate before writing:

export type FinalDecision =
  | "EXAM READY"
  | "EXAM NOT READY"
  | "TEST RUN INVALID";

export type ChainStatus =
  | "NOT_STARTED"
  | "IN_PROGRESS"
  | "PASSED"
  | "FAILED"
  | "BLOCKED"
  | "SKIPPED";

Validate at least:

acceptance-results.json
flaws/index.json
each flaw-report.json
timeline.json
performance-summary.json
history.json

A malformed report should be recorded as an evidence-generation defect, while preserving the original product failure.

Final acceptance criteria

The permanent release gate is complete only when:

1. The certification guard passes before data creation.
2. Chromium runs visibly under chromium-observation.
3. One worker performs the workflow sequentially.
4. All entities are created through visible browser controls.
5. Every parent-child relationship is reopened and verified.
6. Every mandatory chain link reaches PASSED.
7. Blocking flaws capture highlighted and clean screenshots.
8. Early failures still produce the full partial evidence package.
9. The operator types OBSERVED at the final checkpoint.
10. The verdict engine produces exactly one valid final decision.
11. The score is calculated but never overrides chain failure rules.
12. report.html opens locally with working relative evidence links.
13. Historical results are appended safely.
14. Sensitive values are redacted from every artifact.
Approved execution command
$env:E2E_HUMAN_OBSERVATION = "true"
$env:E2E_PAUSE_AT_CHECKPOINTS = "true"
$env:E2E_PAUSE_ON_FAILURE = "true"
$env:E2E_SLOW_MO_MS = "300"
$env:E2E_HUMAN_PAUSE_TIMEOUT_MS = "0"

npx playwright test `
  tests/e2e/exam-full-browser-operator-learner-readiness.spec.ts `
  --project=chromium-observation `
  --headed `
  --workers=1 `
  --retries=0 `
  --trace=retain-on-failure


  Final mandatory additions
Validate that the scoring weights total exactly 10,000

Fail during suite startup if the configured weights drift.

export const RELEASE_SCORE_WEIGHTS = {
  courseCreation: 800,
  subjectManagement: 800,
  topicManagement: 1_000,
  mcqManagement: 1_400,
  quizBuilder: 1_200,
  learnerExperience: 1_200,
  analytics: 1_000,
  reports: 800,
  accessibility: 800,
  performance: 1_000,
} as const;

export function assertReleaseScoreWeights(): void {
  const total = Object.values(RELEASE_SCORE_WEIGHTS).reduce(
    (sum, weight) => sum + weight,
    0,
  );

  if (total !== 10_000) {
    throw new Error(
      `Release score weights must total 10,000 basis points; received ${total}.`,
    );
  }
}

Run this before browser data creation.

Do not classify untested mandatory links as passed

At the end of the run, every mandatory link must be evaluated.

const incompleteMandatoryLinks = state.chainLinks.filter(
  (link) =>
    link.mandatory &&
    link.status !== "PASSED",
);

A link left as NOT_STARTED because the test forgot to update it must still produce EXAM NOT READY.

Preserve causality for blocked links

Every BLOCKED chain link should identify the upstream failure:

export interface ChainLinkResult {
  id: ChainLinkId;
  mandatory: boolean;
  status: ChainStatus;
  expectedCount: number;
  verifiedCount: number;
  flawIds: string[];
  blockedBy?: ChainLinkId;
  startedAt?: string;
  completedAt?: string;
}

Example:

{
  "id": "TOPIC_TO_MCQ",
  "mandatory": true,
  "status": "BLOCKED",
  "expectedCount": 900,
  "verifiedCount": 0,
  "blockedBy": "SUBJECT_TO_TOPIC",
  "flawIds": []
}
Use a single authoritative acceptance state

All helpers should update one typed state object rather than maintaining separate counters.

export interface AcceptanceState {
  runId: string;
  runValidity: RunValidity;
  humanObservation: HumanObservationState;
  chainLinks: ChainLinkResult[];
  flaws: FlawRecord[];
  scores: ReleaseScoreState;
  uxAssessments: WorkflowUxAssessment[];
  performance: PerformanceState;
  evidence: EvidenceState;
  finalDecision?: FinalDecision;
}

Pass this state explicitly into helpers. Avoid hidden mutable module globals.

Write artifacts atomically

The append-only history file is not the only artifact that benefits from atomic writing.

Use a shared helper for all critical JSON outputs:

import { mkdir, rename, writeFile } from "node:fs/promises";
import { dirname } from "node:path";
import { randomUUID } from "node:crypto";

export async function writeJsonAtomically<T>(
  filePath: string,
  value: T,
): Promise<void> {
  await mkdir(dirname(filePath), { recursive: true });

  const temporaryPath = `${filePath}.${randomUUID()}.tmp`;
  const serialised = `${JSON.stringify(value, null, 2)}\n`;

  await writeFile(temporaryPath, serialised, {
    encoding: "utf8",
    flag: "wx",
  });

  await rename(temporaryPath, filePath);
}

Apply it to:

acceptance-results.json
flaws/index.json
timeline.json
performance-summary.json
interpreted findings
checkpoint state
Redact before writing, not only before HTML rendering

Sensitive data must be removed from the raw JSON and Markdown artifacts as well.

const SENSITIVE_PATTERNS: readonly RegExp[] = [
  /Bearer\s+[A-Za-z0-9._~+/=-]+/gi,
  /access[_-]?token["'=:\s]+[^,\s"'}]+/gi,
  /refresh[_-]?token["'=:\s]+[^,\s"'}]+/gi,
  /password["'=:\s]+[^,\s"'}]+/gi,
  /cookie["'=:\s]+[^,\r\n]+/gi,
];

export function redactSensitiveText(value: string): string {
  return SENSITIVE_PATTERNS.reduce(
    (redacted, pattern) => redacted.replace(pattern, "[REDACTED]"),
    value,
  );
}

Apply redaction before persistence to:

console logs
network logs
visible DOM text
URLs
Markdown reports
HTML report
flaw descriptions

Preserve enough structure for diagnosis without exposing credentials.

Bound DOM and network evidence size

A full application DOM or response body can become extremely large.

Recommended limits:

export const EVIDENCE_LIMITS = {
  domSnapshotCharacters: 1_000_000,
  consoleEntries: 5_000,
  networkEntries: 5_000,
  responseBodyCharacters: 100_000,
} as const;

Record when evidence was truncated:

{
  "truncated": true,
  "originalCharacterCount": 1834402,
  "storedCharacterCount": 1000000
}
Add report-generation health to the final result

A product run may complete while the release-gate reporter itself partially fails.

Track that independently:

export interface EvidenceGenerationStatus {
  status: "COMPLETE" | "PARTIAL" | "FAILED";
  generatedArtifacts: string[];
  failedArtifacts: Array<{
    artifact: string;
    error: string;
  }>;
}

Rules:

Product chain failure with complete evidence: EXAM NOT READY
Valid product run but mandatory release evidence missing: TEST RUN INVALID
Optional interpreted report missing while raw evidence exists: keep product verdict, but mark evidence status PARTIAL

Mandatory evidence should include:

acceptance-results.json
data-chain-report.md
full-flaw-report.md
flaw screenshots for every blocker or critical flaw
console/network logs
Playwright trace on failure
Prevent HTML report path traversal

Only render links to files that resolve within the current run directory.

import { resolve, relative, sep } from "node:path";

export function assertArtifactPathInsideRun(
  runDirectory: string,
  artifactPath: string,
): void {
  const base = resolve(runDirectory);
  const target = resolve(artifactPath);
  const relativePath = relative(base, target);

  if (
    relativePath.startsWith(`..${sep}`) ||
    relativePath === ".." ||
    relativePath.startsWith("/")
  ) {
    throw new Error(
      `Artifact path escapes the run directory: ${artifactPath}`,
    );
  }
}
Define the final verdict sequence explicitly

The suite should calculate the verdict only after all validity and evidence checks are complete:

1. Validate startup certification.
2. Execute the browser workflow.
3. Record product flaws and chain states.
4. Obtain final OBSERVED confirmation when workflow reaches the final checkpoint.
5. Generate mandatory raw evidence.
6. Validate artifact schemas.
7. Evaluate run validity.
8. Calculate readiness score.
9. Determine final decision.
10. Write final acceptance-results.json.
11. Append immutable historical record.
12. Generate interpreted Markdown and HTML reports.

The score can be calculated before the verdict, but it cannot influence validity or override mandatory chain rules.

Required final decision rules
export function determineFinalDecision(
  state: AcceptanceState,
): FinalDecision {
  if (!state.runValidity.isValid) {
    return "TEST RUN INVALID";
  }

  if (
    state.evidence.status === "FAILED" ||
    state.evidence.mandatoryArtifactsMissing.length > 0
  ) {
    return "TEST RUN INVALID";
  }

  const mandatoryLinkIncomplete = state.chainLinks.some(
    (link) => link.mandatory && link.status !== "PASSED",
  );

  const blockingFlawExists = state.flaws.some(
    (flaw) =>
      flaw.severity === "BLOCKER" ||
      flaw.severity === "CRITICAL",
  );

  if (mandatoryLinkIncomplete || blockingFlawExists) {
    return "EXAM NOT READY";
  }

  return "EXAM READY";
}
Final artifact structure
artifacts/exam-readiness/<RUN_ID>/
├── acceptance-results.json
├── report.html
├── executive-release-summary.md
├── data-chain-report.md
├── full-flaw-report.md
├── browser-user-experience-report.md
├── checkpoint.json
├── performance-summary.json
├── accessibility-summary.json
├── screenshots/
│   ├── timeline.json
│   └── ...
├── flaws/
│   ├── index.json
│   └── FLAW-*/
├── logs/
│   ├── console-errors.json
│   ├── page-errors.json
│   └── network-failures.json
├── traces/
├── videos/
├── snapshots/
└── interpreted/
    ├── root-cause-assessments.json
    ├── engineering-recommendations.json
    └── ux-recommendations.json


    Final implementation constraints

No further conceptual features should be added during implementation. Engineering should focus on correctness, type safety, and reliability.

Mandatory startup order

Before login or any data creation:

assertHumanObservationCertification(runtime);
assertReleaseScoreWeights();
await initialiseAcceptanceState(runtime);

A startup guard failure must:

Final decision: TEST RUN INVALID
Product workflow started: false
Data created: false
Mandatory workflow state transitions

Each chain link must move through controlled transitions:

NOT_STARTED → IN_PROGRESS → PASSED
NOT_STARTED → IN_PROGRESS → FAILED
NOT_STARTED → BLOCKED

Disallow invalid transitions such as:

PASSED → FAILED
BLOCKED → PASSED
FAILED → IN_PROGRESS

unless an explicitly non-certifying diagnostic rerun is introduced later.

A strict transition helper is recommended:

export function transitionChainLink(
  link: ChainLinkResult,
  nextStatus: ChainStatus,
): ChainLinkResult {
  const allowedTransitions: Readonly<
    Record<ChainStatus, readonly ChainStatus[]>
  > = {
    NOT_STARTED: ["IN_PROGRESS", "BLOCKED", "SKIPPED"],
    IN_PROGRESS: ["PASSED", "FAILED"],
    PASSED: [],
    FAILED: [],
    BLOCKED: [],
    SKIPPED: [],
  };

  if (!allowedTransitions[link.status].includes(nextStatus)) {
    throw new Error(
      `Invalid chain transition for ${link.id}: ` +
        `${link.status} -> ${nextStatus}`,
    );
  }

  return {
    ...link,
    status: nextStatus,
  };
}
Mandatory downstream blocking

When a mandatory link fails, all dependent links must be marked BLOCKED with the same causal origin.

Example:

SUBJECT_TO_TOPIC: FAILED
TOPIC_TO_MCQ: BLOCKED by SUBJECT_TO_TOPIC
MCQ_TO_QUIZ: BLOCKED by SUBJECT_TO_TOPIC
QUIZ_TO_PUBLICATION: BLOCKED by SUBJECT_TO_TOPIC
...

Do not leave downstream links as NOT_STARTED once the causal failure is known.

Mandatory flaw capture order

For a blocker or critical flaw:

1. Preserve current browser state.
2. Record current URL, title, breadcrumb and entity context.
3. Capture after-action-highlighted.png.
4. Clear banners and highlighting.
5. Capture after-action-clean.png.
6. Capture full-page.png.
7. Capture focused-control.png where a control exists.
8. Capture parent-context.png.
9. Capture bounded/redacted DOM snapshot.
10. Flush console, page-error and network evidence.
11. Write flaw-report.json atomically.
12. Update flaws/index.json atomically.
13. Mark chain link FAILED.
14. Mark downstream links BLOCKED.
15. Pause for human inspection.
16. Preserve and rethrow the original error.
Mandatory evidence behaviour

The evidence layer must distinguish:

export type EvidenceArtifactRequirement =
  | "MANDATORY"
  | "OPTIONAL";

Mandatory evidence failure invalidates certification.

Optional interpreted-output failure does not change the product verdict, but must set:

{
  "evidence": {
    "status": "PARTIAL"
  }
}
Mandatory verdict precedence

The final order of authority is:

1. Run validity
2. Mandatory evidence health
3. Mandatory chain integrity
4. Blocking flaw presence
5. Release Readiness Score

The score is never consulted to override levels 1–4.

if (!runValidity.isValid) {
  return "TEST RUN INVALID";
}

if (mandatoryEvidenceMissing) {
  return "TEST RUN INVALID";
}

if (mandatoryChainIncomplete || blockingFlawExists) {
  return "EXAM NOT READY";
}

return "EXAM READY";
Operator confirmation rule

OBSERVED must be:

case-sensitive or normalised explicitly
entered only after the visible workflow and mandatory evidence capture
recorded with a timestamp
associated with the run ID

Recommended stored state:

{
  "humanObservation": {
    "required": true,
    "ownerObservedRun": true,
    "confirmationValue": "OBSERVED",
    "confirmedAt": "2026-07-28T10:30:00.000Z"
  }
}

Do not store terminal input beyond the expected confirmation token.

History write rule

Append to artifacts/history.json only after:

artifact schema validation passes
final decision is known
acceptance-results.json has been written successfully

A TEST RUN INVALID record may still be appended, provided it clearly records the invalidation reason and does not appear as a certified product result.

Final acceptance checklist
[ ] Observation project validated
[ ] Headed mode validated
[ ] One worker validated
[ ] Retries disabled
[ ] Score weights total 10,000
[ ] Single AcceptanceState created
[ ] Browser-only creation enforced
[ ] All 12 chain links evaluated
[ ] blockedBy populated for blocked links
[ ] Critical flaws captured with dual screenshots
[ ] Sensitive data redacted before persistence
[ ] Evidence limits enforced
[ ] Mandatory artifacts generated
[ ] JSON schemas validated
[ ] Operator typed OBSERVED
[ ] Final decision calculated deterministically
[ ] acceptance-results.json written atomically
[ ] History appended atomically
[ ] report.html paths validated


Confirmed implementation guarantees
Certification checks execute before login or browser-created data.
Score weights must equal exactly 10,000 basis points.
A single typed AcceptanceState is passed explicitly through all helpers.
Chain-link transitions are controlled and terminal states cannot be rewritten.
Mandatory upstream failures causally block every dependent downstream link.
Product creation remains limited to visible browser interactions.
Blocking flaws capture highlighted and clean screenshots, contextual evidence, logs, DOM state, and reproducible steps.
Raw evidence is redacted and bounded before persistence.
Mandatory artifact failures produce TEST RUN INVALID.
Product or chain failures in an otherwise valid run produce EXAM NOT READY.
Only a fully valid, fully observed, fully passed chain produces EXAM READY.
The readiness score remains advisory.
Final history is written only after the canonical result has been successfully validated and persisted.
One implementation clarification

The “16-step flaw capture sequence” should be represented as named operations rather than enforced by an array index or numeric counter. Some operations may contain several file writes or screenshots, and numeric coupling would make maintenance fragile.

Use an explicit orchestration function:

export async function handleObservedFailure(
  context: ObservedFailureContext,
): Promise<never> {
  const originalError = context.error;

  try {
    await preserveFailureContext(context);
    await captureHighlightedFailureScreenshot(context);
    await clearObservationDecorations(context.page);
    await captureCleanFailureEvidence(context);
    await captureBoundedRedactedDiagnostics(context);
    await persistFlawArtifacts(context);
    updateFailedChainAndBlockDependants(context.state, context.chainLinkId);
    await pauseForHumanInspection(context.runtime, context.flawId);
  } catch (evidenceError: unknown) {
    recordEvidenceGenerationFailure(
      context.state,
      context.flawId,
      evidenceError,
    );
  }

  throw originalError;
}

This preserves the required order while ensuring the original application failure always remains authoritative.

Required final outcome examples
Valid product success
{
  "runValidity": {
    "isValid": true
  },
  "evidence": {
    "status": "COMPLETE"
  },
  "mandatoryChainLinksPassed": 12,
  "ownerObservedRun": true,
  "finalDecision": "EXAM READY"
}
Valid observed product failure
{
  "runValidity": {
    "isValid": true
  },
  "evidence": {
    "status": "COMPLETE"
  },
  "firstBrokenLink": "SUBJECT_TO_TOPIC",
  "blockingFlawIds": [
    "FLAW-TOPIC-001"
  ],
  "finalDecision": "EXAM NOT READY"
}
Invalid certification run
{
  "runValidity": {
    "isValid": false,
    "invalidReasons": [
      "Final OBSERVED confirmation was not supplied."
    ]
  },
  "finalDecision": "TEST RUN INVALID"
}


Mandatory Runtime Environment

From:

Set-Location "E:\GK Circle v2\app"

set:

$env:E2E_HUMAN_OBSERVATION = "true"
$env:E2E_PAUSE_AT_CHECKPOINTS = "true"
$env:E2E_PAUSE_ON_FAILURE = "true"
$env:E2E_SLOW_MO_MS = "300"
$env:E2E_HUMAN_PAUSE_TIMEOUT_MS = "0"

Also load the real required environment variables without printing secrets:

E2E_BASE_URL
E2E_ADMIN_EMAIL
E2E_ADMIN_PASSWORD
E2E_LEARNER_EMAIL
E2E_LEARNER_PASSWORD
E2E_MAILPIT_URL

Fail as TEST RUN INVALID before data creation when mandatory credentials or runtime configuration are missing.

Mandatory Execution Command

Run:

On Windows PowerShell, use `npx.cmd` rather than `npx`. PowerShell may resolve
`npx` to `npx.ps1`, which is blocked when script execution is disabled and
prevents Playwright from starting Chromium.

npx playwright test `
  tests/e2e/exam-full-browser-operator-learner-readiness.spec.ts `
  --project=chromium-observation `
  --headed `
  --workers=1 `
  --retries=0 `
  --trace=retain-on-failure

Windows PowerShell equivalent:

npx.cmd playwright test `
  tests/e2e/exam-full-browser-operator-learner-readiness.spec.ts `
  --project=chromium-observation `
  --headed `
  --workers=1 `
  --retries=0 `
  --trace=retain-on-failure

Do not merely print this command. Execute it.

Human Visibility Requirements

The Chromium window must visibly open on the operator's desktop.

The run must show:

ROLE: ADMINISTRATOR during administrative work;
current phase;
current step;
entity being created;
progress through the chain;
ROLE: LEARNER during learner activity.

Use realistic visible pauses.

Do not run in the background.

Do not switch to headless mode.

Do not minimise or automatically close the browser before evidence has been captured and the operator has inspected the final state.

Required Administrator Workflow

Through visible UI controls:

Log in as administrator.
Navigate naturally to course administration.
Create one uniquely named UPSC course.
Refresh and reopen the course.
Create all required subjects under that course.
Refresh and reopen representative subjects.
Create topics under each correct subject.
Refresh and verify every parent-child association.
Create or visibly import MCQs under the correct topics.
Reopen representative MCQs.
Create subject quizzes from browser-created questions.
Create the comprehensive mock test.
Reopen and verify quiz configuration.
Publish the quizzes/tests.
Confirm publication through the visible interface.
Log out normally.

Every entity must include the unique run identifier.

Required Learner Workflow

In a separate browser context:

Log in normally as learner.
Discover the browser-created course.
Enrol through visible UI when required.
Confirm unpublished content remains hidden.
Confirm published content is visible.
Open every required quiz/test.
Select answers through visible controls.
Exercise timer, navigation, review marking, persistence and submission.
Submit normally.
Verify result release.
Verify learner analytics.
Log out normally.
Required Instructor and Report Workflow

Through visible UI controls:

Log in as the authorised administrator or instructor.
Open instructor analytics.
Find the learner's attempts.
Reconcile attempts, marks and answer totals.
Generate CSV report.
Generate XLSX report.
Generate PDF report.
Create a scheduled report.
Verify the resulting email in Mailpit through the browser-visible workflow supported by the test.
Confirm exported and emailed data belongs to the created course, quizzes and learner.
Mandatory Chain

The following 12 links must be executed and verified:

COURSE_CREATION
COURSE_TO_SUBJECT
SUBJECT_TO_TOPIC
TOPIC_TO_MCQ
MCQ_TO_QUIZ
QUIZ_TO_PUBLICATION
PUBLISHED_TO_LEARNER
LEARNER_TO_ATTEMPT
ATTEMPT_TO_RELEASE
RESULT_TO_LEARNER_ANALYTICS
LEARNER_TO_INSTRUCTOR_ANALYTICS
ANALYTICS_TO_EXPORT_EMAIL

The final result cannot be EXAM READY unless all 12 links are PASSED.

Mandatory Flaw Behaviour

When any visible product flaw prevents the chain from continuing:

Stop at the flaw.
Preserve the page.
Do not use SQL, APIs, seeds or hidden routes to bypass it.
Highlight the affected control or missing area.
Capture after-action-highlighted.png.
Remove observation decorations.
Capture after-action-clean.png.
Capture full-page.png.
Capture focused-control.png.
Capture parent-context.png.
Capture bounded and redacted DOM evidence.
Capture console, page and network errors.
Write the flaw JSON and Markdown report.
Mark the current link FAILED.
Mark dependent links BLOCKED with blockedBy.
Pause for operator inspection.
Preserve and rethrow the original error.

A valid observed product failure must return:

EXAM NOT READY

It must not be labelled TEST RUN INVALID.

Mandatory Partial Reporting

Even when the run fails during course creation, subject creation, topic creation, MCQ creation, quiz creation or any later stage, generate the partial evidence package in a finally block.

Required output:

artifacts/exam-readiness/<RUN_ID>/

including:

acceptance-results.json
data-chain-report.md
full-flaw-report.md
executive-release-summary.md
report.html
checkpoint.json
screenshots
flaw reports
console logs
page errors
network failures
traces
videos where available

The report must state:

last successful link;
first broken link;
flaw identifier;
expected result;
actual result;
exact reproduction steps;
screenshot paths;
downstream blocked links;
recommended investigation targets;
final verdict.
Final Operator Confirmation

After the workflow and mandatory evidence generation, prompt the operator:

Type OBSERVED to certify that you personally watched this visible browser run:

Only the exact accepted confirmation may set:

{
  "ownerObservedRun": true,
  "confirmationValue": "OBSERVED"
}

Missing confirmation produces:

TEST RUN INVALID
Prohibited Completion Message

Do not report:

Acceptance Test Suite Complete

merely because files were created or unit tests passed.

The only acceptable completion summary must include actual observed execution evidence:

Human-visible Playwright execution: COMPLETED or FAILED AT <CHAIN_LINK>
Browser opened visibly: YES
Operator confirmation: OBSERVED
Run ID: <RUN_ID>
Final decision: EXAM READY | EXAM NOT READY | TEST RUN INVALID
First broken link: <VALUE>
Flaw IDs: <VALUES>
Artifact directory: <PATH>
Playwright process exit code: <CODE>
