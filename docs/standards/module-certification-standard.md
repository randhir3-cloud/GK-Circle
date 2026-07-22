# Module Certification Standard

**Version:** 1.0
**Status:** Approved for local module certification use
**Owner:** Engineering
**Review Frequency:** After every major release
**Applies To:** All GK Circle modules

---

## Summary

Use this standard to certify one module at a time with the workflow:

**Fix -> Verify -> Certify -> Move to next module**

Certification is evidence-based. A plan, test definition, or passing subset is not certification.

Release actions are outside this standard unless explicitly approved by the relevant release gate.

---

## Entry Criteria

Before certification begins:

- Feature implementation is complete.
- No known merge conflicts exist.
- Local environment is operational.
- Required test data is available.
- Runtime is stable.
- Scope is frozen except for defects found during certification.

---

## Required Baseline

Record before certification starts:

```text
Frontend branch:
Backend branch:
Application SHA, if available:

Node.js:
npm/pnpm version:

PostgreSQL version:
Redis version:

Browser versions:
Operating System:

Test date/time:
```

---

## Severity Definitions

- **Critical:** Prevents use of the feature, causes data loss, crashes, security issues, or blocks certification entirely.
- **High:** Core workflow is broken and has no acceptable workaround.
- **Medium:** Feature works but has degraded behavior, confusing UX, incomplete validation, or non-blocking accessibility/performance issues.
- **Low:** Cosmetic issue, minor usability issue, copy issue, or small visual inconsistency.

Use only these result values:

- PASS
- FAIL
- BLOCKED
- NOT RUN
- NOT APPLICABLE

Do not record unexecuted checks as passing.

---

## Defect Handling

- Critical/High defects in the target module block module certification.
- Critical/High defects caused by shared components block certification for every affected surface until regression passes.
- Medium/Low defects may remain only with documented impact, reproduction steps, evidence, and explicit owner acceptance.
- Unrelated defects outside target/shared-component scope must be logged separately and not silently fixed in the certification workstream.

Every defect report must include:

- ID
- Severity
- Screen/module
- Preconditions
- Steps to reproduce
- Expected result
- Actual result
- Root cause, once known
- Affected components
- Evidence
- Status
- Verification result

Use a neutral defect table until severity and root cause are confirmed:

| ID | Severity | Screen | Root Cause | Status | Verified |
|---|---|---|---|---|---|
| M-001 | To assess | Module screen | Under investigation | Open | No |

---

## Root Cause And Fix Acceptance

Every bug fix must:

1. Reproduce the issue.
2. Identify the root cause.
3. Implement the fix.
4. Verify the original issue is resolved.
5. Verify no regression has been introduced.
6. Document the root cause and affected components in the QA report.

A defect is fixed only when:

1. The original issue can no longer be reproduced.
2. The underlying root cause has been corrected.
3. No new console errors have been introduced.
4. No new network errors have been introduced.
5. No accessibility regression has been introduced.
6. No performance regression has been introduced.
7. Shared component regressions have been checked where applicable.
8. Automated tests covering the affected behavior pass.
9. Manual QA for the affected workflow passes.
10. The regression exit checklist remains fully satisfied.

Avoid symptom-only fixes.

---

## Shared Component Impact

If a shared component is modified:

- Identify every consumer of the component.
- List affected screens.
- Regression test affected consumers only.
- Do not certify until every affected workflow passes.

Examples:

- If `Dialog` changes, regression-test every dialog.
- If `Input` changes, regression-test forms using that input.
- If `Navigation` changes, regression-test affected routes and breadcrumbs.
- Do not expand into unrelated modules unless they use the modified shared component.

Shared components include Dialog, Form, Input, Textarea, Select, Combobox, Date Picker, File Upload, Navigation, Breadcrumb, and Toast.

| Change | Regression Scope |
|---|---|
| Dialog | Every dialog using shared `Dialog` |
| Input | Every editable form using modified `Input` |
| Validation | Every affected create/edit form |
| API payload | Every affected create/update endpoint |
| Navigation | Affected routes and breadcrumbs |
| Shared hook | Every consumer of the modified hook |
| Authentication | Affected protected routes |
| Permission logic | Affected Admin, Creator, Student flows |

---

## Manual Operator Certification Policy

### Primary Rule
- Certification must be performed by a human operator using the application through the browser.
- The operator must interact with the application exactly as a normal end user would.
- Do not use Playwright, Selenium, Cypress, browser scripting, DOM manipulation, JavaScript injection, API shortcuts, database edits, or developer tools to perform the primary certification workflow.
- Automation is allowed only after manual verification has completed.

### Manual Browser Testing
The tester must perform every workflow manually. Examples:
- Open browser.
- Login normally.
- Navigate through menus.
- Click buttons manually.
- Open dialogs manually.
- Select dropdown values manually.
- Type into every textbox manually.
- Type into every textarea manually.
- Save manually.
- Cancel manually.
- Close manually.
- Reopen manually.
- No automatic typing.
- No automatic clicking.
- No keyboard macro.
- No autofill.

### Text Entry Verification
Every editable control must be tested by manual typing. For every textbox:
- Click inside textbox.
- Type one character.
- Confirm cursor remains.
- Type slowly.
- Type continuously for 30 seconds.
- Hold Backspace.
- Delete characters.
- Paste text.
- Undo.
- Redo.
- Select text.
- Replace selected text.
- Move cursor with mouse.
- Move cursor with arrow keys.
- Continue typing.
- Save.
- Cancel.
- Reopen.
- Repeat for every textbox.

### Textarea Verification
Repeat the same process for every textarea. Additionally verify:
- multi-line typing
- Enter key
- Shift+Enter
- Home
- End
- Page Up
- Page Down
- large pasted text
- scrolling
- cursor position
- selection preservation

### Dropdown Verification
For every dropdown:
- Open using mouse.
- Open using keyboard.
- Select every option manually.
- Reopen.
- Change option.
- Cancel.
- Save.
- Verify persisted value.

### Checkbox Verification
For every checkbox:
- Click once.
- Click twice.
- Keyboard toggle.
- Save.
- Reload.
- Verify persisted state.

### Radio Button Verification
- Test every option individually.

### Button Verification
Every button must be clicked manually. Examples:
- Save, Cancel, Close, Delete, Archive, Publish, Unpublish, Restore, Refresh, Retry

### Dialog Verification
Every dialog must be opened manually. Verify:
- initial focus
- typing
- closing
- Escape
- outside click
- Cancel
- Save
- reopening
- focus restoration

### Focus Verification
For every editable control:
- click
- type one character
- verify focus stays
- continue typing
- verify focus never moves
- verify cursor position
- verify selection
- verify IME
- verify paste
- verify undo
- verify redo
- Do not use automation.

### Manual Navigation
Verify manually:
- browser Back, browser Forward, Refresh, Deep links, URL editing, bookmarks

### API Verification
- Use the browser normally.
- Do not call APIs directly.
- Observe: exactly one request, payload, response, loading indicator, error handling.

### Accessibility
Use keyboard only. Verify:
- Tab, Shift+Tab, Enter, Escape, Space, Arrow keys

### Regression Rule
If a shared component changes (e.g. Dialog changed), the operator must manually test every dialog using that shared Dialog component. No shortcuts.

### Automated Testing Policy
- Automation is supplementary.
- Automation cannot certify the module.
- Automation may verify: type-check, lint, build, Playwright regression, API regression, unit tests.
- Automation cannot replace manual operator certification.

---

## Verification Requirements

Runtime:

- Frontend compiles without errors.
- Backend starts without errors.
- No uncaught page-load runtime exceptions.
- Browser console follows the console error policy.
- Network requests succeed.
- Authentication works with approved QA accounts.

Console error policy:

- Certification fails for uncaught exceptions, React errors, unexpected failed application network requests, hydration errors, and accessibility violations caused by the application.
- Expected negative-path responses are permitted only when the test intentionally triggered them, the UI handled them correctly, no raw internal error was exposed, and the result is documented.
- Known third-party warnings may be documented if they do not affect functionality.

Accessibility:

- Keyboard-only operation.
- Visible focus indicator.
- Screen-reader labels.
- Correct ARIA relationships.
- Contrast regressions introduced by fixes.
- Dialog focus trap.
- Escape key behavior.

Performance:

- No duplicate API calls for one user action.
- No infinite render loops.
- No noticeable UI lag introduced by the fix.
- No continuously increasing memory usage during repeated open/close testing.
- No layout shift.
- No focus thrashing.

---

## Exit Checklist

Before marking a module as PASS:

- No Critical defects remain.
- No High defects remain.
- Medium/Low defects are documented and explicitly accepted.
- Manual QA passes.
- Required Playwright regression passes.
- Type-check passes.
- Build passes.
- Shared component regressions are tested where applicable.
- No unintended regressions are introduced.
- Evidence exists for failures and fixes.
- No prohibited Git, deployment, staging, production, or NUC action was performed.

Final report status block:

```text
Module:

Certification Status:
PASS / FAIL / BLOCKED

Environment:
Local Development

Critical Defects:
0

High Defects:
0

Medium Defects:
N

Low Defects:
N

Shared Component Regression:
PASS / FAIL / BLOCKED / NOT RUN / NOT APPLICABLE

Automated Regression:
PASS / FAIL / BLOCKED / NOT RUN / NOT APPLICABLE

Manual QA:
PASS / FAIL / BLOCKED / NOT RUN / NOT APPLICABLE

Playwright:
PASS / FAIL / BLOCKED / NOT RUN / NOT APPLICABLE

Type Check:
PASS / FAIL / BLOCKED / NOT RUN / NOT APPLICABLE

Build:
PASS / FAIL / BLOCKED / NOT RUN / NOT APPLICABLE

Production Verification:
NOT RUN

Deployment:
NOT PERFORMED

Git Operations:
NOT PERFORMED
```
