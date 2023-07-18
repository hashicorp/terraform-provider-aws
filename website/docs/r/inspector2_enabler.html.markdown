---
subcategory: "Inspector"
layout: "aws"
page_title: "AWS: aws_inspector2_enabler"
description: |-
  Terraform resource for enabling Amazon Inspector resource scans.
---

# Resource: aws_inspector2_enabler

Terraform resource for enabling Amazon Inspector resource scans.

This resource must be created in the Organization's Administrator Account.

## Example Usage

### Basic Usage

```terraform
resource "aws_inspector2_enabler" "example" {
  account_ids    = ["123456789012"]
  resource_types = ["EC2"]
}
```

### For the Calling Account

```terraform
data "aws_caller_identity" "current" {}

resource "aws_inspector2_enabler" "test" {
  account_ids    = [data.aws_caller_identity.current.account_id]
  resource_types = ["ECR", "EC2"]
}
```

## Argument Reference

The following arguments are required:

* `account_ids` - (Required) Set of account IDs.
  Can contain one of: the Organization's Administrator Account, or one or more Member Accounts.
* `resource_types` - (Required) Type of resources to scan.
  Valid values are `EC2`, `ECR`, and `LAMBDA`.
  At least one item is required.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)
