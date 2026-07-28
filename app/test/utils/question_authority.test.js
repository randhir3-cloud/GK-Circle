import { describe, expect, it } from "vitest";
import {
  ANSWER_REVIEW_STATUSES,
  buildAuthorityPayload,
  parseAnswerKeys,
} from "@/utils/question_authority";

describe("question_authority utils", () => {
  it("parses JSON and array answer keys", () => {
    expect(parseAnswerKeys("[2,3]")).toEqual([2, 3]);
    expect(parseAnswerKeys([4])).toEqual([4]);
    expect(parseAnswerKeys(null, 1)).toEqual([1]);
    expect(parseAnswerKeys("invalid", 3)).toEqual([3]);
  });

  it("builds authority payload with independent official and authoritative keys", () => {
    const payload = buildAuthorityPayload({
      answers: [1],
      officialAnswerKeys: [2],
      authoritativeAnswerKeys: [3],
      answerReviewStatus: "REVISED",
      answerRevisionReason: "Commission notice",
      answerRevisionSource: "UPPSC",
    });

    expect(payload).toEqual({
      answers: [1],
      official_answer: [2],
      authoritative_answer: [3],
      answer_review_status: "REVISED",
      answer_revision_reason: "Commission notice",
      answer_revision_source: "UPPSC",
    });
  });

  it("falls back to operational answers when authority keys are empty", () => {
    const payload = buildAuthorityPayload({
      answers: [4],
      officialAnswerKeys: [],
      authoritativeAnswerKeys: [],
      answerReviewStatus: "",
    });

    expect(payload.official_answer).toEqual([4]);
    expect(payload.authoritative_answer).toEqual([4]);
    expect(payload.answer_review_status).toBe("UNREVIEWED");
  });

  it("exports review status options", () => {
    expect(ANSWER_REVIEW_STATUSES.map((item) => item.value)).toEqual([
      "UNREVIEWED",
      "CONFIRMED",
      "DISPUTED",
      "REVISED",
    ]);
  });
});
