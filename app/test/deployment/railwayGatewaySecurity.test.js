import fs from "node:fs";
import path from "node:path";
import { describe, expect, it } from "vitest";

describe("Railway production gateway", () => {
  const gateway = fs.readFileSync(
    path.resolve(process.cwd(), "../deploy/nginx.railway.conf"),
    "utf8"
  );

  it("does not expose Mailpit", () => {
    expect(gateway).not.toMatch(/\/mailpit\//i);
    expect(gateway).not.toMatch(/mailpit\.railway\.internal/i);
  });

  it("continues to route the Kratos public API", () => {
    expect(gateway).toMatch(/location\s+~\s+\^\/kratos\//);
    expect(gateway).toMatch(/kratos\.railway\.internal:4433/);
  });
});
