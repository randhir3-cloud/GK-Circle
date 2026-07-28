export const QUESTION_TYPE_SINGLE = 1;
export const QUESTION_TYPE_SURVEY = 2;

export const SAVE_STATUS = {
  IDLE: "idle",
  SAVING: "saving",
  SAVED: "saved",
  FAILED: "failed",
};

export const PALETTE_STATUS = {
  NOT_VISITED: "not_visited",
  VISITED_UNANSWERED: "visited_unanswered",
  ANSWERED: "answered",
  SAVING: "saving",
  SAVE_FAILED: "save_failed",
};

export const ATTEMPT_STATUS_IN_PROGRESS = "IN_PROGRESS";
export const ATTEMPT_STATUS_SUBMITTED = "SUBMITTED";
export const ATTEMPT_STATUS_AUTO_SUBMITTED = "AUTO_SUBMITTED";

export const SUBMIT_PHASE = {
  IDLE: "idle",
  SUBMIT_REQUESTED: "submit_requested",
  QUEUE_CLOSING: "queue_closing",
  SUBMITTING: "submitting",
  SUBMITTED: "submitted",
  ERROR: "error",
  OFFLINE_EXPIRED: "offline_expired",
};

export const TIMER_WARNING_THRESHOLD_SECONDS = 300; // 5 min
export const TIMER_CRITICAL_THRESHOLD_SECONDS = 60; // 1 min
export const TIMER_RESYNC_INTERVAL_SECONDS = 90; // server clock resync

export const FORBIDDEN_PLAYER_FIELDS = [
  "answers",
  "official_answer",
  "authoritative_answer",
  "answer_review_status",
  "is_correct",
  "score",
  "total_score",
  "max_score",
];
