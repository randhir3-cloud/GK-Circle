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
