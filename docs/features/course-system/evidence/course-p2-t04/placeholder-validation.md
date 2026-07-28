# COURSE-P2-T04 Placeholder Validation

## Syntax

Placeholders appear only in string leaves inside block `data`:

```text
{{identifier}}
```

Identifier rules:

- first character: ASCII letter
- remaining: ASCII letters, digits, underscore
- length: 1–64

## Scan

1. Recursively walk nested objects and arrays inside `data`.
2. For each string, find `{{` … `}}` tokens.
3. Unclosed / nested braces / empty `{{}}` → `ErrLearningItemPlaceholderSyntax`.
4. Closed token with invalid identifier → `ErrLearningItemPlaceholderInvalid`.
5. Lone `{` is ordinary text.

## Allowed

- zero placeholders
- repeated placeholders
- multiple distinct placeholders
- placeholders in nested objects and array string values

## Preserved

Original `data` bytes and placeholder text are not rewritten after validation.
