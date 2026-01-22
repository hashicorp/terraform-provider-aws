---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_account"
description: |-
  Get information about an account in an organization.
---

# Data Source: aws_organizations_account

Get information about an account in an organization.

## Example Usage

### Basic Usage

```terraform
data "aws_organizations_account" "example" {
  account_id = "AWS ACCOUNT ID"
}
```

## Argument Reference

This data source supports the following arguments:

* `account_id` - (Required) Account ID number of a delegated administrator account in the organization.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the organization.
* `email` - Email address of the owner assigned to the new member account.
* `joined_method` - Method by which the account joined the organization.
* `joined_timestamp` - Date the account became a part of the organization.
* `name` - Friendly name for the member account.
* `parent_id` - Parent Organizational Unit ID or Root ID for the account.
* `state` - State of the account in the organization.
* `tags` - Map of tags for the resource.
