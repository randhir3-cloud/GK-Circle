export default defineNuxtPlugin(() => {
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
