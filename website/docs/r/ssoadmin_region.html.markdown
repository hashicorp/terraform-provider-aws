---
subcategory: "SSO Admin"
layout: "aws"
page_title: "AWS: aws_ssoadmin_region"
description: |-
  Manages an IAM Identity Center Region.
---

# Resource: aws_ssoadmin_region

Manages an additional Region for an IAM Identity Center instance.

~> Removing this resource removes the additional Region from the IAM Identity Center instance. The primary Region cannot be removed.

## Example Usage

```terraform
data "aws_ssoadmin_instances" "example" {}

resource "aws_ssoadmin_region" "example" {
  instance_arn = one(data.aws_ssoadmin_instances.example.arns)
  region_name  = "us-west-2"
}
```

## Argument Reference

The following arguments are required:

* `instance_arn` - (Required, Forces new resource) ARN of the IAM Identity Center instance.
* `region_name` - (Required, Forces new resource) Name of the additional AWS Region to add.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `added_date` - Timestamp when the Region was added.
* `id` - Resource ID in the format `instance_arn,region_name`.
* `is_primary_region` - Whether the Region is the primary Region for the IAM Identity Center instance.
* `status` - Current Region status.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `120m`)
* `delete` - (Default `120m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IAM Identity Center Regions using the `instance_arn,region_name`. For example:

```terraform
import {
  to = aws_ssoadmin_region.example
  id = "arn:aws:sso:::instance/ssoins-1234567890abcdef,us-west-2"
}
```

Using `terraform import`, import IAM Identity Center Regions using the `instance_arn,region_name`. For example:

```console
% terraform import aws_ssoadmin_region.example arn:aws:sso:::instance/ssoins-1234567890abcdef,us-west-2
```
