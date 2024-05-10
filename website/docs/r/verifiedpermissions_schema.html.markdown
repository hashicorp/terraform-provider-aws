---
subcategory: "Verified Permissions"
layout: "aws"
page_title: "AWS: aws_verifiedpermissions_schema"
description: |-
  This is a Terraform resource for managing an AWS Verified Permissions Policy Store Schema.
---

# Resource: aws_verifiedpermissions_schema

This is a Terraform resource for managing an AWS Verified Permissions Policy Store Schema.

## Example Usage

### Basic Usage

```terraform
resource "aws_verifiedpermissions_schema" "example" {
  policy_store_id = aws_verifiedpermissions_policy_store.example.policy_store_id

  definition {
    value = jsonencode({
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

* `policy_store_id` - (Required) The ID of the Policy Store.
* `definition` - (Required) The definition of the schema.
    * `value` - (Required) A JSON string representation of the schema.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `namespaces` - (Optional) Identifies the namespaces of the entities referenced by this schema.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Verified Permissions Policy Store using the `policy_store_id`. For example:

```terraform
import {
  to = aws_verifiedpermissions_schema.example
  id = "DxQg2j8xvXJQ1tQCYNWj9T"
}
```

Using `terraform import`, import Verified Permissions Policy Store Schema using the `policy_store_id`. For example:

```console
 % terraform import aws_verifiedpermissions_schema.example DxQg2j8xvXJQ1tQCYNWj9T
```
