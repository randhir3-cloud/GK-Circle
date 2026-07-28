import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";
import AttemptResultHeader from "@/components/attempt/AttemptResultHeader.vue";
import AttemptResultPending from "@/components/attempt/AttemptResultPending.vue";
import AttemptScoreCard from "@/components/attempt/AttemptScoreCard.vue";
import AttemptSummary from "@/components/attempt/AttemptSummary.vue";
import AttemptQuestionReview from "@/components/attempt/AttemptQuestionReview.vue";

describe("AttemptResultPending", () => {
  it("renders pending release message and instructions link", () => {
    const wrapper = mount(AttemptResultPending, {
      props: {
        attemptId: "12345678-abcd-efgh-1234-567890abcdef",
        submittedAt: "2026-07-28T10:00:00Z",
        message: "Results for this assessment have not been released yet.",
        instructionsPath: "/attempt/quizzes/q123",
      },
    });

    expect(wrapper.text()).toContain("Results Pending Release");
    expect(wrapper.text()).toContain(
      "Results for this assessment have not been released yet."
    );
    expect(wrapper.text()).toContain("Back to Instructions");
  });
});

describe("AttemptResultHeader", () => {
  it("renders status badge and attempt id correctly", () => {
    const wrapper = mount(AttemptResultHeader, {
      props: {
        attemptId: "12345678-abcd-efgh-1234-567890abcdef",
        status: "SUBMITTED",
        submittedAt: "2026-07-28T10:00:00Z",
        instructionsPath: "/attempt/quizzes/q123",
      },
    });

    expect(wrapper.text()).toContain("Assessment Results");
    expect(wrapper.text()).toContain("Submitted");
    expect(wrapper.text()).toContain("12345678");
  });

  it("renders Auto-Submitted badge when status is AUTO_SUBMITTED", () => {
    const wrapper = mount(AttemptResultHeader, {
      props: {
        attemptId: "12345678-abcd-efgh-1234-567890abcdef",
        status: "AUTO_SUBMITTED",
        submittedAt: "2026-07-28T10:00:00Z",
        instructionsPath: "/attempt/quizzes/q123",
      },
    });

    expect(wrapper.text()).toContain("Auto-Submitted");
  });
});

describe("AttemptScoreCard", () => {
  it("renders total score, max score, percentage and time duration", () => {
    const wrapper = mount(AttemptScoreCard, {
      props: {
        summary: {
          total_score: 8,
          max_score: 10,
          percentage: 80,
          duration_seconds: 150,
          passed: true,
        },
      },
    });

    expect(wrapper.text()).toContain("8");
    expect(wrapper.text()).toContain("/ 10");
    expect(wrapper.text()).toContain("80%");
    expect(wrapper.text()).toContain("2 min");
    expect(wrapper.text()).toContain("PASSED");
  });

  it("hides score and percentage when canShowScore is false", () => {
    const wrapper = mount(AttemptScoreCard, {
      props: {
        summary: {
          total_score: 8,
          max_score: 10,
          percentage: 80,
          duration_seconds: 150,
          passed: true,
        },
        canShowScore: false,
      },
    });

    expect(wrapper.text()).not.toContain("Total Score");
    expect(wrapper.text()).not.toContain("Percentage");
    expect(wrapper.text()).toContain("Time Taken");
  });
});

describe("AttemptSummary", () => {
  it("renders metrics grid correctly", () => {
    const wrapper = mount(AttemptSummary, {
      props: {
        summary: {
          answered: 8,
          correct: 6,
          incorrect: 2,
          unanswered: 2,
          unscored: 0,
        },
      },
    });

    expect(wrapper.text()).toContain("Performance Summary");
    expect(wrapper.text()).toContain("Answered");
    expect(wrapper.text()).toContain("Correct");
    expect(wrapper.text()).toContain("Incorrect");
    expect(wrapper.text()).toContain("Unanswered");
  });
});

describe("AttemptQuestionReview", () => {
  const mockQuestions = [
    {
      id: "q-1",
      position: 0,
      question: "Capital of France?",
      type: 1,
      options: [
        { id: 1, text: "Paris", selected: true, correct: true },
        { id: 2, text: "London", selected: false, correct: false },
      ],
      points: 5,
      is_marked_review: false,
      is_correct: true,
      score: 5,
      explanation: "Paris is the capital city of France.",
    },
  ];

  it("renders question text, options and explanation", () => {
    const wrapper = mount(AttemptQuestionReview, {
      props: {
        questions: mockQuestions,
        canShowCorrectness: true,
        canShowExplanations: true,
      },
    });

    expect(wrapper.text()).toContain("Question Review");
    expect(wrapper.text()).toContain("Capital of France?");
    expect(wrapper.text()).toContain("Paris");
    expect(wrapper.text()).toContain("Paris is the capital city of France.");
  });

  it("hides correctness indicators when canShowCorrectness is false", () => {
    const wrapper = mount(AttemptQuestionReview, {
      props: {
        questions: mockQuestions,
        canShowCorrectness: false,
        canShowExplanations: true,
      },
    });

    expect(wrapper.text()).toContain("Capital of France?");
    expect(wrapper.find(".filter-btn--correct").exists()).toBe(false);
    expect(wrapper.find(".badge--correct").exists()).toBe(false);
  });
});
