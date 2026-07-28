export const COLLECTION_KIND_STATIC = "STATIC";
export const COLLECTION_KIND_DYNAMIC = "DYNAMIC";

export const RESOLUTION_STATUS_RESOLVED = "RESOLVED";
export const RESOLUTION_STATUS_METADATA_PENDING = "METADATA_PENDING";
export const RESOLUTION_STATUS_EMPTY_FILTER_ALL = "ALL_QUIZ_QUESTIONS";

export const emptyDynamicFilter = () => ({
  subject: "",
  topic: "",
  year: "",
  difficulty: "",
  pyq_status: "",
});

export const filterFromApi = (filter) => {
  if (!filter) {
    return emptyDynamicFilter();
  }

  return {
    subject: filter.subject ?? "",
    topic: filter.topic ?? "",
    year:
      filter.year === null || filter.year === undefined
        ? ""
        : String(filter.year),
    difficulty: filter.difficulty ?? "",
    pyq_status:
      filter.pyq_status === null || filter.pyq_status === undefined
        ? ""
        : filter.pyq_status
        ? "true"
        : "false",
  };
};

export const filterToApiPayload = (formFilter) => {
  const payload = {};
  const subject = String(formFilter?.subject || "").trim();
  const topic = String(formFilter?.topic || "").trim();
  const difficulty = String(formFilter?.difficulty || "").trim();
  const yearRaw = String(formFilter?.year || "").trim();
  const pyqRaw = String(formFilter?.pyq_status || "").trim();

  if (subject) payload.subject = subject;
  if (topic) payload.topic = topic;
  if (difficulty) payload.difficulty = difficulty;
  if (yearRaw) {
    const year = Number(yearRaw);
    if (!Number.isInteger(year)) {
      throw new Error("Year must be a whole number");
    }
    payload.year = year;
  }
  if (pyqRaw === "true") payload.pyq_status = true;
  if (pyqRaw === "false") payload.pyq_status = false;

  return payload;
};

export const describeResolutionStatus = (status) => {
  switch (status) {
    case RESOLUTION_STATUS_RESOLVED:
      return "Ordered STATIC membership";
    case RESOLUTION_STATUS_METADATA_PENDING:
      return "DYNAMIC filters stored — taxonomy resolution pending";
    case RESOLUTION_STATUS_EMPTY_FILTER_ALL:
      return "DYNAMIC empty filter — all quiz-linked questions";
    default:
      return status || "Unknown resolution status";
  }
};

export const memberQuestionIds = (collection) =>
  (collection?.members || [])
    .slice()
    .sort((a, b) => a.position - b.position)
    .map((member) => member.question_id);

export const questionTitleById = (questions, questionId) => {
  const match = (questions || []).find(
    (question) =>
      question.question_id === questionId || question.id === questionId
  );
  return match?.question || `Question ${questionId}`;
};
