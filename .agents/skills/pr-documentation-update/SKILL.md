---
name: pr-documentation-update
description: "Review a PR's end user documentation updates."
---

<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Skill: Review PR Documentation Update

Determine whether a GitHub Pull Request should include documentation updates. If so, review the documentation updates for accuracy, completeness, and adherence to style guidelines. Provide feedback on necessary changes.
If the PR should include documentation updates but does not, provide feedback on what documentation should be added.
If the PR should not include documentation updates but does, provide feedback on what documentation should be removed.

Review only the documentation changes included in the PR.
Do not review other code changes, except to the extent necessary to understand the documentation changes.
Do not review any unchanged documentation.

Authoritative reference: [docs/end-user-documentation.md](../../../docs/end-user-documentation.md). When this skill and that document disagree, the document wins.

## When to use

Trigger this skill when the user:
- Provides a `https://github.com/hashicorp/terraform-provider-aws/pull/<N>` URL and asks for a documentation review.
- Says "review documentation", or similar, with a PR URL.

## Inputs

Required:
- A GitHub PR URL. Extract `<PR_NUMBER>` with the regex `/pull/(\d+)`.

If the user provides only a PR number, ask for the full URL (or confirm the repo is `hashicorp/terraform-provider-aws`).
