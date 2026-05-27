---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_patch_group"
description: |-
  Provides an SSM Patch Group resource
---

# Resource: aws_ssm_patch_group

Provides an SSM Patch Group resource

## Example Usage

```terraform
resource "aws_ssm_patch_baseline" "example" {
  name             = "patch-baseline"
  approved_patches = ["KB123456"]
}

resource "aws_ssm_patch_group" "example" {
  baseline_id = aws_ssm_patch_baseline.example.id
  patch_group = "patch-group-name"
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `baseline_id` - (Required) The ID of the patch baseline to register the patch group with.
* `patch_group` - (Required) The name of the patch group that should be registered with the patch baseline.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the patch group and ID of the patch baseline separated by a comma (`,`).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_ssm_patch_group.example
  identity = {
    baseline_id = "pb-1234567890abcdef0"
    patch_group = "patch-group-name"
  }
}

resource "aws_ssm_patch_group" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `baseline_id` (String) The ID of the patch baseline.
* `patch_group` (String) The name of the patch group.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an SSM Patch Group using the `patch_group` and `baseline_id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_ssm_patch_group.example
  id = "patch-group-name,pb-1234567890abcdef0"
}
```

Using `terraform import`, import an SSM Patch Group using the `patch_group` and `baseline_id` separated by a comma (`,`). For example:

```console
% terraform import aws_ssm_patch_group.example patch-group-name,pb-1234567890abcdef0
```
