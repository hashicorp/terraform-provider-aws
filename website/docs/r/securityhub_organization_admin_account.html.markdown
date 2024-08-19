---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_organization_admin_account"
description: |-
  Manages a Security Hub administrator account for an organization.
---

# Resource: aws_securityhub_organization_admin_account

Manages a Security Hub administrator account for an organization. The AWS account utilizing this resource must be an Organizations primary account. More information about Organizations support in Security Hub can be found in the [Security Hub User Guide](https://docs.aws.amazon.com/securityhub/latest/userguide/designate-orgs-admin-account.html).

## Example Usage

```terraform
resource "aws_organizations_organization" "example" {
  aws_service_access_principals = ["securityhub.amazonaws.com"]
  feature_set                   = "ALL"
}

resource "aws_securityhub_account" "example" {}

resource "aws_securityhub_organization_admin_account" "example" {
  depends_on = [aws_organizations_organization.example]

  admin_account_id = "123456789012"
}

# Auto enable security hub in organization member accounts
resource "aws_securityhub_organization_configuration" "example" {
  auto_enable = true
}
```

## Argument Reference

This resource supports the following arguments:

* `admin_account_id` - (Required) The AWS account identifier of the account to designate as the Security Hub administrator account.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS account identifier.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Security Hub Organization Admin Accounts using the AWS account ID. For example:

```terraform
import {
  to = aws_securityhub_organization_admin_account.example
  id = "123456789012"
}
```

Using `terraform import`, import Security Hub Organization Admin Accounts using the AWS account ID. For example:

```console
% terraform import aws_securityhub_organization_admin_account.example 123456789012
```
