import { describe, it, expect, vi } from "vitest";
import { useInstructorReports } from "~/composables/instructor_reports";

// Mock $fetch globally
global.$fetch = vi.fn();
global.useRuntimeConfig = () => ({ public: { apiBase: "/api/v1" } });

describe("useInstructorReports composable", () => {
  it("requests one-time export via POST /instructor/reports/exports", async () => {
    global.$fetch.mockResolvedValueOnce({
      data: { id: "job-123", status: "QUEUED" },
    });

    const { requestExport } = useInstructorReports();
    const payload = {
      export_type: "PORTFOLIO_OVERVIEW",
      export_format: "CSV",
    };
    const result = await requestExport(payload);

    expect(global.$fetch).toHaveBeenCalledWith(
      "/api/v1/instructor/reports/exports",
      expect.objectContaining({
        method: "POST",
        body: payload,
      })
    );
    expect(result.data.id).toBe("job-123");
  });

  it("generates correct download URL", () => {
    const { getDownloadUrl } = useInstructorReports();
    const url = getDownloadUrl("report-456");
    expect(url).toBe("/api/v1/instructor/reports/exports/report-456/download");
  });

  it("creates schedule via POST /instructor/reports/schedules", async () => {
    global.$fetch.mockResolvedValueOnce({
      data: { id: "sched-789", enabled: true },
    });

    const { createSchedule } = useInstructorReports();
    const payload = {
      title: "Weekly Summary",
      export_type: "PORTFOLIO_OVERVIEW",
      export_format: "PDF",
      schedule_type: "WEEKLY",
      cron_expr: "0 0 * * 1",
      timezone: "UTC",
    };
    const result = await createSchedule(payload);

    expect(global.$fetch).toHaveBeenCalledWith(
      "/api/v1/instructor/reports/schedules",
      expect.objectContaining({
        method: "POST",
        body: payload,
      })
    );
    expect(result.data.id).toBe("sched-789");
  });
});
