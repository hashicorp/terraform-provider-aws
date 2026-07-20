---
name: changelog
description: "Add a `.changelog/<PR_NUMBER>.txt` entry for a GitHub Pull Request, commit, and push."
---

<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Skill: Add Changelog Entry

Generate a `.changelog/<PR_NUMBER>.txt` entry for a GitHub Pull Request, commit it on the current branch, and push.

Authoritative reference: [docs/changelog-process.md](../../../docs/changelog-process.md). When this skill and that document disagree, the document wins.

## When to use

Trigger this skill when the user:
- Says "add changelog", "create changelog entry", "write a release note", or similar.

Do **not** trigger for:
- Edits to `CHANGELOG.md` directly (that file is generated — never modify it by hand).
- PRs that are docs-only, test-only, code refactors, or dependency bumps with no operator-visible effect.

Do **not** generate a changelog file if the PR changes do not affect the terraform practitioners in the end.

Do **not** look at pre-existing changelog files when given the PR, not even if it is present in the diff.

## Inputs

- The `<PR_NUMBER>` and diff are available from the PR context for which the cloud agent was invoked.

## Show, commit and push

1. Print the generated file contents back to the user.
2. Run:

   ```bash
   git add .changelog/<PR_NUMBER>.txt
   git commit -m "Add CHANGELOG for #<PR_NUMBER>"
   git push
   ```
3. Never run `git push --force`, `--force-with-lease`, or `--no-verify`. Never switch or create branches.