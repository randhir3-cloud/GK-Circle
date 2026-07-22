# Student Learning Journey API Contract (v1)

This document locks down the API contract for the student learning and assessment experience.

## Base Prefix
All endpoints are prefixed with `/api/v1`.

---

## 1. Recommendations Endpoint
Retrieve personalized learning recommendations, daily goals, revision queue size, and weak topics for the logged-in student.

- **Route:** `GET /tests/student/dashboard/recommendations`
- **Authentication:** Bearer JWT (student role required)
- **Rate Limit:** 60 requests/minute
- **Latency SLA:** ≤ 300 ms

### Response Schema (200 OK)
```json
{
  "weakTopics": [
    {
      "topicId": "string",
      "topicName": "string",
      "accuracy": 0.55,
      "totalQuestions": 40
    }
  ],
  "nextRecommendedTest": {
    "testId": "string",
    "title": "string",
    "difficulty": "easy",
    "reason": "Constitution basic introduction practice suggested"
  },
  "revisionQueueSize": 12,
  "dailyGoal": {
    "target": 20,
    "completed": 5,
    "description": "Attempt 20 questions today",
    "type": "QUESTIONS"
  },
  "recommendedDifficulty": "medium"
}
```

---

## 2. Topic Stats Endpoint
Retrieve curriculum completion progress and learning metrics for a specific topic.

- **Route:** `GET /tests/topics/:topicId/stats`
- **Authentication:** Bearer JWT (student role required)
- **Latency SLA:** ≤ 300 ms

### Response Schema (200 OK)
```json
{
  "totalAttempted": 15,
  "accuracy": 72.5,
  "incorrectPracticeCount": 4,
  "recommendedDifficulty": "medium"
}
```

---

## 3. Modified Attempt Retrieve Endpoint
Modified to return correctness reveal details for already checked answers on non-`EXAM` attempts to enable resuming/continuing sessions.

- **Route:** `GET /tests/:testId/attempts/:attemptId`
- **Authentication:** Bearer JWT (owner only)
- **Latency SLA:** ≤ 300 ms

### Response Schema Changes
```json
{
  "id": "string",
  "questions": [
    {
      "testQuestionId": "string",
      "questionId": "string",
      "sortOrder": 1,
      "marks": 2.0,
      "negativeMarks": 0.66,
      "questionText": "string",
      "questionType": "SINGLE_CORRECT",
      "options": [
        { "id": "opt-1", "optionText": "Option text" }
      ],
      "selectedOptionId": "opt-1",
      "selectedOptionIds": ["opt-1"],
      "reveal": {
        "isCorrect": true,
        "explanation": "Detailed explanation of correct answer.",
        "correctOptionIds": ["opt-1"]
      }
    }
  ]
}
```

---

## 4. Modified Published Test Series details
Modified to accept authenticated student context and return overall progress.

- **Routes:** 
  - `GET /test-series/slug/:slug/published`
  - `GET /test-series/:id/subjects/:subjectLinkId/published`
- **Authentication:** Bearer JWT (student role required)
- **Latency SLA:** ≤ 300 ms

### Response Extensions
For `GET /test-series/slug/:slug/published`:
```json
{
  "id": "string",
  "title": "string",
  "subjects": [
    {
      "id": "string",
      "subject": { "id": "string", "name": "Polity" },
      "completedTestsCount": 4,
      "totalTestsCount": 10
    }
  ]
}
```

For `GET /test-series/:id/subjects/:subjectLinkId/published`:
```json
{
  "id": "string",
  "topics": [
    {
      "id": "string",
      "topic": { "id": "string", "name": "Executive" },
      "completedTestsCount": 2,
      "totalTestsCount": 5
    }
  ]
}
```

---

## Error Codes
| HTTP Status | Reason Code | Description |
|-------------|-------------|-------------|
| 401 | `UNAUTHORIZED` | Token is missing or invalid. |
| 403 | `FORBIDDEN` | Insufficient role or attempt ownership violation. |
| 404 | `NOT_FOUND` | Subject link, topic, or test does not exist. |
| 429 | `TOO_MANY_REQUESTS` | Rate limit exceeded. |
