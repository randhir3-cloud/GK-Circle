# COURSE-P2-T14 Supported Blocks

## Rendered in T14

| Type | Required data |
|---|---|
| TEXT | `text` |
| HEADING | `text`, `level` (`2..6`) |
| IMAGE | safe `url`, `alt`; optional `caption` |
| VIDEO | safe `url`, `title`; optional `caption` |
| PDF | safe `url`, `title` |
| LINK | safe `url`, `label` |
| QUOTE | `text`; optional `attribution` |
| CALLOUT | `text`; optional `attribution` |
| DIVIDER | object data |

## Safe fallback only

- `CODE`
- `TABLE`
- unknown future block types

Backend-supported does not automatically mean T14-renderer-supported. CODE and
TABLE remain valid, persisted, and transportable backend values. T14 represents
them through the unsupported-block fallback until a later renderer task; it
does not classify them as invalid or malformed.

The backend currently supports CALLOUT but does not accept QUOTE during metadata
validation. T14 intentionally understands both presentation types without
changing backend validation or persistence.
