import { describe, expect, it } from "vitest";
import { isProductionRuntime } from "../../server/utils/production-config";

describe("production runtime detection", () => {
  it("allows an explicitly local runtime for a production-optimized image", () => {
    expect(
      isProductionRuntime({ NODE_ENV: "production", APP_ENV: "local" })
    ).toBe(false);
  });

  it("does not allow APP_ENV to weaken an explicit production deployment", () => {
    expect(
      isProductionRuntime({
        NODE_ENV: "production",
        APP_ENV: "production",
      })
    ).toBe(true);
  });

  it("keeps production validation fail-closed when APP_ENV is absent", () => {
    expect(isProductionRuntime({ NODE_ENV: "production" })).toBe(true);
  });

  it("recognizes the Railway production environment", () => {
    expect(
      isProductionRuntime({
        NODE_ENV: "development",
        RAILWAY_ENVIRONMENT: "production",
      })
    ).toBe(true);
  });

  it("recognizes the public production hostname", () => {
    expect(
      isProductionRuntime({
        NODE_ENV: "development",
        NUXT_PUBLIC_BASE_URL: "https://gkcircle.com",
      })
    ).toBe(true);
  });
});
