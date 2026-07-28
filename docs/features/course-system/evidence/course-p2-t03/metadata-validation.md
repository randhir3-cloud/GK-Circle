# COURSE-P2-T03 Metadata Validation

## Envelope

```json
{
  "version": 1,
  "blocks": [
    {
      "id": "b1",
      "type": "TEXT",
      "data": { "text": "hello" }
    }
  ]
}
```

## Rules enforced

| Rule | Outcome |
|---|---|
| Omit / empty create metadata | Defaults to `{"version":1,"blocks":[]}` |
| `version >= 1` | Required; `< 1` or missing → `ErrLearningItemMetadataVersionInvalid` |
| `blocks` array | Required; missing/null → `ErrLearningItemMetadataInvalid`; empty array allowed |
| Block `id` | Non-blank after trim; duplicates → `ErrLearningItemBlockDuplicate` |
| Block `type` | Enum: TEXT, HEADING, IMAGE, VIDEO, PDF, LINK, CALLOUT, CODE, TABLE, DIVIDER |
| Block `data` | Non-null JSON object; arrays/scalars/null/missing rejected |
| Valid non-empty `data` | Preserved on canonical remashal (not coerced to `{}`) |

## Integration

- `CreateLearningItem` and `UpdateLearningItem` call `normalizeLearningItemMetadata` before persistence.
- Invalid metadata never reaches INSERT/UPDATE.
- No HTTP, rendering, or migration changes.
