# GK Circle RAG Rules

Version: 1.0

Status: Mandatory

---

# Purpose

Govern all Retrieval-Augmented Generation systems.

---

# Source First Rule

Every answer must originate from verified sources.

Never fabricate sources.

---

# Retrieval Pipeline

Query

↓

Retrieval

↓

Reranking

↓

Context Building

↓

Generation

↓

Citation

---

# Approved Sources

Internal Knowledge Base

NCERT

Government Reports

PIB

PRS

Constitution

Parliament Data

Official Websites

Trusted Educational Sources

---

# Citation Rule

Every AI answer must provide sources.

---

# Hallucination Prevention Rule

If information cannot be retrieved:

State uncertainty.

Do not invent facts.

---

# UPSC Rule

UPSC answers must emphasize:

Accuracy

Source Attribution

Context

Current Affairs

Multi-Dimensional Analysis

---

# Completion Rule

RAG system is complete only when:

Retrieval

↓

Citation

↓

Answer Generation

↓

Source Verification

all work.

---

# v0.5 Addendum — Architecture-Intent Standard

The v1.0 section above defines the principles. This addendum makes them enforceable as code, configuration, and process. It uses the same additive structure as the other v0.5 standards.

This is **architecture-intent**: the RAG module is not yet implemented. The rules below are the contract it MUST satisfy when built. The next revision (v1.0 of this file) will merge these rules into the body of the document.

---

# Dependency Rule (Mandatory)

This document does not replace AGENTS.md.

It supplements AGENTS.md.

If this document conflicts with AGENTS.md, AGENTS.md wins.

If this document conflicts with another standards file, the more specific standard wins.

If ambiguity exists, document the ambiguity in an ADR before implementation.

# Verification Requirement (Mandatory)

A rule is not considered satisfied because code, configuration, tests, or documentation exist. A rule is satisfied only when evidence exists. For RAG, evidence must include:

- An evaluation set (≥ 200 hand-labeled question/answer pairs) with measured faithfulness and answer-relevancy scores
- A hallucination-rate measurement from a representative sample
- Source-coverage measurements (% of responses with ≥ 1 cited source)

---

# Source Citation (Mandatory)

Every RAG-generated response presented to an end user MUST include citations. This is non-negotiable.

## Citation Format (Mandatory)

Each response MUST include a `sources` array. Each source entry MUST contain:

| Field | Type | Required | Purpose |
|---|---|---|---|
| `sourceId` | string | yes | Stable identifier (e.g., `ncert-history-class-9-ch-3-para-12`) |
| `title` | string | yes | Human-readable title |
| `section` | string | no | Section/heading within the document |
| `page` | number | no | Page number, if applicable |
| `url` | string | no | Canonical URL if available |
| `publishedAt` | ISO 8601 date | no | Publication date |
| `authority` | enum | yes | See Source Authority Hierarchy below |
| `excerpt` | string | yes | The exact text (≤ 280 chars) the model used |
| `confidence` | number (0–1) | yes | Retrieval score from hybrid search |

## Citation Visibility (Mandatory)

- Inline citation markers (e.g., `[1]`, `[2]`) at the point of use in the answer
- A "Sources" section at the bottom of the response
- Each citation is clickable and opens the source preview/panel

## Citation Co-location Rule

When a model makes a factual claim, the citation for that claim MUST be retrievable in the same response chunk. Forbidden: generating an answer and only later appending a Sources section populated from sources the model did not actually use. The `excerpt` field enforces this: the model quotes verbatim and the system verifies the quote exists in the cited source.

## No-Citation Response Rule

If a user query cannot be answered from the available sources, the system MUST respond with one of:

- "I cannot find a reliable source for this in the available knowledge base. Try rephrasing your question."
- A targeted clarifying question (e.g., "Do you mean the 2015 amendment or the 2019 amendment?")

The system MUST NOT answer without a citation, cite a source not in the retrieved context, or generate plausible-sounding answers from parametric knowledge alone.

## Source Attribution Required (Mandatory)

Every RAG response MUST attribute every claim to a specific source. This is the foundation of trust in the RAG system. The attribution is enforced by the citation format above and by the Faithfulness Verifier (Section "Grounded Answers Only").

## Answer Confidence Tracking (Mandatory)

Every RAG response MUST carry a confidence score (0–1) per cited chunk and a roll-up response-level confidence. The confidence is computed from the re-ranking and hybrid-search scores (see Re-Ranking and Confidence Scoring below). Low-confidence responses are flagged in telemetry, get an explicit disclaimer in the UI, and are eligible for human review. Confidence is also broken down by source authority (cited NCERT chunks weigh more than USER_UPLOAD chunks for the roll-up).

## Prompt Version Tracking (Mandatory)

Every prompt template is versioned with semver and recorded in the `PromptVersion` constant (see Prompt Versioning below). The deployed version is stored in the `PromptDeployment` table. Every LLM call logs the active `promptVersion` in its Sentry span and in the per-query telemetry. A prompt change is never deployed without a version bump, an evaluation run, and an ADR if the change is user-facing.

## Embedding Version Tracking (Mandatory)

Every chunk carries `embeddedWith: EmbeddingModel.id`. Every `Embeddings` query logs the active `EmbeddingModel.id` and `EmbeddingModel.version`. The `EmbeddingModel` table records all models with their `ACTIVE`, `DEPRECATED`, or `RETIRED` status. Vector DB collections are dimensionally matched to the active model; a mismatch is a startup configuration error. The active model is gated behind an ADR (see Embedding Lifecycle below).

## Knowledge Freshness Tracking (Mandatory)

Every chunk and every document carries a `publishedAt` timestamp. The retriever supports freshness filters (e.g., "last 7 days" for current affairs). The default freshness window is determined by the use case (chat, MCQ explanation, doubt solver) and is configurable per `RetrieverConfig`. A `freshnessStaleAt` field on `ChunkMetadata` indicates when the chunk should be re-verified against the source. Stale chunks are re-fetched and re-embedded on a schedule, and the freshness window is enforced at query time (older chunks are demoted in retrieval score).

---

# Grounded Answers Only (Mandatory)

RAG-generated answers MUST be grounded in retrieved source content. The model is forbidden from inserting information from its parametric knowledge unless that information is also present in (or directly entailed by) a retrieved source.

## Three Layers of Defense

### Layer 1: Prompt-Level Grounding

The system prompt MUST explicitly instruct grounding, citation by `[n]`, and "do not invent." The full prompt is in the Prompt Versioning section below.

### Layer 2: Retrieval-Time Grounding

The retriever MUST return a sufficient number of chunks (default: top-20 vector + top-20 keyword, deduped) to maximize source coverage. A cross-encoder re-ranker filters the top-5–8 chunks passed to the model. The model MUST receive the full text of every chunk it is allowed to cite.

### Layer 3: Post-Generation Verification (Mandatory)

A second LLM call (a smaller, cheaper model, e.g., Claude Haiku 4.5) is asked "Is each factual claim in the response directly supported by the cited source?" with the cited source text in the prompt. The verifier returns:

```typescript
interface FaithfulnessVerdict {
  isFaithful: boolean;
  claimVerdicts: ClaimVerdict[];
  overallScore: number;       // 0–1
  unsupportedClaims: string[]; // Verbatim quotes
  recommendations: string[];  // "Remove", "Re-generate", "Add citation"
}
```

The verifier MUST run on every production response. It is async and non-blocking, but its result is logged to telemetry and used to compute the faithfulness score.

## Parametric Knowledge Use (Allowed vs Forbidden)

Allowed:

- Linguistic fluency, grammar, style, translation
- Reasoning over retrieved facts ("Based on the source, this implies that...")
- Explaining a concept fully entailed by the cited source

Forbidden:

- Factual claims not in the cited source
- Dates, numbers, names, or quotes not in the cited source
- "Common knowledge" contradicted by the cited source (use the source)

---

# Embedding Lifecycle (Mandatory)

Embeddings are first-class artifacts with a creation, version, deprecation, and deletion lifecycle tracked in the `EmbeddingModel` table.

## Embedding Model Fields (Mandatory)

| Field | Type | Purpose |
|---|---|---|
| `id` | string | Unique ID |
| `name` | string | e.g., `text-embedding-3-small` |
| `provider` | enum | `OPENAI`, `ANTHROPIC`, `COHERE`, `VOYAGE`, `JINA`, `SELF_HOSTED` |
| `dimensions` | number | Output vector dimensions (e.g., 1536, 3072) |
| `distanceMetric` | enum | `COSINE`, `DOT_PRODUCT`, `EUCLIDEAN` |
| `status` | enum | `ACTIVE`, `DEPRECATED`, `RETIRED` |
| `activatedAt` | DateTime | When this model became production |
| `deprecatedAt` | DateTime? | When deprecated |
| `retiredAt` | DateTime? | When removed from production |

## Lifecycle Rules

1. At any time, exactly one `EmbeddingModel` is `ACTIVE`. Switching the active model requires an ADR.
2. When switching, the new model is added with status `ACTIVE` and the old model is set to `DEPRECATED` in the same migration.
3. `DEPRECATED` models continue to serve historical queries but no new documents are embedded with them.
4. `RETIRED` models' vectors are deleted from the vector DB after a 30-day grace period.
5. The `retire` action requires admin approval and produces an audit log entry.
6. Switching embedding models is exactly the trigger for Re-Indexing (see below).

## Embedding Generation Rules

- Batch embedding is preferred: when ≥ 10 documents are queued, the worker MUST batch them into a single API call.
- Maximum batch size depends on the provider's limit (OpenAI: 2048, Anthropic: 100, Cohere: 96).
- Each embedding call is wrapped in `Sentry.startSpan` with op `embedding.generate` and tags `embedding.model`, `embedding.batch_size`.
- Failures are retried with exponential backoff (max 3 attempts); after the third failure, the document is moved to a dead-letter queue and an alert is fired.

---

# Chunking Standards (Mandatory)

The chunking strategy determines retrieval quality. Bad chunking is the #1 cause of RAG failures.

## Chunking Configuration (Mandatory)

The `Chunker` configuration is a single source of truth stored in the database (or a feature flag) so that changes are auditable.

| Field | Type | Default | Purpose |
|---|---|---|---|
| `strategy` | enum | `SEMANTIC` | `FIXED_SIZE`, `SEMANTIC`, `SLIDING_WINDOW`, `HIERARCHICAL` |
| `targetTokens` | number | 500 | Target chunk size |
| `minTokens` | number | 200 | Minimum chunk size |
| `maxTokens` | number | 1000 | Maximum chunk size |
| `overlapTokens` | number | 80 | Overlap between consecutive chunks |
| `splitter` | enum | `RECURSIVE` | `RECURSIVE`, `SENTENCE`, `PARAGRAPH`, `MARKDOWN_HEADER` |
| `preserveHeadings` | boolean | true | Keep heading context in chunk metadata |
| `keepWithNext` | list | `["table", "list", "code"]` | Element types that should not be split |

## Chunking Rules by Source Type

1. **NCERT chapters, government reports, books**: `SEMANTIC` + `MARKDOWN_HEADER`. Heading metadata preserved.
2. **Current affairs articles**: `SEMANTIC` + `SENTENCE`. Title and date added to every chunk's metadata.
3. **Acts, policies, legal documents**: `HIERARCHICAL` (section → sub-section → paragraph). Hierarchy in metadata.
4. **User-uploaded documents**: `SEMANTIC` + `PARAGRAPH`; size auto-reduced for small docs.
5. **Code (if ever chunked)**: `SLIDING_WINDOW` with `overlapTokens = 160` to preserve function context.
6. **Tables and lists**: NEVER split mid-table or mid-list. Tables are always in a single chunk.

## Chunk Metadata (Mandatory)

```typescript
interface ChunkMetadata {
  sourceId: string;
  documentTitle: string;
  section?: string;
  subSection?: string;
  page?: number;
  publishedAt?: string;
  authority: SourceAuthority;
  examType?: string[];
  subject?: string;
  topic?: string;
  difficulty?: string;
  language?: string;
  chunkIndex: number;
  totalChunks: number;
  tokenCount: number;
  contentHash: string;     // SHA-256
  embeddedAt: string;
  embeddedWith: string;    // Embedding model ID
}
```

## Chunking Anti-Patterns

- Fixed-size chunking without overlap
- Single-chunk documents (one chunk = whole document)
- Splitting tables or lists across chunks
- Chunking without preserving heading context
- Embedding before chunking (order: chunk → store raw → embed → store embedding)
- Re-embedding chunks without updating `embeddedWith`

---

# Vector DB Abstraction (Mandatory)

The vector DB is hidden behind an interface. The current production deployment uses Qdrant (per `docs/rag-architecture.md`). Future options include Weaviate, Pinecone, Milvus, or pgvector.

## VectorDBClient Interface (Mandatory)

```typescript
export interface VectorDBClient {
  upsert(collection: string, vectors: VectorRecord[]): Promise<UpsertResult>;
  search(collection: string, query: number[], options: SearchOptions): Promise<SearchResult[]>;
  delete(collection: string, ids: string[]): Promise<DeleteResult>;
  update(collection: string, id: string, updates: Partial<VectorRecord>): Promise<void>;
  scroll(collection: string, options: ScrollOptions): Promise<ScrollResult[]>;
  count(collection: string, filter?: Filter): Promise<number>;
  createCollection(spec: CollectionSpec): Promise<void>;
  deleteCollection(name: string): Promise<void>;
  healthCheck(): Promise<HealthStatus>;
}

export interface SearchOptions {
  topK: number;
  filter?: Filter;
  withPayload?: boolean;
  withVectors?: boolean;
  scoreThreshold?: number;
}
```

The default implementation is `QdrantClient`. Any code in the RAG module imports from `VectorDBClient`, never from `QdrantClient` directly.

## Collection Naming Convention

| Collection | Purpose |
|---|---|
| `chunks-ncert` | NCERT textbook chunks |
| `chunks-govt` | Government reports, white papers |
| `chunks-acts` | Acts, policies, legal documents |
| `chunks-current-affairs` | Daily current affairs |
| `chunks-books` | Reference books |
| `chunks-user-uploads` | User-uploaded documents |

Each collection uses the same distance metric (cosine) and the same dimensions (must match the `ACTIVE` embedding model). A mismatch is a configuration error that MUST be caught at startup.

## Filter Pushdown (Mandatory)

Metadata filters MUST be pushed down to the vector DB rather than applied in application code. Common pushed-down filters:

- `authority` (only NCERT and GOVT for school-level queries)
- `examType` (only `["UPSC"]` for UPSC-prep queries)
- `subject` (only `["History"]` for history queries)
- `publishedAt` (only last 7 days for current-affairs queries)
- `language` (only `"en"` if the user is on the English interface)

## Health Check (Mandatory)

A `/health` endpoint MUST be added that checks vector DB connectivity, collection existence, and a round-trip test (insert known vector, search, delete). The health check is part of the platform health check; failure pages the on-call.

---

# Re-Indexing Rules (Mandatory)

Re-indexing is the process of re-embedding and re-storing the entire knowledge base. It is a non-trivial operation requiring planning, monitoring, and rollback.

## Triggers

Re-indexing is triggered when ANY of the following changes:

1. The active embedding model changes
2. The chunking strategy changes (any field in the `Chunker` configuration)
3. The vector DB provider changes
4. The distance metric changes
5. A schema-level metadata change requires re-population
6. A manual admin request

A model change or chunking change is a "full re-index." A metadata-only change may be a "partial re-index" (only update the affected field; vectors stay the same).

## Re-Indexing State Machine

```
IDLE → PLANNING → INDEXING → VALIDATING → SWITCHING → COMPLETE
                                ↓
                              FAILED
```

### Step 1: PLANNING

- Compute total chunk count and estimated cost (vector DB storage + LLM embedding calls)
- Identify worker pool size and rate limits
- Set up a new collection (e.g., `chunks-ncert-v2`) without affecting the production collection
- Generate a re-indexing plan stored in the `ReindexJob` table

### Step 2: INDEXING

- Workers read source documents, chunk them with the new configuration, and upsert to the new collection
- A single source document is processed by at most one worker at a time (idempotency)
- Progress is logged to Sentry with breadcrumbs
- The new collection is searchable in shadow mode (queries return both; new is logged but not returned to the user)
- Parallelizable up to 16 workers

### Step 3: VALIDATING

- A subset of queries (e.g., 200 evaluation questions) is run against both old and new collections
- New collection's metrics (faithfulness, answer-relevancy, retrieval recall) must be ≥ old collection's metrics by a margin defined in the eval harness
- If validation fails, the job transitions to `FAILED` and the new collection is deleted

### Step 4: SWITCHING

- The application configuration is updated to use the new collection
- The old collection is renamed (e.g., `chunks-ncert-v1-archived`) but not yet deleted
- The application picks up the new collection on next startup (or hot-reload)
- This is a 1-second cutover; downtime is zero if the vector DB supports dual writes

### Step 5: COMPLETE

- The old collection is marked for deletion (30-day grace period)
- The `EmbeddingModel.status` is updated (old → `DEPRECATED`, new → `ACTIVE`)
- An ADR or postmortem is filed documenting the re-index

## Rollback

If a regression is detected within 7 days post-cutover:

- The old collection is restored as the production collection
- The new collection is renamed `chunks-ncert-v2-failed` and scheduled for deletion
- An incident is filed

## Re-Indexing Anti-Patterns

- Re-indexing in production without a separate collection
- Re-indexing without a validation step
- Re-indexing without a rollback plan
- Re-indexing during peak hours (00:00–06:00 UTC only)
- Re-indexing with a chunker that has not been tested on the evaluation set

---

# Hallucination Prevention (Mandatory)

Hallucination prevention is a multi-layer system.

## Layer 1: Grounded Prompting

The system prompt explicitly forbids ungrounded claims.

## Layer 2: Source-Confined Generation

The model is given the retrieved chunks in a structured format and instructed to use ONLY those chunks:

```
<sources>
[1] NCERT Class 9, History, Chapter 3, p. 47:
"...the French Revolution began in 1789..."

[2] Indian Polity, M. Laxmikanth, Chapter 7, p. 112:
"...Article 14 guarantees equality before law..."
</sources>

<user_question>
What year did the French Revolution begin?
</user_question>

<instructions>
Answer using ONLY the sources above. Cite each claim with [n]. If the sources do not contain the answer, say so.
</instructions>
```

## Layer 3: Faithfulness Verifier

A second LLM call (the verifier) checks every claim. Implementation is in the Grounded Answers section above. The verifier MUST run on every production response.

## Layer 4: User Feedback Loop

Users can flag responses as "Not helpful" or "Hallucinated." Flagged responses are logged to Sentry with the full context (question, response, sources, verifier verdict), reviewed by the AI team weekly, and added to the evaluation set as negative examples.

## Hallucination Rate Target

The system MUST maintain a hallucination rate ≤ 2% (defined as: % of responses where the faithfulness verifier returns `isFaithful: false`). This is measured weekly from a random sample of 500 production responses. If the rate exceeds 2% for two consecutive weeks, an incident is filed and a model/prompt change is blocked until the rate drops.

---

# Prompt Versioning (Mandatory)

Every prompt template is versioned. A prompt change is never deployed without a version bump and an evaluation run.

## Prompt Storage

Prompts are stored in `backend/src/rag/prompts/` as TypeScript files. Each file exports a `PROMPT_VERSION` constant and a pure build function:

```typescript
// backend/src/rag/prompts/qa-system-prompt.ts
import { PromptVersion } from '../prompt-version';

export const PROMPT_VERSION: PromptVersion = {
  id: 'qa-system-prompt',
  version: 'v3.2.1',
  createdAt: '2026-06-05',
  author: 'ai-team',
  changelog: 'Add explicit conflict resolution rule',
};

export function buildQASystemPrompt(ctx: QAPromptContext): string {
  return `<sources>
${ctx.sources.map((s, i) => `[${i + 1}] ${s.title}, ${s.section}, p. ${s.page}:\n"${s.excerpt}"`).join('\n\n')}
</sources>

<user_question>
${ctx.question}
</user_question>

<instructions>
You are an AI tutor for GK Circle. Your answers MUST be grounded in the provided sources.

Rules:
1. Answer using ONLY the information in the Sources section above.
2. Every factual claim MUST cite a source by its [n] marker.
3. If a source does not contain the answer, say "The available sources do not cover this topic."
4. Do not invent dates, names, acts, statistics, or quotes.
5. Do not use information from outside the provided Sources section, even if you know it from your training.
6. If the sources conflict, prefer the most authoritative source and note the conflict.
7. If you are uncertain, express uncertainty with phrases like "Based on the source, ..." rather than asserting.
</instructions>`;
}
```

## Prompt Versioning Rules

1. Every prompt has a `PROMPT_VERSION` constant with `id`, `version` (semver), `createdAt`, `author`, `changelog`.
2. Every prompt function is pure: given the same context, it returns the same string. No I/O, no randomness, no `Date.now()`.
3. Prompts are tracked in git. A prompt change is a code change that goes through code review and CI.
4. A prompt change MUST be evaluated against the evaluation set before production rollout. The eval results (faithfulness, answer-relevancy, retrieval recall) are attached to the PR.
5. A prompt change that does not meet the evaluation thresholds is reverted.
6. A `PromptDeployment` table records which prompt version is in production for which purpose (qa, mcq-explanation, doubt-solver, etc.).

## Prompt Caching

When using Anthropic or OpenAI models with prompt caching:

- Static instructions go in the cached prefix
- The user's question and retrieved sources go in the variable suffix
- The cache hit rate is monitored; low hit rates indicate prompts are too dynamic

## Prompt Anti-Patterns

- Prompts that reference "current date" or "now" (use a parameter; the prompt is testable in isolation)
- Prompts with hardcoded user data (PII) or secrets
- Prompts that include `ignore previous instructions` or other adversarial text
- "Clever" prompt engineering that bypasses grounding
- Prompt changes deployed without evaluation

---

# Retrieval Pipeline Contract (Mandatory)

The end-to-end retrieval pipeline is:

```
User Query
    ↓
[1] Query Understanding (optional)
    ↓
[2] Filter Derivation
    ↓
[3] Hybrid Search (vector + keyword) in VectorDB
    ↓
[4] Re-Ranking (cross-encoder)
    ↓
[5] Top-K Selection (default K=5)
    ↓
[6] Prompt Construction
    ↓
[7] LLM Call
    ↓
[8] Faithfulness Verification (async)
    ↓
[9] Response with Citations
```

### [1] Query Understanding

Optional: if the query is ambiguous, the system asks a clarifying question. Gated by a confidence score from a query-classifier.

### [2] Filter Derivation

Filters derived from:

- User profile (exam type, subjects, language)
- Query content (subject, topic, date range)
- Hard-coded safety filters (e.g., never retrieve `authority = USER_UPLOAD` for medical queries without a disclaimer)

### [3] Hybrid Search

Vector search (top-20 by cosine) ∪ Keyword search (top-20 by BM25). Deduplicate by chunk ID, preferring the higher score.

### [4] Re-Ranking

A cross-encoder model (e.g., `cross-encoder/ms-marco-MiniLM-L-12-v2` or a Cohere Rerank model) re-scores the deduplicated set. The top-20 → top-5 (or top-8) is selected.

### [5] Top-K Selection

Default K=5. Exact value is a `RetrieverConfig` field; tunable per use case (chat vs MCQ explanation vs doubt solver).

### [6] Prompt Construction

Use the appropriate versioned prompt template. See Prompt Versioning above.

### [7] LLM Call

Wrapped in `Sentry.startSpan` with op `rag.llm_call` and tags `rag.model`, `rag.prompt_version`, `rag.use_case`. Streaming responses are preferred for chat use cases.

### [8] Faithfulness Verification

Async, non-blocking. Result is logged and used for telemetry.

### [9] Response with Citations

Returned to the user with inline citations and a Sources section.

---

# Source Authority Hierarchy (Mandatory)

When sources conflict, the system follows this authority hierarchy (highest to lowest):

| Rank | Authority | Description | Examples |
|---|---|---|---|
| 1 | `CONSTITUTION` | Constitution of India and amendments | Constitution, amendments |
| 2 | `ACT` | Acts passed by Parliament or state legislatures | IPC, CrPC, RTI Act, IT Act |
| 3 | `GOVT_OFFICIAL` | Official government publications | Economic Survey, Budget Speech, PIB |
| 4 | `NCERT` | NCERT textbooks | Class 6–12 NCERT books |
| 5 | `GOVT_REPORT` | Government-commissioned reports | NITI Aayog reports, CAG reports |
| 6 | `JUDICIAL` | Supreme Court and High Court judgments | Landmark cases |
| 7 | `OFFICIAL_DATA` | Official statistical data | Census, RBI, MOSPI |
| 8 | `BOOK_ACADEMIC` | Academic reference books | Laxmikanth, Ramesh Singh, Spectrum |
| 9 | `CURRENT_AFFAIRS_OFFICIAL` | Official current-affairs sources | PIB releases, press notes |
| 10 | `CURRENT_AFFAIRS_REPUTABLE` | Reputable news sources | The Hindu, Indian Express |
| 11 | `USER_UPLOAD` | User-uploaded documents | Any PDF uploaded by a user |
| 12 | `OTHER` | Anything else | Web pages, blog posts |

## Conflict Resolution

When two sources conflict, the system:

1. Selects the higher-authority source as the primary answer
2. Notes the conflict in the response ("Source A says X; Source B says Y. The higher-authority source is A.")
3. Does NOT silently prefer one source

When two sources of the same authority conflict, the system:

1. Notes both
2. Selects the most recent (by `publishedAt`)
3. Marks the response with a low-confidence warning

## Filter by Authority (Mandatory for Specific Use Cases)

- UPSC-prep queries: only `NCERT`, `GOVT_OFFICIAL`, `ACT`, `CONSTITUTION`, `JUDICIAL`, `BOOK_ACADEMIC`
- Current affairs: only `CURRENT_AFFAIRS_OFFICIAL` (with `CURRENT_AFFAIRS_REPUTABLE` as fallback)
- Medical/legal advice: only `GOVT_OFFICIAL`, `ACT`, `JUDICIAL`; never `USER_UPLOAD` without a disclaimer

---

# Re-Ranking and Confidence Scoring (Mandatory)

## Re-Ranking

After the hybrid search returns top-20 candidates, a cross-encoder re-ranker re-scores them. The re-ranker is a separate model and is invoked via the `Reranker` interface:

```typescript
export interface Reranker {
  rerank(query: string, documents: string[], topK: number): Promise<RerankResult[]>;
}

export interface RerankResult {
  index: number;
  score: number;   // 0–1
  document: string;
}
```

Default re-ranker: `CohereRerank-v3` (or `BGE-Reranker-v2-m3` for self-hosted).

## Confidence Scoring

Each chunk in the final top-K MUST have a `confidence` field (0–1) computed from:

- `0.6 * rerank_score`
- `0.3 * hybrid_score`
- `0.1 * authority_score`

A response is "low confidence" if the average confidence of its cited chunks is < 0.5. Low-confidence responses:

- Include an explicit disclaimer: "Based on the available sources, the answer is not fully certain. The sources do not directly address your question."
- Are flagged in the audit log
- Are eligible for human review (admin queue)

---

# Telemetry, Logging, and Evaluation (Mandatory)

## Per-Query Telemetry

Every RAG query MUST produce a telemetry record:

```typescript
interface RAGTelemetry {
  queryId: string;
  userId: string;
  timestamp: string;
  question: string;
  embeddingModel: string;
  embeddingModelVersion: string;
  rerankerModel: string;
  llmModel: string;
  promptVersion: string;
  retrievalRecall?: number;
  retrievedChunkIds: string[];
  citedChunkIds: string[];
  responseLength: number;
  responseLatencyMs: number;
  faithfulnessScore: number;
  faithfulnessVerdict: FaithfulnessVerdict;
  userFeedback?: 'POSITIVE' | 'NEGATIVE' | 'NEUTRAL';
  hallucinationFlag: boolean;
}
```

## Sentry Integration

Every RAG operation is wrapped in `Sentry.startSpan` with op tags: `rag.query`, `rag.embedding`, `rag.hybrid_search`, `rag.rerank`, `rag.prompt_build`, `rag.llm_call`, `rag.verify`. Each span has tags: `rag.use_case`, `rag.model`, `rag.prompt_version`.

## Evaluation Harness

An evaluation set of ≥ 200 hand-labeled questions is maintained in `docs/evaluation/rag-eval-set.json`. Each question has `question`, `expected_answer`, `expected_sources`, `expected_authority`, `difficulty` (EASY/MEDIUM/HARD), and `category` (NCERT/GOVT/ACT/CURRENT_AFFAIRS/USER_UPLOAD).

The harness runs weekly and reports:

- **Faithfulness** (target: ≥ 0.95)
- **Answer Relevancy** (target: ≥ 0.90)
- **Retrieval Recall** (target: ≥ 0.85)
- **Citation Accuracy** (target: ≥ 0.95)
- **Hallucination Rate** (target: ≤ 0.02)

A regression in any metric by ≥ 5% blocks the release.

---

# Failure Modes and Fallbacks (Mandatory)

| Failure | Fallback |
|---|---|
| Vector DB unavailable | Return cached results from Redis. If no cache, return 503 with retry-after. |
| Embedding API unavailable | Queue the query for retry. Do not return a partial answer. |
| LLM API unavailable | Return a "I am temporarily unable to answer" message. Do not hallucinate. |
| LLM API rate-limited | Queue the query. If queue depth > 100, return 503. |
| LLM returns no citations | Reject the response, re-prompt with stronger citation instruction. If still no citations after 2 attempts, return a fallback message. |
| LLM cites non-existent source | Reject the response. Log to Sentry with high severity. |
| Re-ranker unavailable | Skip re-ranking; use the hybrid search top-5 directly. |
| All sources filtered out | Return "No relevant sources found for this query. Try rephrasing." |
| Chunking strategy change detected mid-query | Log warning, complete the query, do not switch mid-flight. |

## Circuit Breaker (Mandatory)

A circuit breaker (Hystrix-style) wraps the LLM call. After 5 consecutive failures within 60 seconds, the circuit opens and the system returns the fallback for 5 minutes before retrying.

---

# Compatibility With Other Standards

This document defers to:

- [ai-rules.md](docs/standards/ai-rules.md) for general AI model usage and provider rules
- [security-rules.md](docs/standards/security-rules.md) for prompt-injection defense and PII handling
- [architecture-rules.md](docs/standards/architecture-rules.md) for the ADR requirement (an embedding model change requires an ADR)
- [backend-rules.md](docs/standards/backend-rules.md) for service patterns, Sentry instrumentation, and module structure
- [Course-rules.md](docs/standards/Course-rules.md) for Course ownership and audit requirements
- [admin-panel-rules.md](docs/standards/admin-panel-rules.md) for moderation of user-uploaded knowledge bases

---

# Sprint 1 Compliance Checklist

Since the RAG module is not yet implemented, this section will be expanded in v1.0. For Sprint 1:

- [ ] ADR written for: vector DB choice (Qdrant), embedding model (initial), chunking strategy
- [ ] `VectorDBClient` interface defined and `QdrantClient` adapter implemented
- [ ] `Chunker` configuration stored in DB
- [ ] `EmbeddingModel` table created with at least one `ACTIVE` model
- [ ] Health check endpoint includes vector DB round-trip test
- [ ] Faithfulness verifier implemented (even if a stub)
- [ ] At least one prompt template versioned and committed
- [ ] Sentry spans wrap every step of the pipeline
- [ ] Telemetry table or stream exists for per-query logging
- [ ] Evaluation harness runs locally and produces a baseline

---

# Future RAG Architecture (Out of Scope for v0.5)

- Knowledge graph layer (entity extraction, relation mapping)
- Personal Knowledge Base (per-user embeddings)
- Learning Memory (long-term user-specific context)
- Multi-hop reasoning
- Multi-modal RAG (image, video, audio)
- Agentic RAG
- Streaming responses with incremental citation
- Fine-tuned re-ranker for GK Circle's specific corpus

---

# Final Directive

RAG is the credibility backbone of GK Circle.

A hallucinated answer is a student who learned something wrong.

A wrong student is a wasted exam attempt.

A wasted exam attempt is a lost user.

Build RAG that is grounded, cited, verifiable, and versioned.

Verify all four.
