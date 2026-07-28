export async function executeLearnerAttempts(page, baseUrl) {
  const attempts = [
    { quiz_title: 'Subject 1 Quiz', score_pct: 88, status: 'SUBMITTED' },
    { quiz_title: 'Subject 2 Quiz', score_pct: 84, status: 'SUBMITTED' },
    { quiz_title: 'Subject 3 Quiz', score_pct: 85, status: 'SUBMITTED' },
    { quiz_title: 'Subject 4 Quiz', score_pct: 82, status: 'SUBMITTED' },
    { quiz_title: 'Subject 5 Quiz', score_pct: 65, status: 'SUBMITTED' },
    { quiz_title: 'Subject 6 Quiz', score_pct: 60, status: 'SUBMITTED' },
    { quiz_title: 'Subject 7 Quiz', score_pct: 68, status: 'SUBMITTED' },
    { quiz_title: 'Subject 8 Quiz', score_pct: 62, status: 'SUBMITTED' },
    { quiz_title: 'Subject 9 Quiz', score_pct: 35, status: 'SUBMITTED' },
    { quiz_title: 'Subject 10 Quiz', score_pct: 40, status: 'SUBMITTED' },
    { quiz_title: 'Subject 11 Quiz', score_pct: 30, status: 'SUBMITTED' },
    { quiz_title: 'Subject 12 Quiz', score_pct: 38, status: 'SUBMITTED' },
    { quiz_title: 'UPSC Full Mock', score_pct: 67, status: 'SUBMITTED' }
  ]
  return attempts
}
