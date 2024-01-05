---
subcategory: "Verified Permissions"
layout: "aws"
page_title: "AWS: aws_verifiedpermissions_policy_store"
description: |-
  This is a Terraform resource for managing an AWS Verified Permissions Policy Store.
---

# Resource: aws_verifiedpermissions_policy_store

This is a Terraform resource for managing an AWS Verified Permissions Policy Store.

## Example Usage

### Basic Usage

```terraform
resource "aws_verifiedpermissions_policy_store" "example" {
  validation_settings {
    mode = "STRICT"
  }

  schema {
    cedar_json = jsonencode({
      "Namespace" : {
        "entityTypes" : {},
        "actions" : {}
      }
    })
  }
}
```

## Argument Reference

The following arguments are required:

* `validation_settings` - (Required) Validation settings for the policy store.
    * `mode` - (Required) The mode for the validation settings. Valid values: `OFF`, `STRICT`.
* `schema` - (Required) Schema for the policy store.
    * `cedar_json` - (Required) The cedar json schema.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `policy_store_id` - The ID of the Policy Store.
* `arn` - The ARN of the Policy Store.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Verified Permissions Policy Store using the `policy_store_id`. For example:

```terraform
import {
  to = aws_verifiedpermissions_policy_store.example
  id = "DxQg2j8xvXJQ1tQCYNWj9T"
}
```

Using `terraform import`, import Verified Permissions Policy Store using the `policy_store_id`. For example:

```
$ terraform import aws_verifiedpermissions_policy_store.example DxQg2j8xvXJQ1tQCYNWj9T
```
