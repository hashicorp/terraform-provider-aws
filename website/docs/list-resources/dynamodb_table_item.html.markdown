---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_table_item"
description: |-
  Lists DynamoDB Table Item resources.
---

# List Resource: aws_dynamodb_table_item

Lists DynamoDB Table Item resources.

## Example Usage

```terraform
list "aws_dynamodb_table_item" "example" {
  provider = aws

  config {
    table_name = "example"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
* `table_name` - (Required) Name of the DynamoDB table to list items from.
