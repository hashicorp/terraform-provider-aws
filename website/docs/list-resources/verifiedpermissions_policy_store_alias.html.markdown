---
subcategory: "Verified Permissions"
layout: "aws"
page_title: "AWS: aws_verifiedpermissions_policy_store_alias"
description: |-
  Lists AWS Verified Permissions Policy Store Alias resources.
---

# List Resource: aws_verifiedpermissions_policy_store_alias

Lists AWS Verified Permissions Policy Store Alias resources.

## Example Usage

```terraform
resource "aws_verifiedpermissions_policy_store" "example" {
  validation_settings {
    mode = "OFF"
  }
}

list "aws_verifiedpermissions_policy_store_alias" "example" {
  provider = aws

  config {
    policy_store_id = aws_verifiedpermissions_policy_store.example.policy_store_id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `policy_store_id` - (Optional) ID of the policy store to list aliases for. When omitted, aliases for all policy stores in the configured Region are returned.
* `region` - (Optional) Region to query. Defaults to provider region.
