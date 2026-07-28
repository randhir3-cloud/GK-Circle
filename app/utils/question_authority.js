export const ANSWER_REVIEW_STATUSES = [
  { value: "UNREVIEWED", label: "Unreviewed" },
  { value: "CONFIRMED", label: "Confirmed" },
  { value: "DISPUTED", label: "Disputed" },
  { value: "REVISED", label: "Revised" },
];

export const parseAnswerKeys = (value, fallback = 1) => {
  if (Array.isArray(value)) return value;
  if (!value) return [fallback];
  try {
    const parsed = JSON.parse(value);
    return Array.isArray(parsed) ? parsed : [fallback];
  } catch {
    return [fallback];
  }
};

export const buildAuthorityPayload = ({
  answers,
  officialAnswerKeys,
  authoritativeAnswerKeys,
  answerReviewStatus,
  answerRevisionReason,
  answerRevisionSource,
}) => {
  const operational = Array.isArray(answers) ? answers : [answers];
  const authoritative =
    authoritativeAnswerKeys?.length > 0 ? authoritativeAnswerKeys : operational;
  const official =
    officialAnswerKeys?.length > 0 ? officialAnswerKeys : authoritative;

  return {
    answers: operational,
    official_answer: official,
    authoritative_answer: authoritative,
    answer_review_status: answerReviewStatus || "UNREVIEWED",
    answer_revision_reason: answerRevisionReason || "",
    answer_revision_source: answerRevisionSource || "",
  };
};
