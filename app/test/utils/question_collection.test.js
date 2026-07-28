import { describe, expect, it } from "vitest";
import {
  describeResolutionStatus,
  filterFromApi,
  filterToApiPayload,
  memberQuestionIds,
  questionTitleById,
  RESOLUTION_STATUS_METADATA_PENDING,
  RESOLUTION_STATUS_RESOLVED,
} from "@/utils/question_collection";

describe("question_collection utils", () => {
  it("converts API filters to form values and back", () => {
    const form = filterFromApi({
      subject: "History",
      year: 2024,
      pyq_status: true,
    });
    expect(form.subject).toBe("History");
    expect(form.year).toBe("2024");
    expect(form.pyq_status).toBe("true");

    expect(filterToApiPayload(form)).toEqual({
      subject: "History",
      year: 2024,
      pyq_status: true,
    });
  });

  it("rejects invalid year values", () => {
    expect(() => filterToApiPayload({ year: "abc" })).toThrow(
      "Year must be a whole number"
    );
  });

  it("orders STATIC member IDs by position", () => {
    expect(
      memberQuestionIds({
        members: [
          { question_id: "b", position: 1 },
          { question_id: "a", position: 0 },
        ],
      })
    ).toEqual(["a", "b"]);
  });

  it("describes resolve statuses for STATIC vs DYNAMIC", () => {
    expect(describeResolutionStatus(RESOLUTION_STATUS_RESOLVED)).toContain(
      "STATIC"
    );
    expect(
      describeResolutionStatus(RESOLUTION_STATUS_METADATA_PENDING)
    ).toContain("DYNAMIC");
  });

  it("resolves question titles from the quiz bank list", () => {
    expect(
      questionTitleById(
        [{ question_id: "q-1", question: "Capital of France?" }],
        "q-1"
      )
    ).toBe("Capital of France?");
  });
});
