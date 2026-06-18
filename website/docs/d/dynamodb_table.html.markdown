---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_table"
description: |-
  Provides a DynamoDB table data source.
---

# Data Source: aws_dynamodb_table

Provides information about a DynamoDB table.

## Example Usage

```terraform
data "aws_dynamodb_table" "tableName" {
  name = "tableName"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name of the DynamoDB table.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

See the [DynamoDB Table Resource](/docs/providers/aws/r/dynamodb_table.html) for details on the
returned attributes - they are identical.
