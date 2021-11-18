---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_organization_configuration"
description: |-
  Manages the Security Hub Organization Configuration
---

# Resource: aws_securityhub_organization_configuration

Manages the Security Hub Organization Configuration.

~> **NOTE:** This resource requires an [`aws_securityhub_organization_admin_account`](/docs/providers/aws/r/securityhub_organization_admin_account.html) to be configured (not necessarily with Terraform). More information about managing Security Hub in an organization can be found in the [Managing administrator and member accounts](https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-accounts.html) documentation

~> **NOTE:** This is an advanced Terraform resource. Terraform will automatically assume management of the Security Hub Organization Configuration without import and perform no actions on removal from the Terraform configuration.

## Example Usage

```terraform
resource "aws_organizations_organization" "example" {
  aws_service_access_principals = ["securityhub.amazonaws.com"]
  feature_set                   = "ALL"
}

resource "aws_securityhub_organization_admin_account" "example" {
  depends_on = [aws_organizations_organization.example]

  admin_account_id = "123456789012"
}

resource "aws_securityhub_organization_configuration" "example" {
  auto_enable = true
}
```

## Argument Reference

The following arguments are supported:

* `auto_enable` - (Required) Whether to automatically enable Security Hub for new accounts in the organization.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - AWS Account ID.

## Import

An existing Security Hub enabled account can be imported using the AWS account ID, e.g.,

```
$ terraform import aws_securityhub_organization_configuration.example 123456789012
```
