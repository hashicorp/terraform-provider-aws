---
subcategory: "Security Hub"
layout: "aws"
page_title: "AWS: aws_securityhub_organization_configuration"
description: |-
  Manages the Security Hub Organization Configuration
---

# Resource: aws_securityhub_organization_configuration

Manages the Security Hub Organization Configuration.

~> **NOTE:** This resource requires an [`aws_securityhub_organization_admin_account`](/docs/providers/aws/r/securityhub_organization_admin_account.html) to be configured (not necessarily with Terraform). More information about managing Security Hub in an organization can be found in the [Managing administrator and member accounts](https://docs.aws.amazon.com/securityhub/latest/userguide/securityhub-accounts.html) documentation.

~> **NOTE:** In order to set the `configuration_type` to `CENTRAL`, the delegated admin must be a member account of the organization and not the management account. Central configuration also requires an [`aws_securityhub_finding_aggregator`](/docs/providers/aws/r/securityhub_finding_aggregator.html) to be configured.

~> **NOTE:** This is an advanced Terraform resource. Terraform will automatically assume management of the Security Hub Organization Configuration without import and perform no actions on removal from the Terraform configuration.

~> **NOTE:** Deleting this resource resets security hub to a local organization configuration with auto enable false.

## Example Usage

### Local Configuration

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

### Central Configuration

```terraform
resource "aws_securityhub_organization_admin_account" "example" {
  depends_on = [aws_organizations_organization.example]

  admin_account_id = "123456789012"
}

resource "aws_securityhub_finding_aggregator" "example" {
  linking_mode = "ALL_REGIONS"

  depends_on = [aws_securityhub_organization_admin_account.example]
}

resource "aws_securityhub_organization_configuration" "example" {
  auto_enable           = false
  auto_enable_standards = "NONE"
  organization_configuration {
    configuration_type = "CENTRAL"
  }

  depends_on = [aws_securityhub_finding_aggregator.example]
}
```

## Argument Reference

This resource supports the following arguments:

* `auto_enable` - (Required) Whether to automatically enable Security Hub for new accounts in the organization.
* `auto_enable_standards` - (Optional) Whether to automatically enable Security Hub default standards for new member accounts in the organization. By default, this parameter is equal to `DEFAULT`, and new member accounts are automatically enabled with default Security Hub standards. To opt out of enabling default standards for new member accounts, set this parameter equal to `NONE`.
* `organization_configuration` - (Optional) Provides information about the way an organization is configured in Security Hub.

`organization_configuration` supports the following:

* `configuration_type` - (Required) Indicates whether the organization uses local or central configuration. If using central configuration, `auto_enable` must be set to `false` and `auto_enable_standards` set to `NONE`. More information can be found in the [documentation for central configuration](https://docs.aws.amazon.com/securityhub/latest/userguide/central-configuration-intro.html). Valid values: `LOCAL`, `CENTRAL`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - AWS Account ID.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `180s`)
* `update` - (Default `180s`)
* `delete` - (Default `180s`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an existing Security Hub enabled account using the AWS account ID. For example:

```terraform
import {
  to = aws_securityhub_organization_configuration.example
  id = "123456789012"
}
```

Using `terraform import`, import an existing Security Hub enabled account using the AWS account ID. For example:

```console
% terraform import aws_securityhub_organization_configuration.example 123456789012
```
