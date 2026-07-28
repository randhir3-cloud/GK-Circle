# COURSE-P2-T14 Renderer Contract

- `metadata.blocks` is rendered directly in array order with the original index
  as the Vue key.
- No sorting, filtering, mutation, normalization, publication logic, visibility
  logic, or adjacency derivation occurs in the frontend.
- Every array entry produces one supported block or one fallback at the same
  position. A failed block never stops a later block.
- Metadata absent, null, or structurally unusable: `Content unavailable.`
- Valid empty `blocks`: `No content available.`
- Known type with malformed required presentation data:
  `This content block is unavailable.`
- Unsupported type: `Unsupported content block.`
- Text is interpolated as escaped plain text; `v-html` is not used.
- HTTP(S) and root-relative content URLs are allowed. Unsafe, protocol-relative,
  malformed, and non-string URLs are rejected without rewriting input.
- External links use a new tab with `noopener noreferrer`; root-relative links
  remain in the current tab.
- Detail previous/next links use only IDs returned by the learner detail API.
- List and detail request-generation counters prevent stale route responses from
  overwriting current state.
