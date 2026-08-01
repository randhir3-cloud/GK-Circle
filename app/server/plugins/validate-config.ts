import { defineNitroPlugin } from "nitropack/runtime";
import { isProductionRuntime } from "../utils/production-config";

export default defineNitroPlugin(() => {
  const isProd = isProductionRuntime(process.env);

  if (!isProd) {
    console.log("Nuxt config validation bypassed in development mode.");
    return;
  }

  const urlsToValidate = [
    {
      name: "NUXT_PUBLIC_BASE_URL",
      val: process.env.NUXT_PUBLIC_BASE_URL,
      expectedScheme: "https:",
    },
    {
      name: "NUXT_PUBLIC_API_URL",
      val: process.env.NUXT_PUBLIC_API_URL,
      expectedScheme: "https:",
    },
    {
      name: "NUXT_PUBLIC_API_SOCKET_URL",
      val: process.env.NUXT_PUBLIC_API_SOCKET_URL,
      expectedScheme: "wss:",
    },
    {
      name: "NUXT_PUBLIC_KRATOS_URL",
      val: process.env.NUXT_PUBLIC_KRATOS_URL,
      expectedScheme: "https:",
    },
  ];

  for (const item of urlsToValidate) {
    if (!item.val) {
      throw new Error(
        `Nitro startup validation failed: Mandatory variable ${item.name} is missing or empty in production.`
      );
    }

    let parsed: URL;
    try {
      parsed = new URL(item.val);
    } catch (e: unknown) {
      const err = e as Error;
      throw new Error(
        `Nitro startup validation failed: ${item.name} is a malformed URL: ${err.message}`
      );
    }

    if (parsed.protocol !== item.expectedScheme) {
      throw new Error(
        `Nitro startup validation failed: ${item.name} must use '${item.expectedScheme}' protocol, got '${parsed.protocol}'`
      );
    }

    const hostname = parsed.hostname;
    if (!hostname) {
      throw new Error(
        `Nitro startup validation failed: ${item.name} has an empty hostname.`
      );
    }

    if (
      hostname === "localhost" ||
      hostname === "127.0.0.1" ||
      hostname === "::1"
    ) {
      throw new Error(
        `Nitro startup validation failed: ${item.name} hostname cannot resolve to localhost/loopback in production: ${hostname}`
      );
    }

    if (
      hostname.endsWith(".up.railway.app") ||
      hostname.endsWith(".railway.internal")
    ) {
      throw new Error(
        `Nitro startup validation failed: ${item.name} hostname cannot use a Railway auto-generated domain in production: ${hostname}`
      );
    }
  }

  console.log("Nuxt production configuration validated successfully.");
});
