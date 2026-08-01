import { defineEventHandler } from "h3";

export default defineEventHandler(async () => {
  const targets = [
    "http://api.railway.internal:3000/api/healthz",
    "http://api.railway.internal:3000/api/v1/healthz",
    "http://api.railway.internal:3000/api/v1/quizzes/public",
  ];

  const results: Record<
    string,
    {
      status?: number;
      statusText?: string;
      elapsed_ms: number;
      error?: string;
      name?: string;
    }
  > = {};

  for (const target of targets) {
    const start = Date.now();
    try {
      const controller = new AbortController();
      const id = setTimeout(() => {
        controller.abort();
      }, 3000);

      const res = await fetch(target, { signal: controller.signal });
      clearTimeout(id);

      results[target] = {
        status: res.status,
        statusText: res.statusText,
        elapsed_ms: Date.now() - start,
      };
    } catch (err) {
      const e = err as Error;
      results[target] = {
        error: e.message,
        name: e.name,
        elapsed_ms: Date.now() - start,
      };
    }
  }

  return results;
});
