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

This data source supports the following arguments:

* `id` - (Required) ID of the Policy Store.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Policy Store.
* `created_date` - Date the Policy Store was created.
* `deletion_protection` - Whether the policy store can be deleted.
* `description` - Description of the Policy Store.
* `last_updated_date` - Date the Policy Store was last updated.
* `tags` - Map of key-value pairs associated with the policy store.
* `validation_settings` - Validation settings for the policy store. See [Validation Settings](#validation_settings-block) below.

### `validation_settings` Block

* `mode` - Mode for the validation settings.
