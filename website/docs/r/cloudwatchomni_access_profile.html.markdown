---
subcategory: "CloudWatch Omni"
layout: "aws"
page_title: "AWS: aws_cloudwatchomni_access_profile"
description: |-
  Manages an Amazon CloudWatch Omni access profile.
---

# Resource: aws_cloudwatchomni_access_profile

Manages an Amazon CloudWatch Omni access profile.

An access profile belongs to a space and is referenced by access grants to authorize principals in that space.

~> **NOTE:** An access profile cannot be deleted while access grants reference it.

## Example Usage

### Basic Usage

```terraform
resource "aws_cloudwatchomni_access_profile" "example" {
  name     = "example"
  space_id = aws_cloudwatchomni_space.example.space_id
}
```

### With Description

```terraform
resource "aws_cloudwatchomni_access_profile" "example" {
  name        = "example"
  description = "Example access profile"
  space_id    = aws_cloudwatchomni_space.example.space_id
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the access profile.
* `space_id` - (Required, Forces new resource) ID of the space that contains the access profile.

The following arguments are optional:

* `description` - (Optional) Description of the access profile.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the access profile.
* `assume_status` - Whether the calling principal can assume the access profile. Valid values: `ALLOWED`, `DENIED`.
* `profile_id` - Unique identifier of the access profile.
* `profile_type` - Who manages the access profile. Valid values: `CUSTOMER_MANAGED`, `SERVICE_MANAGED`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)

## Import

In Terraform v1.12.0 and later, you can use an [`import` block](https://developer.hashicorp.com/terraform/language/import) with the `identity` attribute. For example:

```terraform
import {
  to = aws_cloudwatchomni_access_profile.example
  identity = {
    space_id   = "a1b2c3d4-5678-90ab-cdef-EXAMPLE11111"
    profile_id = "a1b2c3d4-5678-90ab-cdef-EXAMPLE22222"
  }
}

resource "aws_cloudwatchomni_access_profile" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `profile_id` (String) Access profile ID.
* `space_id` (String) Space ID.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import CloudWatch Omni access profiles using `space_id` and `profile_id` separated by a comma (`,`). For example:

```terraform
import {
  to = aws_cloudwatchomni_access_profile.example
  id = "a1b2c3d4-5678-90ab-cdef-EXAMPLE11111,a1b2c3d4-5678-90ab-cdef-EXAMPLE22222"
}
```

Using `terraform import`, import CloudWatch Omni access profiles using `space_id` and `profile_id` separated by a comma (`,`). For example:

```console
% terraform import aws_cloudwatchomni_access_profile.example a1b2c3d4-5678-90ab-cdef-EXAMPLE11111,a1b2c3d4-5678-90ab-cdef-EXAMPLE22222
```
