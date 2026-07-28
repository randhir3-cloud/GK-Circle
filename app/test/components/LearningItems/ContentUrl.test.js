import { describe, expect, it } from "vitest";
import { isExternalContentUrl, isSafeContentUrl } from "@/utils/content_url";

describe("content URL safety", () => {
  it.each([
    "https://example.com/file.pdf",
    "http://example.com/video",
    "/root-relative/path",
  ])("allows %s", (value) => {
    expect(isSafeContentUrl(value)).toBe(true);
  });

  it.each([
    "javascript:alert(1)",
    "data:text/html,hello",
    "vbscript:msgbox(1)",
    "file:///tmp/file",
    "//example.com/path",
    "relative/path",
    "",
    " https://example.com",
    null,
    42,
  ])("rejects %j", (value) => {
    expect(isSafeContentUrl(value)).toBe(false);
  });

  it("distinguishes external HTTP links from root-relative links", () => {
    expect(isExternalContentUrl("https://example.com")).toBe(true);
    expect(isExternalContentUrl("HTTPS://example.com")).toBe(true);
    expect(isExternalContentUrl("/inside")).toBe(false);
  });
});
