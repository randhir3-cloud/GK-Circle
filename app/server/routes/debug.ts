import { defineEventHandler } from "h3";
import dns from "dns";

export default defineEventHandler(async () => {
  console.log("[DEBUG ROUTE] Received /debug request");

  const hosts = [
    "api.railway.internal",
    "kratos.railway.internal",
    "postgres.railway.internal",
    "redis.railway.internal",
  ];

  const dnsResults: Record<
    string,
    {
      addresses?: dns.LookupAddress[];
      elapsed_ms: number;
      error?: string;
    }
  > = {};

  for (const host of hosts) {
    const start = Date.now();
    console.log(`[DEBUG ROUTE] Resolving DNS for: ${host}`);
    try {
      const addresses = await dns.promises.lookup(host, { all: true });
      dnsResults[host] = {
        addresses,
        elapsed_ms: Date.now() - start,
      };
      const elapsed = Date.now() - start;
      console.log(`[DEBUG ROUTE] Resolved ${host} in ${elapsed}ms`);
    } catch (err) {
      const e = err as Error;
      const elapsed = Date.now() - start;
      dnsResults[host] = {
        error: e.message,
        elapsed_ms: elapsed,
      };
      console.log(
        `[DEBUG ROUTE] DNS resolution failed for ${host} in ${elapsed}ms: ${e.message}`
      );
    }
  }

  const targets = [
    "http://api.railway.internal:3000/api/healthz",
    "http://api.railway.internal:3000/api/v1/healthz",
  ];

  const fetchResults: Record<
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
    console.log(`[DEBUG ROUTE] Fetching target: ${target}`);
    try {
      const controller = new AbortController();
      const id = setTimeout(() => {
        controller.abort();
      }, 2000);

      const res = await fetch(target, { signal: controller.signal });
      clearTimeout(id);

      const elapsed = Date.now() - start;
      fetchResults[target] = {
        status: res.status,
        statusText: res.statusText,
        elapsed_ms: elapsed,
      };
      console.log(
        `[DEBUG ROUTE] Fetched ${target} in ${elapsed}ms status ${res.status}`
      );
    } catch (err) {
      const e = err as Error;
      const elapsed = Date.now() - start;
      fetchResults[target] = {
        error: e.message,
        name: e.name,
        elapsed_ms: elapsed,
      };
      console.log(
        `[DEBUG ROUTE] Fetch failed for ${target} in ${elapsed}ms error ${e.message}`
      );
    }
  }

  return {
    dns: dnsResults,
    fetches: fetchResults,
  };
});
