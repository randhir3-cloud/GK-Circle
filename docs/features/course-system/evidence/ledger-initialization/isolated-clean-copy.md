# Isolated Clean-Copy Verification

An exported `HEAD` workspace was created from commit
`eeac599f05eaf936c7f61db4a3deeac3c9063f59`. It was not copied from the
dirty working directory.

Only these paths were overlaid:

- `package.json`;
- `scripts/course-system-status.js`;
- `docs/development/**`;
- `docs/features/course-system/evidence/ledger-initialization/**`.

The root ledger package has no third-party dependencies, so no installation was
required. The checker uses Node built-ins only.

Results:

- 33 controlled files were hashed.
- Consecutive no-op sync digests were identical:
  `2af046a647934151d6c2be70760e4775a7dd93355ac05125e9801b5d5c89f9fe`.
- Before/after check digests were identical.
- Check exited 0.
- A human-authored sentinel outside the generated markers survived sync with
  identical outside-region hashes.
- All negative fixtures failed for their expected invariant.

The temporary export, archive, negative fixtures, and isolated Playwright output
were removed only after this evidence was preserved.
