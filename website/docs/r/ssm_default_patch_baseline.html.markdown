---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_default_patch_baseline"
description: |-
  Terraform resource for managing an AWS Systems Manager Default Patch Baseline.
---

# Resource: aws_ssm_default_patch_baseline

Terraform resource for registering an AWS Systems Manager Default Patch Baseline.

## Example Usage

### Basic Usage

```terraform
resource "aws_ssm_default_patch_baseline" "example" {
  baseline_id      = aws_ssm_patch_baseline.example.id
  operating_system = aws_ssm_patch_baseline.example.operating_system
}

resource "aws_ssm_patch_baseline" "example" {
  name             = "example"
  approved_patches = ["KB123456"]
}
```

## Argument Reference

The following arguments are required:

* `baseline_id` - (Required) ID of the patch baseline.
  Can be an ID or an ARN.
  When specifying an AWS-provided patch baseline, must be the ARN.
* `operating_system` - (Required) The operating system the patch baseline applies to.
  Valid values are
  `AMAZON_LINUX`,
  `AMAZON_LINUX_2`,
  `AMAZON_LINUX_2022`,
  `CENTOS`,
  `DEBIAN`,
  `MACOS`,
  `ORACLE_LINUX`,
  `RASPBIAN`,
  `REDHAT_ENTERPRISE_LINUX`,
  `ROCKY_LINUX`,
  `SUSE`,
  `UBUNTU`, and
  `WINDOWS`.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import the Systems Manager Default Patch Baseline using the patch baseline ID, patch baseline ARN, or the operating system value. For example:

Using the patch baseline ID:

```terraform
import {
  to = aws_ssm_default_patch_baseline.example
  id = "pb-1234567890abcdef1"
}
```

Using the patch baseline ARN:

```terraform
import {
  to = aws_ssm_default_patch_baseline.example
  id = "arn:aws:ssm:us-west-2:123456789012:patchbaseline/pb-1234567890abcdef1"
}
```

Using the operating system value:

```terraform
import {
  to = aws_ssm_default_patch_baseline.example
  id = "CENTOS"
}
```

**Using `terraform import` to import** the Systems Manager Default Patch Baseline using the patch baseline ID, patch baseline ARN, or the operating system value. For example:

Using the patch baseline ID:

```console
% terraform import aws_ssm_default_patch_baseline.example pb-1234567890abcdef1
```

Using the patch baseline ARN:

```console
% terraform import aws_ssm_default_patch_baseline.example arn:aws:ssm:us-west-2:123456789012:patchbaseline/pb-1234567890abcdef1
```

Using the operating system value:

```console
% terraform import aws_ssm_default_patch_baseline.example CENTOS
```
