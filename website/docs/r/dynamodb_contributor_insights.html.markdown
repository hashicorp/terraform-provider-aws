---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_contributor_insights"
description: |-
  Provides a DynamoDB contributor insights resource
---

# Resource: aws_dynamodb_contributor_insights

Provides a DynamoDB contributor insights resource

## Example Usage

```terraform
resource "aws_dynamodb_contributor_insights" "test" {
  table_name = "ExampleTableName"
}
```

## ArgumentReference

The following arguments are supported:

* `table_name` - (Required) The name of the table to attach contributor insight
* `index_name` - (Optional) The global secondary index name

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

* `status` - The contributor insights status

## Import

`aws_dynamodb_contributor_insights` can be imported using the `table_name`, or `table_name-index_name`, followed by the account number, e.g.,

```
$ terraform import aws_dynamodb_contributor_insights.test ExampleTableName/123456789012
```
