import { describe, expect, it } from "vitest";
import {
  buildVerificationBody,
  configuredRedirectOrigins,
  findVerificationSubmitNode,
  replacementVerificationState,
  resolveAllowedRedirect,
  retryAfterSeconds,
  safeNodeUrl,
  verificationSubmissionBlocked,
  valuesFromVerificationFlow,
} from "~/utils/verificationFlow";

const flow = {
  ui: {
    nodes: [
      {
        type: "input",
        group: "default",
        attributes: { name: "first", value: "one", type: "text" },
      },
      {
        type: "input",
        group: "default",
        attributes: { name: "csrf_token", value: "csrf", type: "hidden" },
      },
      {
        type: "input",
        group: "code",
        attributes: { name: "method", value: "code", type: "submit" },
      },
    ],
  },
};

describe("verification flow helpers", () => {
  it("derives values without changing node order", () => {
    const before = flow.ui.nodes.slice();
    expect(valuesFromVerificationFlow(flow)).toEqual({
      first: "one",
      csrf_token: "csrf",
      method: "code",
    });
    expect(flow.ui.nodes).toEqual(before);
  });

  it("builds browser-form data with hidden and clicked submit nodes", () => {
    const body = buildVerificationBody(
      flow,
      { first: "updated" },
      flow.ui.nodes[2]
    );
    expect([...body.entries()]).toEqual([
      ["first", "updated"],
      ["csrf_token", "csrf"],
      ["method", "code"],
    ]);
  });

  it("allows only configured redirect origins", () => {
    const origins = configuredRedirectOrigins({
      currentOrigin: "https://gkcircle.com",
      baseUrl: "https://gkcircle.com",
      configuredOrigins: "https://www.gkcircle.com",
    });
    expect(
      resolveAllowedRedirect(
        "/instructor",
        "https://gkcircle.com/verification",
        origins
      )
    ).toBe("https://gkcircle.com/instructor");
    expect(
      resolveAllowedRedirect(
        "https://example.invalid/steal",
        "https://gkcircle.com",
        origins
      )
    ).toBeNull();
  });

  it("rejects executable node URLs", () => {
    expect(safeNodeUrl("javascript:alert(1)", "https://gkcircle.com")).toBe("");
    expect(safeNodeUrl("/safe", "https://gkcircle.com")).toBe(
      "https://gkcircle.com/safe"
    );
  });

  it("parses numeric and date Retry-After values", () => {
    expect(retryAfterSeconds("12", 0)).toBe(12);
    expect(retryAfterSeconds("Thu, 01 Jan 1970 00:00:10 GMT", 0)).toBe(10);
  });

  it("replaces the complete flow and does not retain stale values", () => {
    const replacementFlow = {
      id: "replacement",
      ui: {
        nodes: [
          {
            type: "input",
            attributes: { name: "code", type: "text", value: "" },
          },
        ],
      },
    };
    const replacement = replacementVerificationState(replacementFlow);

    expect(replacement.flow).toBe(replacementFlow);
    expect(replacement.values).toEqual({ code: "" });
    expect(replacement.values).not.toHaveProperty("csrf_token");
  });

  it("selects the active resend node without reordering the source", () => {
    const before = flow.ui.nodes.slice();
    expect(findVerificationSubmitNode(flow, "method", "code")).toBe(
      flow.ui.nodes[2]
    );
    expect(flow.ui.nodes).toEqual(before);
  });

  it("blocks duplicate submissions while allowing the active resend", () => {
    expect(
      verificationSubmissionBlocked({
        isLoading: true,
        resending: false,
        resendSubmission: false,
      })
    ).toBe(true);
    expect(
      verificationSubmissionBlocked({
        isLoading: false,
        resending: true,
        resendSubmission: false,
      })
    ).toBe(true);
    expect(
      verificationSubmissionBlocked({
        isLoading: false,
        resending: true,
        resendSubmission: true,
      })
    ).toBe(false);
  });
});
