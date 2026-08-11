---
name: breaking-changes
description: "Review a PR for possible breaking changes."
---

<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Skill: Review PR For Breaking Changes

Assume the @maintainer persona.

Determine whether a GitHub Pull Request makes a breaking change. If so provide feedback on reasons for the breaking change and possible next steps.
If the PR does not make any breaking changes, note that.

Do NOT consider any `breaking-change` label applied to the PR.

Review only the code changes included in the PR.

Authoritative reference: [docs/breaking-changes.md](../../../docs/breaking-changes.md). When this skill and that document disagree, the document wins.

## When to use

Trigger this skill when the user:
- Provides a `https://github.com/hashicorp/terraform-provider-aws/pull/<N>` URL and asks for a breaking change review.
- Says "review breaking change", or similar, with a PR URL.

## Inputs

Required:
- A GitHub PR URL. Extract `<PR_NUMBER>` with the regex `/pull/(\d+)`.

If the user provides only a PR number, ask for the full URL (or confirm the repo is `hashicorp/terraform-provider-aws`).
