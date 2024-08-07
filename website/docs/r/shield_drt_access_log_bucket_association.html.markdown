---
subcategory: "Shield"
layout: "aws"
page_title: "AWS: aws_shield_drt_access_log_bucket_association"
description: |-
  Terraform resource for managing an AWS Shield DRT Access Log Bucket Association.
---

# Resource: aws_shield_drt_access_log_bucket_association

Terraform resource for managing an AWS Shield DRT Access Log Bucket Association.
Up to 10 log buckets can be associated for DRT Access sharing with the Shield Response Team (SRT).

## Example Usage

### Basic Usage

```terraform
resource "aws_shield_drt_access_role_arn_association" "test" {
  role_arn = "arn:aws:iam:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:${var.shield_drt_access_role_name}"
}

resource "aws_shield_drt_access_log_bucket_association" "test" {
  log_bucket              = var.shield_drt_access_log_bucket
  role_arn_association_id = aws_shield_drt_access_role_arn_association.test.id
}
```

## Argument Reference

The following arguments are required:

* `log_bucket` - (Required) The Amazon S3 bucket that contains the logs that you want to share.
* `role_arn_association_id` - (Required) The ID of the Role Arn association used for allowing Shield DRT Access.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Shield DRT access log bucket associations using the `log_bucket`. For example:

```terraform
import {
  to = aws_shield_drt_access_log_bucket_association.example
  id = "example-bucket"
}
```

Using `terraform import`, import Shield DRT access log bucket associations using the `log_bucket`. For example:

```console
% terraform import aws_shield_drt_access_log_bucket_association.example example-bucket
```
