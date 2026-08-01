export default defineNuxtPlugin((nuxtApp) => {
  // On the server side (SSR), route API and Kratos requests via internal private domains
  // directly in $fetch instead of mutating global runtime config which leaks to the client.
  if (import.meta.server) {
    const originalFetch = globalThis.$fetch;
    type FetchType = (
      request: string | Request | URL,
      options?: Record<string, unknown>
    ) => Promise<unknown>;

    globalThis.$fetch = ((
      request: string | Request | URL,
      options?: Record<string, unknown>
    ) => {
      let req = request;
      if (typeof req === "string") {
        const publicApiUrl = nuxtApp.$config.public.apiUrl as
          | string
          | undefined;
        const publicKratosUrl = nuxtApp.$config.public.kratosUrl as
          | string
          | undefined;

        if (publicApiUrl && req.startsWith(publicApiUrl)) {
          req = req.replace(
            publicApiUrl,
            "http://api.railway.internal:3000/api/v1"
          );
        } else if (publicKratosUrl && req.startsWith(publicKratosUrl)) {
          req = req.replace(
            publicKratosUrl,
            "http://kratos.railway.internal:4433"
          );
        }
      }
      return originalFetch(req, options);
    }) as unknown as FetchType;
  }

  const config = useRuntimeConfig();

  // Sanitize baseUrl
  const base = config.public.baseUrl;
  if (base && !base.startsWith("http://") && !base.startsWith("https://")) {
    config.public.baseUrl = "https://" + base;
  }

  // Sanitize apiUrl
  const api = config.public.apiUrl;
  if (api && !api.startsWith("http://") && !api.startsWith("https://")) {
    config.public.apiUrl = "https://" + api;
  }

  // Sanitize apiSocketUrl
  const ws = config.public.apiSocketUrl;
  if (ws && !ws.startsWith("ws://") && !ws.startsWith("wss://")) {
    config.public.apiSocketUrl = "wss://" + ws;
  }

  // Sanitize kratosUrl
  const kratos = config.public.kratosUrl;
  if (
    kratos &&
    !kratos.startsWith("http://") &&
    !kratos.startsWith("https://")
  ) {
    config.public.kratosUrl = "https://" + kratos;
  }
});
