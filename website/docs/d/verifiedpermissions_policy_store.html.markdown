---
subcategory: "Verified Permissions"
layout: "aws"
page_title: "AWS: aws_verifiedpermissions_policy_store"
description: |-
  Terraform data source for managing an AWS Verified Permissions Policy Store.
---

# Data Source: aws_verifiedpermissions_policy_store

Terraform data source for managing an AWS Verified Permissions Policy Store.

## Example Usage

### Basic Usage

```terraform
data "aws_verifiedpermissions_policy_store" "example" {
  id = "example"
}
```

## Argument Reference

The following arguments are required:

* `id` - (Required) The ID of the Policy Store.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - The ARN of the Policy Store.
* `created_date` - The date the Policy Store was created.
* `last_updated_date` - The date the Policy Store was last updated.
* `validation_settings` - Validation settings for the policy store.
