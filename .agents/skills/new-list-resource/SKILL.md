---
name: new-list-resource
description: "Implement a new list resource in the Terraform AWS provider."
---

<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Skill: New List Resource

Assume the @maintainer persona.

Use this skill to add new list resource.

Refer to [docs/new-list-resource.md](../../../docs/add-a-new-list-resource.md) for instructions.
The task is complete when all list acceptance tests (`TestAcc{Resource}_List`) are passing.

## When to use

Trigger this skill when the user:

- Says "add a new list resource", or similar.

## Inputs

Required:

- The target resource name (e.g. `aws_s3_bucket`).

If the user provides a human readable name (e.g. "S3 Bucket") rather than the Terraform resource, confirm the target resource before proceeding.
