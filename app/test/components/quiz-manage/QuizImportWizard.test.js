import { mount } from "@vue/test-utils";
import { beforeEach, describe, expect, it, vi } from "vitest";
import QuizImportWizard from "@/components/quiz-manage/QuizImportWizard.vue";

const createPreviewJob = vi.fn();
const commitPreviewJob = vi.fn();

vi.mock("@/composables/quiz_import_jobs", () => ({
  useQuizImportJobsApi: () => ({
    createPreviewJob,
    commitPreviewJob,
    getQuizImportJobAPIError: (_error, fallback) => fallback,
  }),
}));

describe("QuizImportWizard", () => {
  beforeEach(() => {
    createPreviewJob.mockReset();
    commitPreviewJob.mockReset();
  });

  it("renders upload step by default", () => {
    const wrapper = mount(QuizImportWizard, {
      props: { open: true, quizId: "quiz-1" },
      global: {
        stubs: {
          Teleport: true,
          NavigationLink: true,
        },
      },
    });

    expect(wrapper.text()).toContain("Validate CSV");
    expect(wrapper.text()).toContain("Choose CSV File");
  });

  it("shows preview errors and valid rows", async () => {
    const wrapper = mount(QuizImportWizard, {
      props: { open: true, quizId: "quiz-1" },
      global: {
        stubs: {
          Teleport: true,
          NavigationLink: true,
        },
      },
    });

    wrapper.vm.previewJob = {
      id: "job-1",
      status: "PREVIEWED",
      source_filename: "bank.csv",
      valid_row_count: 1,
      error_row_count: 1,
      total_rows: 2,
      preview: {
        valid_rows: [{ row_number: 2, question: "Capital?" }],
        errors: [{ row_number: 3, messages: ["empty question text"] }],
      },
    };
    wrapper.vm.step = "preview";
    await wrapper.vm.$nextTick();

    expect(wrapper.text()).toContain("Invalid rows");
    expect(wrapper.text()).toContain("Capital?");
    expect(wrapper.text()).toContain("empty question text");
  });

  it("shows duplicate rows separately from validation errors", async () => {
    const wrapper = mount(QuizImportWizard, {
      props: { open: true, quizId: "quiz-1" },
      global: {
        stubs: {
          Teleport: true,
          NavigationLink: true,
        },
      },
    });

    wrapper.vm.previewJob = {
      id: "job-1",
      status: "PREVIEWED",
      source_filename: "bank.csv",
      valid_row_count: 1,
      error_row_count: 2,
      total_rows: 3,
      preview: {
        valid_rows: [{ row_number: 2, question: "Capital?" }],
        errors: [
          {
            row_number: 3,
            kind: "duplicate",
            messages: ["duplicate of row 2 in this CSV file"],
            duplicate_of_row: 2,
          },
          {
            row_number: 4,
            messages: ["empty question text"],
          },
        ],
      },
    };
    wrapper.vm.step = "preview";
    await wrapper.vm.$nextTick();

    expect(wrapper.text()).toContain("Duplicate rows");
    expect(wrapper.text()).toContain("Invalid rows");
    expect(wrapper.text()).toContain("duplicate of row 2");
  });

  it("disables commit when there are no valid rows", async () => {
    const wrapper = mount(QuizImportWizard, {
      props: { open: true, quizId: "quiz-1" },
      global: {
        stubs: {
          Teleport: true,
          NavigationLink: true,
        },
      },
    });

    wrapper.vm.previewJob = {
      id: "job-1",
      status: "PREVIEWED",
      valid_row_count: 0,
      error_row_count: 1,
      total_rows: 1,
      preview: {
        valid_rows: [],
        errors: [{ row_number: 2, messages: ["bad"] }],
      },
    };
    wrapper.vm.step = "preview";
    await wrapper.vm.$nextTick();

    const commitButton = wrapper
      .findAll("button")
      .find((btn) => btn.text().includes("Import"));
    expect(commitButton?.attributes("disabled")).toBeDefined();
  });

  it("shows success summary on result step", async () => {
    const wrapper = mount(QuizImportWizard, {
      props: { open: true, quizId: "quiz-1" },
      global: {
        stubs: {
          Teleport: true,
          NavigationLink: true,
        },
      },
    });

    wrapper.vm.previewJob = {
      error_row_count: 1,
    };
    wrapper.vm.resultJob = {
      commit_result: { committed_count: 2 },
    };
    wrapper.vm.step = "result";
    await wrapper.vm.$nextTick();

    expect(wrapper.text()).toContain("Import complete");
    expect(wrapper.text()).toContain("2 question(s)");
  });
});
