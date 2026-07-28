import { describe, expect, it, vi } from "vitest";
import {
  createAttemptTimer,
  formatTimerDisplay,
} from "@/composables/attempt_timer";

describe("formatTimerDisplay", () => {
  it("formats seconds as mm:ss", () => {
    expect(formatTimerDisplay(0)).toBe("00:00");
    expect(formatTimerDisplay(65)).toBe("01:05");
    expect(formatTimerDisplay(3599)).toBe("59:59");
  });

  it("formats hours as h:mm:ss when > 3600 seconds", () => {
    expect(formatTimerDisplay(3600)).toBe("1:00:00");
    expect(formatTimerDisplay(3665)).toBe("1:01:05");
    expect(formatTimerDisplay(7325)).toBe("2:02:05");
  });

  it("returns Untimed for null or invalid inputs", () => {
    expect(formatTimerDisplay(null)).toBe("--:--");
    expect(formatTimerDisplay(undefined)).toBe("--:--");
    expect(formatTimerDisplay(NaN)).toBe("--:--");
  });
});

describe("createAttemptTimer", () => {
  it("handles untimed attempts when expiresAt is missing", () => {
    const timer = createAttemptTimer({ expiresAt: null });
    expect(timer.hasDeadline).toBe(false);
    expect(timer.formattedTime.value).toBe("Untimed");
    expect(timer.isExpired.value).toBe(false);
    expect(timer.isWarning.value).toBe(false);
    expect(timer.isCritical.value).toBe(false);
  });

  it("calculates remaining seconds and formats display", () => {
    vi.useFakeTimers();
    try {
      const now = 1700000000000;
      vi.setSystemTime(now);
      const future = new Date(now + 120 * 1000).toISOString();
      const timer = createAttemptTimer({ expiresAt: future });

      expect(timer.hasDeadline).toBe(true);
      expect(timer.remainingSeconds.value).toBe(120);
      expect(timer.formattedTime.value).toBe("02:00");
      expect(timer.isWarning.value).toBe(true); // <= 300s
      expect(timer.isCritical.value).toBe(false); // > 60s
      timer.stop();
    } finally {
      vi.useRealTimers();
    }
  });

  it("triggers onExpired callback when deadline has passed", () => {
    const past = new Date(Date.now() - 5000).toISOString();
    const onExpired = vi.fn();
    const timer = createAttemptTimer({
      expiresAt: past,
      onExpired,
    });

    expect(timer.isExpired.value).toBe(true);
    expect(onExpired).toHaveBeenCalledTimes(1);
    timer.stop();
  });

  it("resyncs with status endpoint and detects terminal status", async () => {
    const future = new Date(Date.now() + 600 * 1000).toISOString();
    const onTerminalStatus = vi.fn();
    const api = {
      getAttemptStatus: vi.fn().mockResolvedValue({
        status: "SUBMITTED",
        expires_at: future,
        remaining_seconds: 600,
      }),
    };

    const timer = createAttemptTimer({
      expiresAt: future,
      quizId: "quiz-1",
      attemptId: "att-1",
      api,
      onTerminalStatus,
    });

    await timer.resync();
    expect(api.getAttemptStatus).toHaveBeenCalledWith("quiz-1", "att-1");
    expect(onTerminalStatus).toHaveBeenCalledWith("SUBMITTED");
    timer.stop();
  });

  it("cleans up resources on stop()", () => {
    const future = new Date(Date.now() + 600 * 1000).toISOString();
    const timer = createAttemptTimer({ expiresAt: future });
    expect(() => timer.stop()).not.toThrow();
  });
});
