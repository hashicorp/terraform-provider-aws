<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# IBM Bob Project Rules

This repository implements the [Terraform AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs): a Go bridge that maps AWS APIs to Terraform resources, data sources, ephemeral resources, and actions using the AWS SDK for Go v2. It is a [muxed](https://developer.hashicorp.com/terraform/plugin/mux) provider using both the Terraform Plugin Framework (new work) and Plugin SDKv2 (legacy).

## Where the guidance lives

Agent guidance is tool-agnostic and shared. Do not duplicate it here.

- [`AGENTS.md`](../../AGENTS.md) — repository overview, stack, conventions, build/test commands, and AI-usage policy. Bob loads this automatically.
- [`.agents/`](../../.agents) — personas (`contributor`, `maintainer`, `tcm`) and skills (`.agents/skills/`) for scoped tasks such as PR review, changelog entries, and documentation fixes.

Default to the `@contributor` persona unless another is requested, and invoke the relevant skill for the task at hand.
