---
subcategory: "Organizations"
layout: "aws"
page_title: "AWS: aws_organizations_accounts"
description: |-
  Terraform data source for requesting the AWS Organizations Accounts in a given Organization Unit (OU).
---

# Data Source: aws_organizations_accounts

  Terraform data source for requesting the AWS Organizations Accounts in a given Organization Unit (OU).

## Example Usage

### Basic Usage

```terraform
data "aws_organizations_accounts" "example" {
  parent_id="ou-8dgp-84bcaox2"
}
```

## Argument Reference

The following arguments are required:

* `parent_id` - (Required) Organization Unit Id.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `accounts` - List of AWS Accounts depending on the Organizational Unit (OU) provided in argument `parent_id`.