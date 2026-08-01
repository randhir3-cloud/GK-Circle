export default defineNuxtPlugin(() => {
  const config = useRuntimeConfig();

  // On the server side (SSR), route API and Kratos requests via internal private domains
  if (import.meta.server) {
    config.public.apiUrl = "http://api.railway.internal:3000/api/v1";
    config.public.kratosUrl = "http://kratos.railway.internal:4433";
  }

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
