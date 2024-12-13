---
subcategory: "GuardDuty"
layout: "aws"
page_title: "AWS: aws_guardduty_organization_admin_account"
description: |-
  Manages a GuardDuty Organization Admin Account
---

# Resource: aws_guardduty_organization_admin_account

Manages a GuardDuty Organization Admin Account. The AWS account utilizing this resource must be an Organizations primary account. More information about Organizations support in GuardDuty can be found in the [GuardDuty User Guide](https://docs.aws.amazon.com/guardduty/latest/ug/guardduty_organizations.html).

## Example Usage

```terraform
resource "aws_organizations_organization" "example" {
  aws_service_access_principals = ["guardduty.amazonaws.com"]
  feature_set                   = "ALL"
}

resource "aws_guardduty_detector" "example" {}

resource "aws_guardduty_organization_admin_account" "example" {
  depends_on = [aws_organizations_organization.example]

  admin_account_id = "123456789012"
}
```

## Argument Reference

This resource supports the following arguments:

* `admin_account_id` - (Required) AWS account identifier to designate as a delegated administrator for GuardDuty.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS account identifier.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import GuardDuty Organization Admin Account using the AWS account ID. For example:

```terraform
import {
  to = aws_guardduty_organization_admin_account.example
  id = "123456789012"
}
```

Using `terraform import`, import GuardDuty Organization Admin Account using the AWS account ID. For example:

```console
% terraform import aws_guardduty_organization_admin_account.example 123456789012
```
