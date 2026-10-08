---
name: resource-identity
description: "Add resource identity to a resource in the Terraform AWS provider."
---

<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Skill: Add Resource Identity Support

Assume the @maintainer persona.

Use this skill to add resource identity to an existing resource, or when creating a net-new resource.

Refer to [docs/resource-identity.md](../../../docs/resource-identity.md) for instructions.
The task is complete when all generated identity acceptance tests (`TestAcc{Resource}_Identity`) are passing.

## When to use

Trigger this skill when the user:

- Says "add resource identity support", or similar while referring to a Terraform resource.

## Inputs

Required:

- The target resource name (e.g. `aws_s3_bucket`).

If the user provides a human readable name (e.g. "S3 Bucket") rather than the Terraform resource, confirm the target resource before proceeding.
