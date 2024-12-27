---
subcategory: "EBS (EC2)"
layout: "aws"
page_title: "AWS: aws_ebs_snapshot_block_public_access"
description: |-
  Manages EBS snapshot public access block configuration.
---

# Resource: aws_ebs_snapshot_block_public_access

Provides a resource to manage the state of the "Block public access for snapshots" setting on region level.

~> **NOTE:** Removing this Terraform resource disables blocking.

## Example Usage

```terraform
resource "aws_ebs_snapshot_block_public_access" "example" {
  state = "block-all-sharing"
}
```

## Argument Reference

This resource supports the following arguments:

* `state` - (Required) The mode in which to enable "Block public access for snapshots" for the region. Allowed values are `block-all-sharing`, `block-new-sharing`, `unblocked`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the current state. For example:

```terraform
import {
  to = aws_ebs_snapshot_block_public_access.example
  id = "default"
}
```

Using `terraform import`, import the state. For example:

```console
% terraform import aws_ebs_snapshot_block_public_access.example default
```
