---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_tableitem"
description: |-
  Terraform data source for managing an AWS DynamoDB TableItem.
---

# Data Source: aws_dynamodb_tableitem

Terraform data source for managing an AWS DynamoDB TableItem.

## Example Usage

### Basic Usage

```terraform
data "aws_dynamodb_tableitem" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the TableItem.
* `example_attribute` - Concise description.
