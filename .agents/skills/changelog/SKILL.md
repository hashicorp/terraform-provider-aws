---
name: changelog
description: "Add a `.changelog/<PR_NUMBER>.txt` entry from a GitHub Pull Request URL, commit, and push (with confirmation)."
---

<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Skill: Add Changelog Entry From PR URL

Generate a `.changelog/<PR_NUMBER>.txt` entry from a GitHub Pull Request URL, commit it on the current branch, and push only after explicit user confirmation.

Authoritative reference: [docs/changelog-process.md](../../../docs/changelog-process.md). When this skill and that document disagree, the document wins.

## When to use

Trigger this skill when the user:
- Provides a `https://github.com/hashicorp/terraform-provider-aws/pull/<N>` URL and asks for a changelog.
- Says "add changelog", "create changelog entry", "write a release note", or similar, with a PR URL.

Do **not** trigger for:
- Edits to `CHANGELOG.md` directly (that file is generated — never modify it by hand).
- PRs that are docs-only, test-only, code refactors, or dependency bumps with no operator-visible effect.

Do **not** generate a changelog file if the PR changes do not affect the terraform practitioners in the end.

Do **not** look at pre-existing changelog files when given the PR, not even if it is present in the diff.

## Inputs

Required:
- A GitHub PR URL. Extract `<PR_NUMBER>` with the regex `/pull/(\d+)`.

If the user provides only a PR number, ask for the full URL (or confirm the repo is `hashicorp/terraform-provider-aws`).

## Show, commit, and gate the push

1. Print the generated file contents back to the user.
2. Run:

   ```bash
   git add .changelog/<PR_NUMBER>.txt
   git commit -m "Add CHANGELOG for #<PR_NUMBER>"
   ```

3. **Stop** and ask: "Ready to push to the current branch?" Only run `git push` after the user confirms.
4. Never run `git push --force`, `--force-with-lease`, or `--no-verify`. Never switch or create branches.