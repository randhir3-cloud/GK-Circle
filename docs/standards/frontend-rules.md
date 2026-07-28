# Nuxt Frontend Rules

- Use Nuxt 3, Vue 3, Pinia, and the existing component/composable structure in `app/`.
- Do not introduce Next.js or React architecture.
- Use runtime configuration for API, WebSocket, Kratos, and public URLs.
- Keep permissions and scoring authoritative in the API.
- Preserve responsive layouts and keyboard-accessible interactions.
- Reuse existing components and stores before creating alternatives.
- User-facing terminology should use GK Circle and PCS examination language.
- Never expose secrets through `NUXT_PUBLIC_*` variables.
- Run `npm run lint`, `npm test -- --run`, and `npm run build` for relevant changes.

## Real-data completion

- Runtime feature evidence must use the existing composable or transport layer against real API routes; do not complete features with hardcoded arrays, mock JSON, or intercepted success responses.
- Do not fabricate permissions, publication state, visibility, ordering, or success state in the client.
- Render loading, empty, success, and public error states from the real request lifecycle.
- After writes, verify refresh or reopen behavior against the persisted API result.
- Preserve authoritative API ordering wherever the contract requires it.
- Unit-level mocked requests remain allowed when clearly identified as unit evidence, not runtime or persistence proof.
