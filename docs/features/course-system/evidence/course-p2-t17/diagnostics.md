# T17 Browser Diagnostics

The T17 test attaches listeners for:

- browser console errors;
- uncaught page errors;
- failed network requests, excluding navigation-aborted requests;
- HTTP 5xx responses.

Final focused and full runs asserted all four collections were empty. The
signed-out learner detail HTTP 401 was expected and separately asserted.

The visible signed-out public message is the current Kratos middleware error
text. T17 verifies denial and does not redefine or normalize that contract.

Playwright failure policy was followed:

- the timeout failure reproduced on retry and was investigated;
- it was not accepted as a pass;
- the retry trace was preserved under `traces/`;
- fresh corrected runs passed without retry-only recovery.
