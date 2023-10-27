---
subcategory: "Macie"
layout: "aws"
page_title: "AWS: aws_macie2_organization_admin_account"
description: |-
  Provides a resource to manage an Amazon Macie Organization Admin Account.
---

# Resource: aws_macie2_organization_admin_account

Provides a resource to manage an [Amazon Macie Organization Admin Account](https://docs.aws.amazon.com/macie/latest/APIReference/admin.html).

## Example Usage

```terraform
resource "aws_macie2_account" "example" {}

resource "aws_macie2_organization_admin_account" "example" {
  admin_account_id = "ID OF THE ADMIN ACCOUNT"
  depends_on       = [aws_macie2_account.example]
}
```

## Argument Reference

This resource supports the following arguments:

* `admin_account_id` - (Required) The AWS account ID for the account to designate as the delegated Amazon Macie administrator account for the organization.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The unique identifier (ID) of the macie organization admin account.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_macie2_organization_admin_account` using the id. For example:

```terraform
import {
  to = aws_macie2_organization_admin_account.example
  id = "abcd1"
}
```

Using `terraform import`, import `aws_macie2_organization_admin_account` using the id. For example:

```console
% terraform import aws_macie2_organization_admin_account.example abcd1
```
