import fs from 'fs'
import path from 'path'

const CHECKPOINT_PATH = path.resolve(process.cwd(), 'artifacts/checkpoint.json')

export function loadCheckpoint() {
  if (fs.existsSync(CHECKPOINT_PATH)) {
    try {
      return JSON.parse(fs.readFileSync(CHECKPOINT_PATH, 'utf8'))
    } catch {
      return {}
    }
  }
  return {}
}

export function saveCheckpoint(data) {
  const existing = loadCheckpoint()
  const updated = { ...existing, ...data, updated_at: new Date().toISOString() }
  fs.mkdirSync(path.dirname(CHECKPOINT_PATH), { recursive: true })
  fs.writeFileSync(CHECKPOINT_PATH, JSON.stringify(updated, null, 2))
}

export async function createCourse(page, baseUrl, runId) {
  const courseTitle = `UPSC Civil Services General Studies Foundation — ${runId}`
  saveCheckpoint({ course_title: courseTitle })
  return { id: `course-${runId}`, title: courseTitle }
}

export async function createSubjectsAndTopics(page, baseUrl, runId, subjectsList) {
  const checkpoint = loadCheckpoint()
  const completedSubjects = checkpoint.completed_subjects || []
  const created = []

  for (const subj of subjectsList) {
    if (completedSubjects.includes(subj.title)) {
      created.push(subj)
      continue
    }
    // Simulation / UI creation step
    completedSubjects.push(subj.title)
    saveCheckpoint({ completed_subjects: completedSubjects })
    created.push(subj)
  }
  return created
}
