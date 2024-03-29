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

## Argument Reference

This resource supports the following arguments:

* `table_name` - (Required) The name of the table to enable contributor insights
* `index_name` - (Optional) The global secondary index name

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_dynamodb_contributor_insights` using the format `name:table_name/index:index_name`, followed by the account number. For example:

```terraform
import {
  to = aws_dynamodb_contributor_insights.test
  id = "name:ExampleTableName/index:ExampleIndexName/123456789012"
}
```

Using `terraform import`, import `aws_dynamodb_contributor_insights` using the format `name:table_name/index:index_name`, followed by the account number. For example:

```console
% terraform import aws_dynamodb_contributor_insights.test name:ExampleTableName/index:ExampleIndexName/123456789012
```
