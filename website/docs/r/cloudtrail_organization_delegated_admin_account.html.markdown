---
subcategory: "CloudTrail"
layout: "aws"
page_title: "AWS: aws_cloudtrail_organization_delegated_admin_account"
description: |-
  Provides a resource to manage an AWS CloudTrail Delegated Administrator.
---

# Resource: aws_cloudtrail_organization_delegated_admin_account

Provides a resource to manage an AWS CloudTrail Delegated Administrator.

## Example Usage

Basic usage:

```terraform
resource "aws_cloudtrail_organization_delegated_admin_account" "example" {
  account_id = data.aws_caller_identity.delegated.account_id
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

* `account_id` - (Required) An organization member account ID that you want to designate as a delegated administrator.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - The Amazon Resource Name (ARN) of the delegated administrator's account.
* `email` - The email address that is associated with the delegated administrator's AWS account.
* `name` - The friendly name of the delegated administrator's account.
* `service_principal` - The AWS CloudTrail service principal name.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import delegated administrators using the delegate account `id`. For example:

```terraform
import {
  to = aws_cloudtrail_organization_delegated_admin_account.example
  id = "12345678901"
}
```

Using `terraform import`, import delegated administrators using the delegate account `id`. For example:

```console
% terraform import aws_cloudtrail_organization_delegated_admin_account.example 12345678901
```
