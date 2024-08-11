---
subcategory: "CloudTrail"
layout: "aws"
page_title: "AWS: aws_cloudtrail_organization_admin_account"
description: |-
  Provides a resource to manage an AWS CloudTrail Delegated Administrator.
---

# Resource: aws_cloudtrail_organization_admin_account

Provides a resource to manage an AWS CloudTrail Delegated Administrator.

## Example Usage

Basic usage:

```terraform
resource "aws_cloudtrail_organization_admin_account" "example" {
  delegated_admin_account_id = data.aws_caller_identity.delegated.account_id
}

data "aws_caller_identity" "delegated" {
  provider = aws.cloudtrail_delegate_account
}

provider "aws" {
  alias = "cloudtrail_delegate_account"

  # authentication arguments omitted
}
```

## Argument Reference

This resource supports the following arguments:

* `delegated_admin_account_id` - (Required)

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Organizations ARN for the delegate account.
* `id` - The Organizations member account ID that you want to enable as the CloudTrail account.
* `email` - The Organizations email for the delegate account.
* `name` - The Organizations name for the delegate account.
* `service_principal` - The AWS service principal.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import IPAMs using the delegate account `id`. For example:

```terraform
import {
  to = aws_cloudtrail_organization_admin_account.example
  id = "12345678901"
}
```

Using `terraform import`, import IPAMs using the delegate account `id`. For example:

```console
% terraform import aws_cloudtrail_organization_admin_account.example 12345678901
```
