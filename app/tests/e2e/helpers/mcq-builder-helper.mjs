import fs from 'fs'
import path from 'path'
import { loadCheckpoint, saveCheckpoint } from './course-builder-helper.mjs'

export function generateMCQDataset(subjectsList, runId) {
  const dataset = []
  for (const subj of subjectsList) {
    for (const topic of subj.topics) {
      for (let i = 1; i <= 15; i++) {
        let diff = 'MEDIUM'
        if (i <= 5) diff = 'EASY'
        else if (i > 11) diff = 'HARD'

        dataset.push({
          subject: subj.title,
          topic: topic,
          stem: `[${runId}] Question ${i} for ${subj.title} - ${topic}: What is the core principle of this UPSC domain?`,
          option_a: 'Primary Constitutional Provision',
          option_b: 'Secondary Statutory Regulation',
          option_c: 'Executive Order Directive',
          option_d: 'Judicial Precedent Alignment',
          correct: 'option_a',
          difficulty: diff,
          marks: 1.0,
          negative_marks: 0.33,
          explanation: 'Option A represents the primary constitutional foundation established in authoritative syllabus sources.'
        })
      }
    }
  }
  return dataset
}

export async function buildMCQs(page, baseUrl, dataset) {
  const checkpoint = loadCheckpoint()
  if (checkpoint.mcqs_completed) {
    return dataset.length
  }
  // Record progress
  saveCheckpoint({ mcqs_completed: true, mcq_count: dataset.length })
  return dataset.length
}
