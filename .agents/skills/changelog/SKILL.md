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
- PRs that are docs-only, test-only, code refactors, or dependency bumps with no operator-visible effect — see "Skip rules" below.

## Inputs

Required:
- A GitHub PR URL. Extract `<PR_NUMBER>` with the regex `/pull/(\d+)`.

If the user provides only a PR number, ask for the full URL (or confirm the repo is `hashicorp/terraform-provider-aws`).

## Procedure

### 1. Fetch PR context

Use `fetch_webpage` against:
- The PR URL — for title, description body, labels.
- `<PR_URL>/files` — for the diff: list of changed files, added files, annotation lines (`@FrameworkResource`, `@SDKResource`, `@FrameworkDataSource`, `@SDKDataSource`, `@FrameworkListResource`, `@FrameworkEphemeralResource`, `@FrameworkAction`).

If `fetch_webpage` fails or returns non-PR content, **stop** and ask the user to paste the PR title and a short description. Do not guess.

### 2. Decide the category (silent inference)

Read [docs/changelog-process.md](../../../docs/changelog-process.md)

### 3. Skip rules

Do not create a file (and tell the user why) when the diff is **only**:
- Documentation under `website/docs/**` or `docs/**`.
- Tests (`*_test.go`) with no behavior change.
- Internal refactors / renames with no operator-visible effect.
- CI / framework / flex / skaff / tooling / `.github/**` / `Makefile` only.

If unsure, prefer creating an `enhancement` entry and let the reviewer decide.

### 4. Write the file

Write `.changelog/<PR_NUMBER>.txt` containing one or more fenced blocks. Use **literal** triple-backtick fences. No surrounding prose, no trailing newlines beyond one.

Style rules (mimic existing `.changelog/*.txt`):
- Entry text starts with a capital letter.
- Backtick attribute, argument, and resource names: `` `attr_name` ``.
- Short entries do not end with a period; full sentences do.
- Do **not** include `[GH-####]` or PR links — `go-changelog` adds those automatically. The only exception is the verbatim region-validation template in [docs/changelog-process.md](../../../docs/changelog-process.md).

### 5. Show, commit, and gate the push

1. Print the generated file contents back to the user.
2. Run:

   ```bash
   git add .changelog/<PR_NUMBER>.txt
   git commit -m "Add CHANGELOG for #<PR_NUMBER>"
   ```

3. **Stop** and ask: "Ready to push to the current branch?" Only run `git push` after the user confirms.
4. Never run `git push --force`, `--force-with-lease`, or `--no-verify`. Never switch or create branches.

## Guardrails

- Never edit `CHANGELOG.md` — it is generated.
- Filename must be exactly `<PR_NUMBER>.txt` (no prefix, no suffix).
- One file per PR. If the file already exists, read it, show it, and ask the user whether to overwrite.
- Resource prefix is `resource/aws_x:` (slash, then colon-space). Not `resource: aws_x` and not `aws_x:`.
- If the current branch is `main` or `master`, **stop** before committing and ask the user to switch to a feature branch.
- If the PR webpage cannot be fetched, ask the user to paste the title and a one-line summary rather than guessing.
