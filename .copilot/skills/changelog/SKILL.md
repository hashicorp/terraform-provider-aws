---
name: changelog
description: "Add a Changie change fragment (`.changes/<major>.x/unreleased/*.yaml`) for a GitHub Pull Request, commit, and push."
---

<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Skill: Add Changelog Entry

Generate one or more [Changie](https://changie.dev/) change fragments for a GitHub Pull Request, commit them on the current branch, and push.

Authoritative reference: [docs/changelog-process.md](../../../docs/changelog-process.md). When this skill and that document disagree, the document wins.

## When to use

Trigger this skill when the user:
- Says "add changelog", "create changelog entry", "write a release note", or similar.

Do **not** trigger for:
- Edits to `CHANGELOG.md` directly (that file is generated — never modify it by hand).
- PRs that are docs-only, test-only, code refactors, or dependency bumps with no operator-visible effect.

Do **not** generate a fragment if the PR changes do not affect Terraform practitioners.

Do **not** create `.changelog/<PR_NUMBER>.txt` files. That `go-changelog` format is retired, and CI rejects new files.

Do **not** look at pre-existing fragments when given the PR, not even if they are present in the diff.

## Inputs

- The `<PR_NUMBER>` and diff are available from the PR context for which the cloud agent was invoked.

## Create the fragment(s)

Create one fragment per change with `changie new --interactive=false`. The kinds and fields are defined in `.changie.yaml`.

| Kind | Use for | Fields |
|---|---|---|
| `feature` | New resource, data source, ephemeral resource, function, list resource, or action | `NewType`, `Name`, `PullRequest` |
| `enhancement` | New argument or attribute, or new behavior on existing functionality | `Impact`, `Body`, `PullRequest` |
| `bug` | Bug fix | `Impact`, `Body`, `PullRequest` |
| `note` | Deprecation or other operator-facing information | `Impact`, `Body`, `PullRequest` |
| `breaking-change` | Breaking change (major versions only) | `Impact`, `Body`, `PullRequest` |

- `NewType` is one of `New Resource`, `New Data Source`, `New Ephemeral Resource`, `New Function`, `New List Resource`, or `New Action`.
- `Impact` is the prefix shown before the colon, such as `resource/aws_example_thing`, `data-source/aws_example_thing`, or `provider`.
- `Body` starts with a capitalized verb and has no trailing period.
- Each `--custom` value is parsed as a comma-separated list. If a value contains a comma, wrap the whole `Key=Value` pair in double quotes inside single quotes, as in the second example.

Examples:

```bash
changie new --interactive=false --kind feature \
  --custom "NewType=New Resource" \
  --custom "Name=aws_example_thing" \
  --custom "PullRequest=<PR_NUMBER>"

changie new --interactive=false --kind enhancement \
  --custom "Impact=resource/aws_example_thing" \
  --custom '"Body=Add `not_broken` and `fixed` arguments, with defaults"' \
  --custom "PullRequest=<PR_NUMBER>"
```

Changie names fragments with a one-second timestamp (`<kind>-YYYYMMDD-HHMMSS.yaml`). When creating several fragments of the same kind, run `sleep 1` between them so one doesn't overwrite another.

## Show, commit and push

1. Print the generated file contents back to the user.
2. Run:

   ```bash
   git add .changes/
   git commit -m "Add CHANGELOG for #<PR_NUMBER>"
   git push
   ```
3. Never run `git push --force`, `--force-with-lease`, or `--no-verify`. Never switch or create branches.
