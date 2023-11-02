---
subcategory: "Detective"
layout: "aws"
page_title: "AWS: aws_detective_organization_admin_account"
description: |-
  Manages a Detective Organization Admin Account
---

# Resource: aws_detective_organization_admin_account

Manages a Detective Organization Admin Account. The AWS account utilizing this resource must be an Organizations primary account. More information about Organizations support in Detective can be found in the [Detective User Guide](https://docs.aws.amazon.com/detective/latest/adminguide/accounts-orgs-transition.html).

## Example Usage

```terraform
resource "aws_organizations_organization" "example" {
  aws_service_access_principals = ["detective.amazonaws.com"]
  feature_set                   = "ALL"
}

resource "aws_detective_organization_admin_account" "example" {
  depends_on = [aws_organizations_organization.example]

  account_id = "123456789012"
}
```

## Argument Reference

The following arguments are supported:

* `account_id` - (Required) AWS account identifier to designate as a delegated administrator for Detective.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS account identifier.

## Import

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_detective_organization_admin_account` using `account_id`. For example:

```terraform
import {
  to = aws_detective_organization_admin_account.example
  id = 123456789012
}
```

Using `terraform import`, import `aws_detective_organization_admin_account` using `account_id`. For example:

```console
% terraform import aws_detective_organization_admin_account.example 123456789012
```
